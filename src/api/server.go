package server

import (
	"context"
	"monitoring-system/src/api/gin_server"
	"monitoring-system/src/config"
	"monitoring-system/src/factory"
	"monitoring-system/src/pkg/logger"
	"monitoring-system/src/pkg/validator"
	"net/http"
	"strconv"
)

type Server struct {
	log        logger.Logger
	config     *config.Config
	gin_server *gin_server.Gin
	server     *http.Server
	validator  validator.Validator
}

func New(config *config.Config, logger logger.Logger, factory *factory.Factory) *Server {
	gin := gin_server.New(logger, factory, validator.NewValidatorImpl())

	return &Server{
		config:     config,
		gin_server: gin,
		log:        logger,
		validator:  validator.NewValidatorImpl(),
	}
}

func (s *Server) Start(ctx context.Context, staticFilesPath string) error {
	s.log.Info("Starting server %s:%d", s.config.Api.Host, s.config.Api.Port)

	s.gin_server.SetupCors()
	s.gin_server.SetupMiddlewares()
	s.gin_server.SetupApi(ctx, staticFilesPath)

	go func() {
		<-ctx.Done()
		s.log.Info("Shutdown Server ...")

		if err := s.server.Shutdown(ctx); err != nil {
			s.log.Error("Server Shutdown: %v", err)
		}
		s.log.Info("Server exiting")
	}()

	s.server = &http.Server{
		Addr:    s.config.Api.Host + ":" + strconv.Itoa(s.config.Api.Port),
		Handler: s.gin_server.Gin,
	}

	err := s.server.ListenAndServe()
	if err != nil {
		s.log.Error("Error starting server: %v", err)
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		s.log.Info("Stopping server")
		if err := s.server.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
