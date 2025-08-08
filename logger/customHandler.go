package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"sync"
)

const levelSectionWidth = 8

type customHandler struct {
	baseAttrs []slog.Attr
	writer    io.Writer
	opts      slog.HandlerOptions
	mu        *sync.Mutex
}

func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := append([]slog.Attr{}, h.baseAttrs...)
	newAttrs = append(newAttrs, attrs...)
	return &customHandler{baseAttrs: newAttrs, writer: h.writer, opts: h.opts, mu: &sync.Mutex{}}
}

func (h *customHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf strings.Builder

	levelSection := fmt.Sprintf("[%s]", r.Level.String())
	buf.WriteString(levelSection)
	pad := levelSectionWidth - len(levelSection)
	buf.WriteString(strings.Repeat(" ", pad))

	buf.WriteString(r.Time.Format("2006-01-02 15:04:05,123"))

	buf.WriteString(" - ")
	buf.WriteString(strings.TrimSuffix(r.Message, "\n"))

	if h.opts.AddSource && r.PC != 0 {
		buf.WriteString(" - ")
		buf.WriteString("source=")
		frame := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := frame.Next()
		if f.File != "" {
			buf.WriteString(fmt.Sprintf("%s:%d", f.File, f.Line))
		}
	}

	mergedAttrs := append([]slog.Attr{}, h.baseAttrs...)
	r.Attrs(func(attr slog.Attr) bool {
		mergedAttrs = append(mergedAttrs, attr)
		return true
	})

	for _, attr := range mergedAttrs {
		buf.WriteString(" - ")
		buf.WriteString(attr.String())
	}

	buf.WriteString("\n")

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.writer.Write([]byte(buf.String()))
	return err
}

func (h *customHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}
