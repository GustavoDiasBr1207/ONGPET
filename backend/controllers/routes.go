package controllers

import (
	"net/http"

	v0 "ongpet/controllers/v1"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configura todas as rotas da API OngPet
func SetupRoutes() *gin.Engine {
	r := gin.Default()

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
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1
	api := r.Group("/api/v1")
	{
		// Auth (NÃO retorna error)
		api.POST("/auth/login", v0.Login)

		// ONG (retornam error)
		api.GET("/ongs", Handle(v0.ReadOngs))
		api.POST("/ongs", RequireAuth(), Handle(v0.CreateOng))
		api.PUT("/ongs/:id", RequireAuth(), Handle(v0.UpdateOng))
		api.DELETE("/ongs/:id", RequireAuth(), Handle(v0.DeleteOng))

		// Pet (NÃO retornam error)
		api.GET("/pets", Handle(v0.ReadPets))
		api.POST("/pets", RequireAuth(), Handle(v0.CreatePet))
		api.POST("/pets/:id/imagens", RequireAuth(), Handle(v0.UploadPetImages))
		api.PUT("/pets/:id", RequireAuth(), Handle(v0.UpdatePet))
		api.DELETE("/pets/:id", RequireAuth(), Handle(v0.DeletePet))
	}

	// 404
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "rota não encontrada"})
	})

	return r
}
