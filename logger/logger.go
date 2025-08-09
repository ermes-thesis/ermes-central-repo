package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
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
		err = os.MkdirAll(filepath.Dir(logPath), 0755)
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

func (l *Logger) Infof(format string, args ...interface{}) {
	ctx := context.Background()
	level := slog.LevelInfo
	if !l.Logger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])

	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pcs[0])
	_ = l.Handler().Handle(ctx, r)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	ctx := context.Background()
	level := slog.LevelError
	if !l.Logger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])

	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pcs[0])
	_ = l.Handler().Handle(ctx, r)
}

func (l *Logger) ErrorfErr(format string, args ...interface{}) error {
	errStr := fmt.Sprintf(format, args...)
	ctx := context.Background()
	level := slog.LevelError
	if !l.Logger.Enabled(ctx, level) {
		return errors.New(errStr)
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])

	r := slog.NewRecord(time.Now(), level, errStr, pcs[0])
	_ = l.Handler().Handle(ctx, r)

	return errors.New(errStr)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	ctx := context.Background()
	level := slog.LevelDebug
	if !l.Logger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])

	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pcs[0])
	_ = l.Handler().Handle(ctx, r)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	ctx := context.Background()
	level := slog.LevelWarn
	if !l.Logger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])

	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pcs[0])
	_ = l.Handler().Handle(ctx, r)
}

func (l *Logger) InfofWith(format string, args []any, attrs ...any) {
	ctx := context.Background()
	level := slog.LevelInfo
	if !l.Logger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])

	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pcs[0])
	r.Add(attrs...)
	_ = l.Handler().Handle(ctx, r)
}

func (l *Logger) DebugfWith(format string, args []any, attrs ...any) {
	ctx := context.Background()
	level := slog.LevelDebug
	if !l.Logger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])

	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pcs[0])
	r.Add(attrs...)
	_ = l.Handler().Handle(ctx, r)
}

func (l *Logger) WarnfWith(format string, args []any, attrs ...any) {
	ctx := context.Background()
	level := slog.LevelWarn
	if !l.Logger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])

	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, args...), pcs[0])
	r.Add(attrs...)
	_ = l.Handler().Handle(ctx, r)
}

func (l *Logger) ErrorfWith(format string, args []any, attrs ...any) {
	errStr := fmt.Sprintf(format, args...)

	ctx := context.Background()
	level := slog.LevelInfo
	if !l.Logger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])

	r := slog.NewRecord(time.Now(), level, errStr, pcs[0])
	r.Add(attrs...)
	_ = l.Handler().Handle(ctx, r)
}

func (l *Logger) ErrorfErrWith(format string, args []any, attrs ...any) error {
	errStr := fmt.Sprintf(format, args...)

	ctx := context.Background()
	level := slog.LevelInfo
	if !l.Logger.Enabled(ctx, level) {
		return errors.New(errStr)
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])

	r := slog.NewRecord(time.Now(), level, errStr, pcs[0])
	r.Add(attrs...)
	_ = l.Handler().Handle(ctx, r)

	return errors.New(errStr)
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger:  l.Logger.With(args...),
		logFile: l.logFile,
	}
}
