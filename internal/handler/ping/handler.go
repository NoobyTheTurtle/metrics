package ping

type Handler struct {
	db     DBPinger
	logger PingLogger
}

func NewHandler(db DBPinger, logger PingLogger) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
	}
}
