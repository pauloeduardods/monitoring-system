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
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Gin struct {
	logger    logger.Logger
	Gin       *gin.Engine
	validator validator.Validator
	factory   *factory.Factory
}

func New(logger logger.Logger, factory *factory.Factory, validator validator.Validator) *Gin {
	gin := gin.Default()
	return &Gin{
		logger:    logger,
		Gin:       gin,
		validator: validator,
		factory:   factory,
	}
}

func (s *Gin) SetupCors() {
	s.Gin.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Ajuste a origem do seu frontend aqui
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func (s *Gin) SetupMiddlewares() {
	s.Gin.Use(middleware.RecoveryHandler(s.logger))
	s.Gin.Use(gin.Logger())
	s.Gin.Use(middleware.ErrorHandler(s.logger))

}

func (s *Gin) SetupApi(ctx context.Context, staticFilesPath string) error {
	//Api Routes
	s.Gin.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	s.Gin.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/home")
	})

	apiRoutes := s.Gin.Group("/api/v1")

	//Middlewares
	authMiddleware := middleware.NewAuthMiddleware(s.factory.UserManager.Infra.AuthService, s.factory.UserManager.Infra.AuthRepo, s.logger)

	//Websocket
	ginWs := apiRoutes.Group("/ws")
	wsServer := websocket.NewWebSocketServer(ctx, s.logger, ginWs, s.factory, authMiddleware)
	err := wsServer.Start()
	if err != nil {
		return err
	}

	//Static files
	s.Gin.StaticFS("/web", http.Dir(staticFilesPath))

	//Handlers
	authHandler := handlers.NewAuthHandler(s.factory.UserManager.UseCases, s.validator)
	monitorHandlers := handlers.NewCameraHandler(s.factory.Monitoring.UseCases)

	//Routes
	routes.ConfigAuthRoutes(apiRoutes, authHandler, authMiddleware)
	routes.ConfigMonitoringRoutes(apiRoutes, monitorHandlers, authMiddleware)
	return nil
}
