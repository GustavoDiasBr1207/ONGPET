package controllers

import (
	"net/http"

	v0 "ongpet/controllers/v1" // importa os handlers
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

// SetupRoutes configura todas as rotas da API OngPet
func SetupRoutes() *gin.Engine {
	r := gin.Default() // usa Default para já ter Logger e Recovery

	// CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Grupo API v1
	api := r.Group("/api/v1")
	{
		// Auth
		api.POST("/auth/login", v0.Login)

		// ONG
		api.GET("/ongs", v0.GetOngs)
		api.POST("/ongs", RequireAuth(), v0.CreateOng)

		// Pet
		api.GET("/pets", v0.GetPets)
		api.POST("/pets", RequireAuth(), v0.CreatePet)
	}

	// Rota default para não encontradas
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "rota não encontrada"})
	})

	return r
}
