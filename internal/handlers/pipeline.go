// Package handlers contains HTTP handlers for the application.
//
// This file provides a Pipeline which controls execution of the handler and any
// middleware for a given route. The pipeline is responsible for logging request
// data, handling errors and panic recovery. Generics are used to allow typed
// context to be shared by middlewares and the handler
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

// PipelineContext holds per-request logger and state for the pipeline.
type PipelineContext[T any] struct {
	logger *slog.Logger
	state  *T
}

// NoState is a placeholder type for handlers that do not require
// pipeline state.
type NoState struct{}

type (
	// AppHandler is the signature for a handler in the pipeline. Pipeline state
	// is passed to the handler in PipelineContext and any AppErrors returned
	// will be handled by the pipeline.
	AppHandler[T any] func(http.ResponseWriter, *http.Request, *PipelineContext[T]) *etc.AppError

	// Middleware is the signature for a middleware function in the pipeline.
	// This receives the next AppHandler in the pipeline and returns a new
	// AppHandler. The returned AppHandler can be used, for instance, to perform
	// validations and exit early with an AppError if validations fail, or to
	// enrich the PipelineContext before calling the next AppHandler in the
	// pipeline.
	Middleware[T any] func(AppHandler[T]) AppHandler[T]
)

// ErrorWriter writes an AppError to the provided ResponseWriter.
type ErrorWriter interface {
	Write(http.ResponseWriter, *etc.AppError) error
}

// PipelineBuilder is used as a factory for building a Pipeline
type PipelineBuilder[T any] struct {
	errorHandler ErrorWriter
	rootLogger   *slog.Logger
}

// NewPipelineBuilder creates a PipelineBuilder using the given error writer
// and log sink.
func NewPipelineBuilder[T any](
	errorHandler ErrorWriter,
	logSink io.Writer,
) *PipelineBuilder[T] {
	return &PipelineBuilder[T]{
		errorHandler: errorHandler,
		rootLogger:   slog.New(slog.NewJSONHandler(logSink, &slog.HandlerOptions{})),
	}
}

// New composes the provided middlewares and handler into an http.HandlerFunc.
// It initializes request context, recovers from panics, records timing, and
// delegates error responses to the configured ErrorWriter.
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
					slog.String("response_error", appError.String()),
				)
				http.Error(sw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

// AddLoggerKV appends key/value attributes to the pipeline logger.
func (d *PipelineContext[T]) AddLoggerKV(attrs ...slog.Attr) {
	anyAttrs := make([]any, len(attrs))
	for i, attr := range attrs {
		anyAttrs[i] = attr
	}
	d.logger = d.logger.With(anyAttrs...)
}
