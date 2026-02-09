package main

import (
	"log"
	"os"
	"ongpet/controllers"
	"ongpet/database"
	"ongpet/models"

	"github.com/joho/godotenv"
	_ "ongpet/docs" // importa os docs do Swagger
)

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

	r := controllers.SetupRoutes()

	log.Println("🚀 API rodando em http://localhost:8080")
	r.Run(":8080")
}
