package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

type Logger struct {
	*slog.Logger
	logFile *os.File
}

type Config struct {
	Level         slog.Level
	Format        string
	Output        *LogFile
	StdoutEnabled bool
	AddSource     bool
}

type LogFile struct {
	LogDir      string
	LogFileName string
}

func NewLogger(config Config) (*Logger, error) {
	var (
		writer  io.Writer = os.Stdout
		logFile *os.File
		err     error
	)

	if config.Output != nil {
		logPath := filepath.Join(config.Output.LogDir, config.Output.LogFileName)
		err = os.MkdirAll(filepath.Dir(logPath), 0666)
		if err != nil {
			return nil, err
		}
		logFile, err = os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return nil, err
		}

		if config.StdoutEnabled {
			writer = io.MultiWriter(logFile, os.Stdout)
		} else {
			writer = logFile
		}
	}

	handlerOpts := slog.HandlerOptions{
		AddSource: config.AddSource,
		Level:     config.Level,
	}

	var handler slog.Handler
	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(writer, &handlerOpts)
	default:
		handler = &customHandler{
			writer: writer,
			opts:   handlerOpts,
			mu:     &sync.Mutex{},
		}
	}

	return &Logger{
		Logger:  slog.New(handler),
		logFile: logFile,
	}, nil

}

func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}
