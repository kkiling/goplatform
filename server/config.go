package server

type Config struct {
	Host                    string `yaml:"host"`
	GrpcPort                int    `yaml:"grpc_port"`
	HttpPort                int    `yaml:"http_port"`
	MaxSendMessageLength    int    `yaml:"max_send_message_length"`
	MaxReceiveMessageLength int    `yaml:"max_receive_message_length"`
	ShutdownTimeout         int    `yaml:"shutdown_timeout"`
}
