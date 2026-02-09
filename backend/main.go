package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	_ "ongpet/docs" // importa os docs gerados pelo swag

	"ongpet/database"
	"ongpet/models"
)

// @title API OngPet
// @version 1.0
// @description API para gerenciar ONGs, Pets e adoções.
// @contact.name Gustavo Dias
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Carrega variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env não encontrado, usando variáveis do sistema")
	}

	// Conecta no banco
	db := database.Connect(os.Getenv("DATABASE_URL"))

	// Roda migrations
	if err := db.AutoMigrate(
		&models.Ong{},
		&models.Pet{},
		&models.FormularioModelo{},
		&models.CampoFormulario{},
		&models.RespostaFormulario{},
		&models.PedidoAdocao{},
	); err != nil {
		log.Fatal("❌ Erro ao rodar migrations:", err)
	}

	// Define modo de execução
	if os.Getenv("GIN_MODE") != "" {
		gin.SetMode(os.Getenv("GIN_MODE"))
	}

	// Cria servidor
	r := gin.Default()

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Rotas API v1
	api := r.Group("/api/v1")
	{
		// ONG
		api.GET("/ongs", GetOngs)
		api.POST("/ongs", CreateOng)

		// Pet
		api.GET("/pets", GetPets)
		api.POST("/pets", CreatePet)
	}

	log.Println("🚀 API rodando em http://localhost:8080")
	r.Run(":8080")
}

// ==========================
// Handlers com Swagger
// ==========================

// Ong Handlers

// @Summary Lista todas as ONGs
// @Description Retorna todas as ONGs cadastradas
// @Tags ONG
// @Produce json
// @Success 200 {array} models.Ong
// @Router /api/v1/ongs [get]
func GetOngs(c *gin.Context) {
	db := database.GetDB()
	var ongs []models.Ong
	db.Find(&ongs)
	c.JSON(200, ongs)
}

// @Summary Cria uma nova ONG
// @Description Cria uma ONG no sistema
// @Tags ONG
// @Accept json
// @Produce json
// @Param ong body models.Ong true "Nova ONG"
// @Success 201 {object} models.Ong
// @Failure 400 {object} map[string]string
// @Router /api/v1/ongs [post]
func CreateOng(c *gin.Context) {
	db := database.GetDB()
	var ong models.Ong
	if err := c.ShouldBindJSON(&ong); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	db.Create(&ong)
	c.JSON(201, ong)
}

// Pet Handlers

// @Summary Lista todos os Pets
// @Description Retorna todos os pets cadastrados
// @Tags Pet
// @Produce json
// @Success 200 {array} models.Pet
// @Router /api/v1/pets [get]
func GetPets(c *gin.Context) {
	db := database.GetDB()
	var pets []models.Pet
	db.Find(&pets)
	c.JSON(200, pets)
}

// @Summary Cria um novo Pet
// @Description Cria um Pet no sistema
// @Tags Pet
// @Accept json
// @Produce json
// @Param pet body models.Pet true "Novo Pet"
// @Success 201 {object} models.Pet
// @Failure 400 {object} map[string]string
// @Router /api/v1/pets [post]
func CreatePet(c *gin.Context) {
	db := database.GetDB()
	var pet models.Pet
	if err := c.ShouldBindJSON(&pet); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	db.Create(&pet)
	c.JSON(201, pet)
}
