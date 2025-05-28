package droverride

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

type DebugHandler struct {
	opts  Options
	mutex *sync.Mutex
	out   io.Writer
}

type Options struct {
	Level slog.Leveler
}

func NewDebugHandler(out io.Writer, opts *Options) *DebugHandler {
	handler := &DebugHandler{out: out, mutex: &sync.Mutex{}}
	if opts != nil {
		handler.opts = *opts
	}
	if handler.opts.Level == nil {
		handler.opts.Level = slog.LevelInfo
	}

	return handler
}

func (h *DebugHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *DebugHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := make([]byte, 0, 1024)

	// If a timestamp is present print it out
	if !r.Time.IsZero() {
		buf = fmt.Appendf(buf, "\033[38;2;117;117;117m[%d:%d:%02d.%03d]\033[0m ", r.Time.Hour(), r.Time.Minute(), r.Time.Second(), r.Time.Nanosecond()/1e6)
	}

	// Pretty print a colored version of the provided level
	switch r.Level {
	case slog.LevelDebug:
		buf = fmt.Appendf(buf, "\033[90m%s\033[0m ", r.Level)
	case slog.LevelInfo:
		buf = fmt.Appendf(buf, "\033[96m%s\033[0m ", r.Level)
	case slog.LevelWarn:
		buf = fmt.Appendf(buf, "\033[33m%s\033[0m ", r.Level)
	case slog.LevelError:
		buf = fmt.Appendf(buf, "\033[91m%s\033[0m ", r.Level)
	}

	// Add the provided message to the output buffer
	buf = fmt.Appendf(buf, "\033[97m%s\033[0m", r.Message)

	// If there are any attributes attached to the message, print them out line by line via the pattern "key: value"
	if r.NumAttrs() > 0 {
		buf = append(buf, " \033[38;2;176;176;176m\n  {\n"...)
		r.Attrs(func(a slog.Attr) bool {
			buf = fmt.Appendf(buf, "    %s: %s,\n", a.Key, a.Value)
			return true
		})
		buf = append(buf, "  }\033[0m\n"...)
	} else {
		buf = append(buf, "\033[0m\n"...)
	}

	// Lock out the mutex so that there aren't multiple things being written to the io writer
	h.mutex.Lock()
	defer h.mutex.Unlock()
	_, err := h.out.Write(buf)
	return err
}

func (h *DebugHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &DebugHandler{}
}

func (h *DebugHandler) WithGroup(name string) slog.Handler {
	return &DebugHandler{}
}
