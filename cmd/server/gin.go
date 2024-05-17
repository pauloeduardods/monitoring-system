package server

import (
	"monitoring-system/cmd/server/gin/handlers"
	"monitoring-system/cmd/server/gin/middleware"
	"monitoring-system/cmd/server/gin/routes"
	"monitoring-system/cmd/server/websocket"
	"monitoring-system/internal/auth"
	"net/http"

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

	//Websocket
	ginWs := s.gin.Group("/ws")
	wsServer := websocket.NewWebSocketServer(s.ctx, s.log, ginWs, s.cam)
	wsServer.Start()

	//Static files
	s.gin.StaticFS("/static", http.Dir("web/static"))

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
