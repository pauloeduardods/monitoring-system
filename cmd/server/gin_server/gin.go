package gin_server

import (
	"context"
	"monitoring-system/cmd/server/gin_server/handlers"
	"monitoring-system/cmd/server/gin_server/middleware"
	"monitoring-system/cmd/server/gin_server/routes"
	"monitoring-system/cmd/server/modules"
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
	modules   *modules.Modules
}

func New(ctx context.Context, logger logger.Logger, modules *modules.Modules, validator validator.Validator) *Gin {
	gin := gin.Default()
	return &Gin{
		log:       logger,
		Gin:       gin,
		validator: validator,
		ctx:       ctx,
		modules:   modules,
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

	//Websocket
	ginWs := s.Gin.Group("/ws")
	wsServer := websocket.NewWebSocketServer(s.ctx, s.log, ginWs, s.modules.Internal.Camera)
	wsServer.Start()

	//Static files
	s.Gin.StaticFS("/static", http.Dir("web/static"))

	//Handlers
	authHandler := handlers.NewAuthHandler(s.modules.Services.Auth, s.validator)

	//Middlewares
	// authMiddleware := middleware.NewAuthMiddleware(s.modules.Services.Auth)

	//Routes
	routes.ConfigAuthRoutes(s.Gin, authHandler)
	return nil
}
