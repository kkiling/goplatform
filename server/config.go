package server

type Config struct {
	Host                    string
	GrpcPort                int
	HttpPort                int
	MaxSendMessageLength    int
	MaxReceiveMessageLength int
	ShutdownTimeout         int
}
