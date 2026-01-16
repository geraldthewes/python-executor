package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/geraldthewes/python-executor/docs/swagger"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(server *Server, logger *logrus.Logger) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(Logger(logger))
	router.Use(Recovery(logger))
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Execution endpoints
		v1.POST("/exec/sync", server.ExecuteSync)
		v1.POST("/exec/async", server.ExecuteAsync)
		v1.GET("/executions/:id", server.GetExecution)
		v1.DELETE("/executions/:id", server.KillExecution)

		// Simple JSON execution endpoint (Replit/Piston-compatible)
		v1.POST("/eval", server.ExecuteEval)
	}

	// Swagger documentation
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
