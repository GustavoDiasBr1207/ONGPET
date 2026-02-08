package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	_ "ongpet/docs"

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
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env não encontrado, usando variáveis do sistema")
	}

	db := database.Connect(os.Getenv("DATABASE_URL"))

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

	if os.Getenv("GIN_MODE") != "" {
		gin.SetMode(os.Getenv("GIN_MODE"))
	}

	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/ongs", GetOngs)
		api.POST("/ongs", CreateOng)
		api.GET("/pets", GetPets)
		api.POST("/pets", CreatePet)
	}

	log.Println("🚀 API rodando em http://localhost:8080")
	r.Run(":8080")
}

// Handlers ONG
func GetOngs(c *gin.Context) {
	db := database.GetDB()
	var ongs []models.Ong
	db.Find(&ongs)
	c.JSON(200, ongs)
}

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

// Handlers Pet
func GetPets(c *gin.Context) {
	db := database.GetDB()
	var pets []models.Pet
	db.Find(&pets)
	c.JSON(200, pets)
}

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
