package controllers

import (
	"net/http"
	"time"

	v1 "ongpet/controllers/v1"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configura todas as rotas da API OngPet
func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// CORS
	config := cors.Config{
		AllowOrigins: []string{
			"https://adote-serra.vercel.app",
			"http://localhost:3000",
			"http://localhost:5173",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}

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
		// Auth
		api.POST("/auth/login", v1.Login)

		// ONG
		api.GET("/ongs", Handle(v1.ReadOngs))
		api.GET("/ongs/:id", Handle(v1.ReadOng))
		api.POST("/ongs", RequireAuth(), Handle(v1.CreateOng))
		api.PUT("/ongs/:id", RequireAuth(), Handle(v1.UpdateOng))
		api.DELETE("/ongs/:id", RequireAuth(), Handle(v1.DeleteOng))
		api.POST("/ongs/:id/logo", RequireAuth(), Handle(v1.UploadOngLogo))
		api.DELETE("/ongs/:id/logo", RequireAuth(), Handle(v1.DeleteOngLogo))

		// Pet
		api.GET("/pets", Handle(v1.ReadPets))
		api.GET("/pets/:id", Handle(v1.ReadPet))
		api.POST("/pets", RequireAuth(), Handle(v1.CreatePet))
		api.PUT("/pets/:id", RequireAuth(), Handle(v1.UpdatePet))
		api.DELETE("/pets/:id", RequireAuth(), Handle(v1.DeletePet))
		api.POST("/pets/:id/imagens", RequireAuth(), Handle(v1.UploadPetImages))
		api.DELETE("/pets/:id/imagens/:imageId", RequireAuth(), Handle(v1.DeletePetImage))

		// Formulário
		api.GET("/formularios", Handle(v1.ReadFormularios))
		api.GET("/formularios/:id", Handle(v1.ReadFormulario))
		api.POST("/formularios", RequireAuth(), Handle(v1.CreateFormulario))
		api.PUT("/formularios/:id", RequireAuth(), Handle(v1.UpdateFormulario))
		api.DELETE("/formularios/:id", RequireAuth(), Handle(v1.DeleteFormulario))
		api.POST("/formularios/:id/imagem", RequireAuth(), Handle(v1.UploadFormularioImagem))
		api.DELETE("/formularios/:id/imagem", RequireAuth(), Handle(v1.DeleteFormularioImagem))

		// Campos do formulário
		api.POST("/formularios/:id/campos", RequireAuth(), Handle(v1.CreateCampoFormulario))
		api.PUT("/formularios/:id/campos/:campoId", RequireAuth(), Handle(v1.UpdateCampoFormulario))
		api.DELETE("/formularios/:id/campos/:campoId", RequireAuth(), Handle(v1.DeleteCampoFormulario))

		// Pedidos de adoção
		api.GET("/pedidos-adocao", RequireAuth(), Handle(v1.ReadPedidosAdocao))
		api.GET("/pedidos-adocao/:id", RequireAuth(), Handle(v1.ReadPedidoAdocao))
		api.POST("/pedidos-adocao", Handle(v1.CreatePedidoAdocao))
		api.PUT("/pedidos-adocao/:id/status", RequireAuth(), Handle(v1.UpdateStatusPedidoAdocao))
		api.DELETE("/pedidos-adocao/:id", RequireAuth(), Handle(v1.DeletePedidoAdocao))
		api.POST("/pedidos-adocao/:pedidoId/respostas/:respostaId/imagem", Handle(v1.UploadRespostaImagem))

		// Testes de Email
		api.GET("/test/email-config", Handle(v1.CheckEmailConfig))
		api.POST("/test/email", Handle(v1.SendTestEmail))
	}

	// 404
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "rota não encontrada"})
	})

	return r
}