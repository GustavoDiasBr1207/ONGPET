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
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

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

	debugEnv()

	db := database.Connect(os.Getenv("DATABASE_URL"))

	utils.InitSupabase()

	if err := utils.InitMailer(); err != nil {
		log.Println("⚠️ Falha ao inicializar mailer:", err)
	}

	if err := utils.InitWhatsAppService(); err != nil {
		log.Println("⚠️ Falha ao inicializar WhatsApp service:", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	utils.StartAcompanhamentoReminderScheduler(ctx)

	err := db.AutoMigrate(
		&models.Ong{},
		&models.Pet{},
		&models.PetImage{},
		&models.PedidoAdocao{},
		&models.Banner{},
		&models.FormularioModelo{},
		&models.CampoFormulario{},
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

	// 🔧 Roda APÓS o AutoMigrate para garantir que índices parciais
	// sobrescrevam qualquer constraint recriada pelo GORM
	database.RunMigrations()

	r := controllers.SetupRoutes()

	log.Println("🚀 API rodando em http://localhost:8080")
	r.Run(":8080")
}