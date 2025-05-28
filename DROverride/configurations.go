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
	goas  []groupOrAttrs
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

	if r.NumAttrs() > 0 {
		buf = append(buf, " \033[38;2;176;176;176m\n  {\n"...)
	}

	goas := h.goas
	if r.NumAttrs() == 0 {
		// If the record has no Attrs, remove groups at the end of the list; they are empty.
		for len(goas) > 0 && goas[len(goas)-1].group != "" {
			goas = goas[:len(goas)-1]
		}
	}

	indentation := 2
	for _, goa := range goas {
		if goa.group != "" {
			buf = fmt.Appendf(buf, "%*s%s: {\n", indentation*2, " ", goa.group)
			indentation++
		} else {
			for _, a := range goa.attrs {
				buf = fmt.Appendf(buf, "%*s%s: %s,\n", indentation*2, " ", a.Key, a.Value)
			}
		}
	}

	// If there are any attributes attached to the message, print them out line by line via the pattern "key: value"
	if r.NumAttrs() > 0 {
		r.Attrs(func(a slog.Attr) bool {
			buf = fmt.Appendf(buf, "%*s%s: %s,\n", indentation*2, " ", a.Key, a.Value)
			return true
		})

		indentation--
		for indentation > 0 {
			buf = fmt.Appendf(buf, "%*s}\n", indentation*2, " ")
			indentation--
		}
	}
	buf = append(buf, "\033[0m\n"...)

	// Lock out the mutex so that there aren't multiple things being written to the io writer
	h.mutex.Lock()
	defer h.mutex.Unlock()
	_, err := h.out.Write(buf)
	return err
}

// Implementation taken from https://github.com/golang/example/blob/master/slog-handler-guide/indenthandler2/indent_handler.go
// TODO: Could fix up to be more specific & optimized for

type groupOrAttrs struct {
	group string      // group name if non-empty
	attrs []slog.Attr // attrs if non-empty
}

func (h *DebugHandler) withGroupOrAttrs(goa groupOrAttrs) *DebugHandler {
	h2 := *h
	h2.goas = make([]groupOrAttrs, len(h.goas)+1)
	copy(h2.goas, h.goas)
	h2.goas[len(h2.goas)-1] = goa
	return &h2
}

func (h *DebugHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	return h.withGroupOrAttrs(groupOrAttrs{attrs: attrs})
}

func (h *DebugHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return h.withGroupOrAttrs(groupOrAttrs{group: name})
}
