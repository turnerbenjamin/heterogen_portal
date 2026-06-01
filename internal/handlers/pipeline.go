package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
)

type PipelineContext[T any] struct {
	logger *slog.Logger
	state  T
}
type NoPipelineState struct{}

type (
	AppHandler[T any] func(http.ResponseWriter, *http.Request, *PipelineContext[T]) *etc.AppError
	Middleware[T any] func(AppHandler[T]) AppHandler[T]
)

type PipelineBuilder[T any] struct {
	newSeed      func(*http.Request) T
	errorHandler ErrorWriter
	logSink      io.Writer
}

type ErrorWriter interface {
	Write(http.ResponseWriter, *etc.AppError) error
}

func (d *PipelineContext[T]) AddLoggerKV(attrs ...slog.Attr) {
	anyAttrs := make([]any, len(attrs))
	for i, attr := range attrs {
		anyAttrs[i] = attr
	}
	d.logger = d.logger.With(anyAttrs...)
}

func NewPipelineBuilder(
	errorWriter ErrorWriter,
	logSink io.Writer,
) *PipelineBuilder[NoPipelineState] {
	return &PipelineBuilder[NoPipelineState]{
		newSeed:      func(r *http.Request) NoPipelineState { return NoPipelineState{} },
		errorHandler: errorWriter,
		logSink:      logSink,
	}
}

func NewPipelineWithStateBuilder[T any](
	newSeed func(*http.Request) T,
	errorHandler ErrorWriter,
	logSink io.Writer,
) *PipelineBuilder[T] {
	return &PipelineBuilder[T]{
		newSeed:      newSeed,
		errorHandler: errorHandler,
		logSink:      logSink,
	}
}

func (p *PipelineBuilder[T]) New(
	middlewares []Middleware[T],
	handler AppHandler[T],
) http.HandlerFunc {
	pipeline := handler

	i := len(middlewares) - 1
	for i >= 0 {
		pipeline = middlewares[i](pipeline)
		i--
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		pipelineData := &PipelineContext[T]{
			logger: slog.New(slog.NewJSONHandler(p.logSink, &slog.HandlerOptions{})),
			state:  p.newSeed(r),
		}

		pipelineData.AddLoggerKV(
			slog.String("request_method", r.Method),
			slog.String("request_path", r.URL.Path),
			slog.String("request_time", startTime.Format(time.RFC3339Nano)),
		)
		defer func() {
			pipelineData.logger.Info(
				"",
				slog.Int64("request_duration_ms", time.Since(startTime).Milliseconds()),
			)
		}()

		sw := &statusSpyWriter{ResponseWriter: w}

		appError := pipeline(sw, r, pipelineData)
		if appError != nil {
			err := p.errorHandler.Write(sw, appError)
			if err != nil {
				pipelineData.AddLoggerKV(
					slog.String("response_error", err.Error()),
				)
			} else {
				pipelineData.AddLoggerKV(
					slog.String("response_error", appError.String()),
				)
			}
		}

		pipelineData.logger = pipelineData.logger.With(
			slog.Int("response_status_code", sw.statusCode),
		)
	})
}
