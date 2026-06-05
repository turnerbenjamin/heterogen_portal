package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/turnerbenjamin/heterogen_portal/internal/etc"
)

type PipelineContext[T any] struct {
	logger *slog.Logger
	state  *T
}
type NoPipelineState struct{}

type (
	AppHandler[T any] func(http.ResponseWriter, *http.Request, *PipelineContext[T]) *etc.AppError
	Middleware[T any] func(AppHandler[T]) AppHandler[T]
)

type PipelineBuilder[T any] struct {
	errorHandler ErrorWriter
	rootLogger   *slog.Logger
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

func NewPipelineBuilder[T any](
	errorHandler ErrorWriter,
	logSink io.Writer,
) *PipelineBuilder[T] {
	return &PipelineBuilder[T]{
		errorHandler: errorHandler,
		rootLogger:   slog.New(slog.NewJSONHandler(logSink, &slog.HandlerOptions{})),
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
		sw := &statusSpyWriter{ResponseWriter: w}

		startTime := time.Now()
		slogger := p.rootLogger.With(
			slog.String("request_method", r.Method),
			slog.String("request_path", r.URL.Path),
			slog.String("request_time", startTime.Format(time.RFC3339Nano)),
		)

		defer func() {
			if rec := recover(); rec != nil {
				slogger.Error(
					"panic",
					slog.String("panic", fmt.Sprint(rec)),
					slog.String("stack", string(debug.Stack())),
				)
				http.Error(sw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		pipelineData := &PipelineContext[T]{
			logger: slogger,
			state:  new(T),
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

		appError := pipeline(sw, r, pipelineData)
		if appError != nil {
			err := p.errorHandler.Write(sw, appError)
			if err != nil {
				pipelineData.AddLoggerKV(
					slog.String("writer_error", err.Error()),
					slog.String("response_error", appError.Error()),
				)
				http.Error(sw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			} else {
				pipelineData.AddLoggerKV(
					slog.String("response_error", appError.Error()),
				)
			}
		}

		pipelineData.logger = pipelineData.logger.With(
			slog.Int("response_status_code", sw.statusCode),
		)

	})
}
