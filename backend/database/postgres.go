package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var db *gorm.DB

func Connect(dsn string) *gorm.DB {
	var err error

	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // 🔥 DESLIGA prepared statements do pgx
	}), &gorm.Config{
		PrepareStmt: false, // 🔥 DESLIGA prepared statements do GORM
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 🔒 sem plural automático
		},
	})

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
