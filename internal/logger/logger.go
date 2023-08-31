package logger

import (
	"go.uber.org/zap"
)

type Logger struct {
	zap *zap.Logger
}

func NewLogger(level string) (*Logger, error) {

	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	// создаём новую конфигурацию логера
	config := zap.NewProductionConfig()

	// устанавливаем уровень
	config.Level = lvl

	// config.OutputPaths = []string{"stdout", "./logs/" + logFile}
	logger, err := config.Build(zap.AddCaller())
	if err != nil {
		return nil, err
	}
	return &Logger{zap: logger}, err
}

func (l Logger) Debug(msg string, fields ...zap.Field) {
	l.writer().Debug(msg, fields...)
}

func (l Logger) Info(msg string, fields ...zap.Field) {
	l.writer().Info(msg, fields...)
}

func (l Logger) writer() *zap.Logger {
	var noOpLogger = zap.NewNop()
	if l.zap == nil {
		return noOpLogger
	}
	return l.zap
}
