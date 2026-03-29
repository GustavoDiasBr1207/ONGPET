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

	"ongpet/controllers"
	"ongpet/database"
	"ongpet/utils"

	"github.com/joho/godotenv"
	_ "ongpet/docs" // Swagger
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env não encontrado, usando variáveis do sistema")
	}

	// 🔍 Debug: imprime as variáveis de ambiente
	debugEnv()

	// 🔌 Conexão única — RLS controlado pelo JWT injetado em cada sessão
	database.Connect(os.Getenv("DATABASE_URL"))

	utils.InitSupabase()

	// 📧 Inicializar mailer para envio de emails
	if err := utils.InitMailer(); err != nil {
		log.Println("⚠️ Falha ao inicializar mailer:", err)
	}

	// 🔧 Rodar migrações de tabelas e índices
	database.RunMigrations()

	r := controllers.SetupRoutes()

	log.Println("🚀 API rodando em http://localhost:8080")
	r.Run(":8080")
}