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
	"ongpet/utils"

	_ "ongpet/docs" // Swagger

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env não encontrado, usando variáveis do sistema")
	}

	// 🔍 Debug: imprime as variáveis de ambiente
	debugEnv()

	// 🔌 Conexão única — RLS controlado pelo JWT injetado em cada sessão
	db := database.Connect(os.Getenv("DATABASE_URL"))

	utils.InitSupabase()

	// 📧 Inicializar mailer para envio de emails
	if err := utils.InitMailer(); err != nil {
		log.Println("⚠️ Falha ao inicializar mailer:", err)
	}

	// 💬 Inicializar WhatsApp service para envio de mensagens
	if err := utils.InitWhatsAppService(); err != nil {
		log.Println("⚠️ Falha ao inicializar WhatsApp service:", err)
	}

	// 📅 Iniciar agendador de lembretes de acompanhamento (roda às 19:05 BRT)
	utils.StartAcompanhamentoReminderScheduler()

	err := db.AutoMigrate(
		// base
		&models.Ong{},

		// dependem de Ong
		&models.Pet{},
		&models.PetImage{},
		&models.PedidoAdocao{},

		// banner
		&models.Banner{},

		// formulário
		&models.FormularioModelo{},
		&models.CampoFormulario{},

		// depende de PedidoAdocao + CampoFormulario
		&models.RespostaFormulario{},
	)

	if err != nil {
		errMsg := err.Error()

		if strings.Contains(errMsg, "already exists") ||
			strings.Contains(errMsg, "42P07") {
			log.Println("⚠️ Tabelas já existem, ignorando AutoMigrate")
		} else {
			log.Fatal("❌ Erro ao rodar migrations:", err)
		}
	}

	// 🔧 Rodar migrações de tabelas e índices
	database.RunMigrations()

	r := controllers.SetupRoutes()

	log.Println("🚀 API rodando em http://localhost:8080")
	r.Run(":8080")
}