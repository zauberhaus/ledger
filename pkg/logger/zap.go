package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type ZapLogger struct {
	*zap.SugaredLogger
	level Level
}

func NewZapLogger(options ...interface{}) Logger {
	skip := -1
	level := InfoLevel

	for _, o := range options {
		switch v := o.(type) {
		case Level:
			level = v
		case int:
			skip = v
		}
	}

	var caller zap.Option

	if skip == -1 {
		caller = zap.WithCaller(false)
	} else {
		caller = zap.AddCallerSkip(skip)
	}

	return NewZapLoggerWithOptions(
		level,
		caller,
	)
}

func NewZapLoggerWithOptions(level Level, opts ...zap.Option) Logger {

	encoder := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.Level(level)),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "console",
		EncoderConfig:    encoder,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build(opts...)
	if err != nil {
		log.Fatal(err)
	}

	return &ZapLogger{
		logger.Sugar(),
		level,
	}
}

func NewObservableZapLogger(level Level) (Logger, *observer.ObservedLogs) {
	return NewObservableZapLoggerWithOptions(level,
		zap.AddCaller(),
	)
}

func NewObservableZapLoggerWithOptions(level Level, opts ...zap.Option) (Logger, *observer.ObservedLogs) {
	core, recorded := observer.New(zapcore.Level(level))
	logger := zap.New(core, opts...)

	return &ZapLogger{
		logger.Sugar(),
		level,
	}, recorded
}

func (z *ZapLogger) With(args ...interface{}) Logger {
	return &ZapLogger{
		z.SugaredLogger.With(args...),
		z.level,
	}
}

func (z *ZapLogger) GetLevel() Level {
	return Level(z.level)
}
