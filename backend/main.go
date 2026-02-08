package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ongpet/database"
	"ongpet/models"
)

func main() {
	// 🔹 carrega variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env não encontrado, usando variáveis do sistema")
	}

	// 🔹 conecta no banco
	db := database.Connect(os.Getenv("DATABASE_URL"))

	// 🔹 roda migrations
	err := db.AutoMigrate(
		&models.Ong{},
		&models.Pet{},
		&models.PedidoAdocao{},
		&models.FormularioModelo{},
		&models.CampoFormulario{},
		&models.RespostaFormulario{},
	)
	if err != nil {
		log.Fatal("❌ Erro ao rodar migrations:", err)
	}

	// 🔹 modo de execução
	if os.Getenv("GIN_MODE") != "" {
		gin.SetMode(os.Getenv("GIN_MODE"))
	}

	// 🔹 cria servidor
	r := gin.Default()

	// 🔹 health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Println("🚀 API rodando em http://localhost:8080")
	r.Run(":8080")
}
