package api

import (
	"monitoring-system/cmd/api/gin/handlers"
	"monitoring-system/cmd/api/gin/middleware"
	"monitoring-system/cmd/api/gin/routes"
	"monitoring-system/internal/auth"

	"github.com/gin-gonic/gin"
)

func (s *Server) SetupCors() {
	cors := middleware.Cors{
		Origin:      "*",
		Methods:     "GET, POST, PUT, DELETE, OPTIONS",
		Headers:     "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token, X-Auth-Token, X-Requested-With",
		Credentials: false,
	}
	s.gin.Use(cors.CorsMiddleware())
}

func (s *Server) SetupMiddlewares() {
	s.gin.Use(gin.CustomRecovery(middleware.RecoveryHandler(s.log)))
	s.gin.Use(gin.Logger())
	s.gin.Use(middleware.ErrorHandler(s.log))

}

func (s *Server) SetupApi() error {
	// static.SetupStaticFiles(s.gin)

	//Repositories
	authRepository, err := auth.NewAuthRepository(s.sqlDB, s.log)
	if err != nil {
		s.log.Error("Error creating auth repository %v", err)
		return err
	}

	//Services

	authService, err := auth.NewAuthService(authRepository, s.log)
	if err != nil {
		s.log.Error("Error creating auth service %v", err)
		return err
	}

	//Handlers

	authHandler := handlers.NewAuthHandler(authService, s.validator)

	//Middlewares

	// authMiddleware := middleware.NewAuthMiddleware(authService)

	//Routes
	routes.ConfigAuthRoutes(s.gin, authHandler)
	return nil
}
