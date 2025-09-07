package grpc

func NewGRPCServer(storage HandlerStorage, db DBPinger, logger GRPCLogger) *Server {
	return NewServer(storage, db, logger)
}
