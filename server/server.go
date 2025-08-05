package server

import (
	"context"
	"fmt"
	"github.com/kkiling/goplatform/log"
	"net"
	"net/http"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	rn "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

type HandlerFromEndpoint = func(context.Context, *rn.ServeMux, string, []grpc.DialOption) error

type HandlerService interface {
	RegistrationServerHandlers(*http.ServeMux)
	RegisterServiceHandlerFromEndpoint() HandlerFromEndpoint
	RegisterServiceServer(*grpc.Server)
}

type Server struct {
	cfg               Config
	logger            log.Logger
	grpcServer        *grpc.Server
	gatewayServer     *http.Server
	mux               *rn.ServeMux
	opts              []grpc.DialOption
	errors            chan error
	unaryInterceptors []grpc.UnaryServerInterceptor
}

func NewServer(logger log.Logger, cfg Config) *Server {
	muxOption := rn.WithMarshalerOption(rn.MIMEWildcard, &rn.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{},
	})

	return &Server{
		cfg:           cfg,
		logger:        logger.Named("server"),
		grpcServer:    nil,
		gatewayServer: nil,
		mux:           rn.NewServeMux(muxOption),
		opts: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(cfg.MaxReceiveMessageLength),
				grpc.MaxCallSendMsgSize(cfg.MaxSendMessageLength),
			),
		},
		errors:            make(chan error, 1),
		unaryInterceptors: nil,
	}
}

func (s *Server) WitUnaryServerInterceptor(interceptors ...grpc.UnaryServerInterceptor) {
	s.unaryInterceptors = interceptors
}

func (s *Server) Start(ctx context.Context, swaggerName string, impl ...HandlerService) error {
	s.grpcServer = grpc.NewServer(
		grpc.MaxRecvMsgSize(s.cfg.MaxReceiveMessageLength),
		grpc.MaxSendMsgSize(s.cfg.MaxSendMessageLength),
		// Регистрация интерсепторов
		grpc.ChainUnaryInterceptor(s.unaryInterceptors...),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	host := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.GrpcPort)

	for _, service := range impl {
		// Регистрация grpc методов
		service.RegisterServiceServer(s.grpcServer)
		// Регистрация rest api gateway
		if err := service.RegisterServiceHandlerFromEndpoint()(ctx, s.mux, host, s.opts); err != nil {
			return fmt.Errorf("failed to register HTTP server: %v", err)
		}
	}

	// После инициализации сервера:
	grpc_prometheus.Register(s.grpcServer)
	// Нужно что бы сервер сам отдавал описание методов
	// например для postman
	reflection.Register(s.grpcServer)

	// Запуск grpc сервера
	go func(logger log.Logger) {
		netAddress := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.GrpcPort)

		logger.Infof("start server at %s", netAddress)
		socket, err := net.Listen("tcp", netAddress)
		if err != nil {
			s.errors <- err
			return
		}
		s.errors <- s.grpcServer.Serve(socket)
	}(s.logger.Named("grpc_server"))

	// Запуск rest api сервера с gateway
	go func(logger log.Logger) {
		netAddress := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.HttpPort)

		httpMux := http.NewServeMux()
		for _, service := range impl {
			service.RegistrationServerHandlers(httpMux)
		}

		// OpenApi спецификация апи
		swagger := fmt.Sprintf("/%s.swagger.json", swaggerName)

		httpMux.Handle(swagger, http.FileServer(http.Dir("./swagger")))

		// Swagger в браузере
		httpMux.Handle("/swagger/", httpSwagger.Handler(
			httpSwagger.URL(swagger),
		))
		// Метрики
		httpMux.Handle("/metrics", promhttp.Handler())
		// Обрабатываем остальные запросы через gRPC-Gateway
		httpMux.Handle("/", s.mux)

		s.gatewayServer = &http.Server{
			Addr:    netAddress,
			Handler: httpMux,
		}

		logger.Infof("start gateway at %s", netAddress)
		s.errors <- s.gatewayServer.ListenAndServe()

	}(s.logger.Named("http_server"))

	return <-s.errors
}

func (s *Server) Stop() {
	// Пробуем по хорошему
	go s.grpcServer.GracefulStop()
	// Ждем
	time.Sleep(time.Duration(s.cfg.ShutdownTimeout) * time.Second)
	// Уже по плохому
	s.grpcServer.Stop()
}
