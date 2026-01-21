package api

import (
	"github.com/gin-gonic/gin"
	"singbox-web/internal/api/handlers"
	"singbox-web/internal/api/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Auth routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.Login)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Auth
			protected.POST("/auth/logout", handlers.Logout)
			protected.PUT("/auth/password", handlers.ChangePassword)

			// System routes
			system := protected.Group("/system")
			{
				system.GET("/status", handlers.GetSystemStatus)
				system.POST("/start", handlers.StartSingbox)
				system.POST("/stop", handlers.StopSingbox)
				system.POST("/restart", handlers.RestartSingbox)
				system.GET("/version", handlers.GetSystemVersion)
				system.POST("/upgrade", handlers.UpgradeSingbox)
				system.GET("/config", handlers.GetGeneratedConfig)
			}
		}
	}

	return r
}
