package gin_server

import (
	"context"
	"monitoring-system/cmd/factory"
	"monitoring-system/cmd/server/gin_server/handlers"
	"monitoring-system/cmd/server/gin_server/middleware"
	"monitoring-system/cmd/server/gin_server/routes"
	"monitoring-system/cmd/server/websocket"
	"monitoring-system/pkg/logger"
	"monitoring-system/pkg/validator"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Gin struct {
	log       logger.Logger
	Gin       *gin.Engine
	validator validator.Validator
	ctx       context.Context
	factory   *factory.Factory
}

func New(ctx context.Context, logger logger.Logger, factory *factory.Factory, validator validator.Validator) *Gin {
	gin := gin.Default()
	return &Gin{
		log:       logger,
		Gin:       gin,
		validator: validator,
		ctx:       ctx,
		factory:   factory,
	}
}

func (s *Gin) SetupCors() {
	cors := middleware.Cors{
		Origin:      "*",
		Methods:     "GET, POST, PUT, DELETE, OPTIONS",
		Headers:     "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token, X-Auth-Token, X-Requested-With",
		Credentials: false,
	}
	s.Gin.Use(cors.CorsMiddleware())
}

func (s *Gin) SetupMiddlewares() {
	s.Gin.Use(gin.CustomRecovery(middleware.RecoveryHandler(s.log)))
	s.Gin.Use(gin.Logger())
	s.Gin.Use(middleware.ErrorHandler(s.log))

}

func (s *Gin) SetupApi() error {
	//Api Routes
	s.Gin.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	apiRoutes := s.Gin.Group("/api/v1")

	//Middlewares
	authMiddleware := middleware.NewAuthMiddleware(s.factory.Services.Auth)

	//Websocket
	ginWs := apiRoutes.Group("/ws")
	wsServer := websocket.NewWebSocketServer(s.ctx, s.log, ginWs, s.factory, authMiddleware)
	err := wsServer.Start()
	if err != nil {
		return err
	}

	//Static files
	s.Gin.StaticFS("/web", http.Dir("web/static"))

	//Handlers
	authHandler := handlers.NewAuthHandler(s.factory.Services.Auth, s.validator)

	//Routes
	routes.ConfigAuthRoutes(apiRoutes, authHandler)
	return nil
}
