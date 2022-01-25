package http

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/waldenlake/go-kit/transport"
	"go.uber.org/zap"
	"net"
	"net/http"
	"sync"
	"time"
)

var _ transport.Server = (*Server)(nil)

type ServerOption func(*Server)

func NetWork(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

func Address(address string) ServerOption {
	return func(s *Server) {
		s.address = address
	}
}

func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

func Logger(logger *zap.SugaredLogger) ServerOption {
	return func(s *Server) {
		s.log = logger
	}
}

type Server struct {
	*http.Server

	once sync.Once

	network string
	address string
	timeout time.Duration

	router *gin.Engine

	log *zap.SugaredLogger
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network: "tcp",
		address: "0.0.0.0:8000",
		timeout: 1 * time.Second,
		log:     zap.S(),
	}
	for _, opt := range opts {
		opt(srv)
	}
	srv.router = gin.Default()
	srv.Server = &http.Server{
		Handler: srv.router,
	}
	return srv
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen(s.network, s.address)
	if err != nil {
		return err
	}
	s.BaseContext = func(net.Listener) context.Context {
		return ctx
	}
	s.log.Infof("[HTTP] server listening on: %s", lis.Addr().String())
	if err := s.Serve(lis); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Close(ctx context.Context) error {
	s.log.Info("[HTTP] server stopping")
	return s.Shutdown(ctx)
}
