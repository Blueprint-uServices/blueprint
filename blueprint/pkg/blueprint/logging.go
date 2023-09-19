package blueprint

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"

	"golang.org/x/exp/slog"
)

type BlueprintLoggerHandlerOptions struct {
	SlogOpts slog.HandlerOptions
}

//Implementation of Blueprint's custom Handler of slog.Logger
type BlueprintLoggerHandler struct {
	slog.Handler
	l *log.Logger
}

func (h *BlueprintLoggerHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	timeStr := r.Time.Format("[15:05:05.000]")

	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	b, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}

	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()
	source_str := "[" + f.File + ":" + strconv.Itoa(f.Line) + "]"

	if len(fields) != 0 {
		h.l.Println(timeStr, source_str, level, r.Message, string(b))
	} else {
		h.l.Println(timeStr, source_str, level, r.Message)
	}

	return nil
}

func newBlueprintLoggerHandler(out io.Writer, opts BlueprintLoggerHandlerOptions) *BlueprintLoggerHandler {
	h := &BlueprintLoggerHandler{
		Handler: slog.NewTextHandler(out, &opts.SlogOpts),
		l:       log.New(out, "", 0),
	}

	return h
}

//Initializes the logger when this package is first loaded. This function is guaranteed to be invoked only once so the logger will be initialized only once.
func init() {

	opts := slog.HandlerOptions{
		AddSource: true,
	}
	blOpts := BlueprintLoggerHandlerOptions{SlogOpts: opts}

	logger := slog.New(newBlueprintLoggerHandler(os.Stdout, blOpts))
	slog.SetDefault(logger)
}
