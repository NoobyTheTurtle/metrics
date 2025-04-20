package logger

import (
	"go.uber.org/zap"
)

type ZapLogger struct {
	logger *zap.SugaredLogger
}

func NewZapLogger(level string, isDev bool) (*ZapLogger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	var cfg zap.Config

	if isDev {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	cfg.Level = lvl

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	sugar := l.Sugar()

	return &ZapLogger{
		logger: sugar,
	}, nil
}

func (l *ZapLogger) Debug(format string, args ...any) {
	l.logger.Debugf(format, args)
}

func (l *ZapLogger) Info(format string, args ...any) {
	l.logger.Infof(format, args...)
}

func (l *ZapLogger) Warn(format string, args ...any) {
	l.logger.Warnf(format, args...)
}

func (l *ZapLogger) Error(format string, args ...any) {
	l.logger.Errorf(format, args...)
}

func (l *ZapLogger) Fatal(format string, args ...any) {
	l.logger.Fatalf(format, args...)
}

func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}
