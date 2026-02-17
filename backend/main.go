// @title OngPet API
// @version 1.0
// @description API para gerenciar ONGs e pets
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	"log"
	"os"
	"strings"

	"ongpet/controllers"
	"ongpet/database"
	"ongpet/models"

	"github.com/joho/godotenv"
	_ "ongpet/docs" // Swagger
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env não encontrado, usando variáveis do sistema")
	}

		db := database.Connect(os.Getenv("DATABASE_URL"))

		err := db.AutoMigrate(
		// base
		&models.Ong{},

		// dependem de Ong
		&models.Pet{},
		&models.PedidoAdocao{},

		// formulário
		&models.FormularioModelo{},
		&models.CampoFormulario{},

		// depende de PedidoAdocao + CampoFormulario
		&models.RespostaFormulario{},
	)

	if err != nil {
		errMsg := err.Error()

		// ignora APENAS erro de tabela já existente
		if strings.Contains(errMsg, "already exists") ||
			strings.Contains(errMsg, "42P07") {

			log.Println("⚠️ Tabelas já existem, ignorando AutoMigrate")

		} else {
			log.Fatal("❌ Erro ao rodar migrations:", err)
		}
	}

	r := controllers.SetupRoutes()

	log.Println("🚀 API rodando em http://localhost:8080")
	r.Run(":8080")
}
