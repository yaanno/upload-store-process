package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type Logger struct {
	zerolog.Logger
}

type LoggerConfig struct {
	Level       string `yaml:"level"`
	JSON        bool   `yaml:"json_format"`
	Development bool   `yaml:"development"`
}

func New(cfg LoggerConfig) *Logger {
	// Set global error stack marshaler
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Default log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var output zerolog.Logger
	if cfg.Development {
		// Pretty print in development
		output = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Logger()
	} else {
		// JSON output for production
		output = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return &Logger{output}
}

// Enhanced logging methods
func (l *Logger) WithService(serviceName string) zerolog.Logger {
	return l.With().Str("service", serviceName).Logger()
}

func (l *Logger) WithTrace(traceID string) zerolog.Logger {
	return l.With().Str("trace_id", traceID).Logger()
}
