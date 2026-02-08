package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ongpet/database"
)

func main() {
	// carrega .env
	_ = godotenv.Load()

	// conecta no banco
	db := database.Connect(os.Getenv("DATABASE_URL"))

	// evita warning de variável não usada
	_ = db

	// cria servidor gin
	r := gin.Default()

	// rota de teste
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	log.Println("🚀 API rodando em http://localhost:8080")
	r.Run(":8080")
}
