package api

import (
	"context"
	"database/sql"
	"monitoring-system/config"
	"monitoring-system/pkg/logger"
	"monitoring-system/pkg/validator"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/gin-gonic/gin"
)

type Server struct {
	log       logger.Logger
	config    *config.Config
	gin       *gin.Engine
	server    *http.Server
	validator validator.Validator
	ctx       context.Context
	sqlDB     *sql.DB
}

func New(ctx context.Context, awsConfig *aws.Config, config *config.Config, logger logger.Logger, sqlDB *sql.DB) *Server {
	gin := gin.Default()

	return &Server{
		config:    config,
		gin:       gin,
		log:       logger,
		validator: validator.NewValidatorImpl(),
		ctx:       ctx,
		sqlDB:     sqlDB,
	}
}

func (s *Server) Start() error {
	s.log.Info("Starting server %s:%d", s.config.Host, s.config.Port)
	s.SetupCors()
	s.SetupMiddlewares()
	s.SetupApi()

	go func() {
		<-s.ctx.Done()
		s.log.Info("Shutdown Server ...")

		if err := s.server.Shutdown(s.ctx); err != nil {
			s.log.Error("Server Shutdown: %v", err)
		}
		s.log.Info("Server exiting")
	}()

	s.server = &http.Server{
		Addr:    s.config.Host + ":" + strconv.Itoa(s.config.Port),
		Handler: s.gin,
	}

	return s.server.ListenAndServe()
}

func (s *Server) Stop() error {
	if s.server != nil {
		s.log.Info("Stopping server")
		if err := s.server.Shutdown(s.ctx); err != nil {
			return err
		}
	}
	return nil
}
