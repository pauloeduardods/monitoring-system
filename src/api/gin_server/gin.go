package gin_server

import (
	"context"

	"monitoring-system/src/api/gin_server/handlers"
	"monitoring-system/src/api/gin_server/middleware"
	"monitoring-system/src/api/gin_server/routes"
	"monitoring-system/src/api/websocket"
	"monitoring-system/src/factory"
	"monitoring-system/src/pkg/logger"
	"monitoring-system/src/pkg/validator"
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

	s.Gin.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/home.html")
	})

	apiRoutes := s.Gin.Group("/api/v1")

	//Middlewares
	authMiddleware := middleware.NewAuthMiddleware(s.factory.UserManager.Infra.AuthService, s.factory.UserManager.Infra.AuthRepo)

	//Websocket
	ginWs := apiRoutes.Group("/ws")
	wsServer := websocket.NewWebSocketServer(s.ctx, s.log, ginWs, s.factory, authMiddleware)
	err := wsServer.Start()
	if err != nil {
		return err
	}

	//Static files
	s.Gin.StaticFS("/web", http.Dir("src/web/static"))

	//Handlers
	authHandler := handlers.NewAuthHandler(s.factory.UserManager.UseCases, s.validator)
	monitorHandlers := handlers.NewCameraHandler(s.factory.Monitoring.UseCases)

	//Routes
	routes.ConfigAuthRoutes(apiRoutes, authHandler, authMiddleware)
	routes.ConfigMonitoringRoutes(apiRoutes, monitorHandlers, authMiddleware)
	return nil
}
