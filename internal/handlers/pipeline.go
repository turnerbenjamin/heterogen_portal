package handlers

import (
	"net/http"
	"strconv"

	"github.com/turnerbenjamin/go_gbf/internal/etc"
	"github.com/turnerbenjamin/go_gbf/internal/logging"
)

type (
	AppHandlerWithRaft[T any] func(http.ResponseWriter, *http.Request, logging.Logger, T) etc.AppError
	MiddlewareWithRaft[T any] func(AppHandlerWithRaft[T]) AppHandlerWithRaft[T]
	AppHandler                func(http.ResponseWriter, *http.Request, logging.Logger) etc.AppError
	Middleware                func(AppHandler) AppHandler
)

type PipelineWithRaftBuilder[T any] interface {
	New(
		middlewares []MiddlewareWithRaft[T],
		handler AppHandlerWithRaft[T],
	) http.HandlerFunc
}

type pipelineWithRaftBuiler[T any] struct {
	newSeed      func(*http.Request) T
	newLogger    func(*http.Request) logging.Logger
	errorHandler ErrorHandler
}

type StatusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewPipelineWithRaftBuilder[T any](
	newSeed func(*http.Request) T,
	newLogger func(*http.Request) logging.Logger,
	errorHandler ErrorHandler,
) PipelineWithRaftBuilder[T] {
	return &pipelineWithRaftBuiler[T]{
		newSeed:      newSeed,
		newLogger:    newLogger,
		errorHandler: errorHandler,
	}
}

func (p *pipelineWithRaftBuiler[T]) New(
	middlewares []MiddlewareWithRaft[T],
	handler AppHandlerWithRaft[T],
) http.HandlerFunc {
	pipeline := handler

	i := len(middlewares) - 1
	for i >= 0 {
		pipeline = middlewares[i](pipeline)
		i--
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seed := p.newSeed(r)
		logger := p.newLogger(r)
		sw := StatusResponseWriter{ResponseWriter: w}

		err := pipeline(&sw, r, logger, seed)
		if err != nil {
			p.errorHandler.Write(w, logger, err)
		}

		logger.AddKV("status_code", strconv.Itoa(sw.statusCode))
		_ = logger.WriteLog()
	})
}

type PipelineBuilder interface {
	New(
		middlewares []Middleware,
		handler AppHandler,
	) http.HandlerFunc
}

type pipelineBuilder struct {
	newLogger    func(*http.Request) logging.Logger
	errorHandler ErrorHandler
}

func NewPipelineBuilder(
	newLogger func(*http.Request) logging.Logger,
	errorHandler ErrorHandler,
) PipelineBuilder {
	return &pipelineBuilder{
		newLogger:    newLogger,
		errorHandler: errorHandler,
	}
}

func (pb *pipelineBuilder) New(
	middlewares []Middleware,
	handler AppHandler,
) http.HandlerFunc {
	pipeline := handler

	i := len(middlewares) - 1
	for i >= 0 {
		pipeline = middlewares[i](pipeline)
		i--
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := pb.newLogger(r)
		sw := StatusResponseWriter{ResponseWriter: w}

		err := pipeline(&sw, r, logger)
		if err != nil {
			pb.errorHandler.Write(w, logger, err)
		}

		logger.AddKV("status_code", strconv.Itoa(sw.statusCode))
		_ = logger.WriteLog()
	})
}

func (w *StatusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *StatusResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}
