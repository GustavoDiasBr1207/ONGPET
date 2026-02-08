package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func Connect(dsn string) *gorm.DB {
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Erro ao conectar no banco:", err)
	}
	log.Println("✅ Conexão com o banco realizada com sucesso")
	return db
}

func GetDB() *gorm.DB {
	if db == nil {
		log.Fatal("❌ Banco de dados não inicializado")
	}
	return db
}
