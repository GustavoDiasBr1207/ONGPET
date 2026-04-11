package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var db *gorm.DB

func gormConfig() *gorm.Config {
	return &gorm.Config{
		PrepareStmt: false,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}
}

// Connect inicializa a conexão com o banco.
// Use a DATABASE_URL com a senha do banco (não service_role key).
func Connect(dsn string) *gorm.DB {
	var err error
	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), gormConfig())
	if err != nil {
		log.Fatal("❌ Erro ao conectar no banco:", err)
	}

	// Configurar connection pool para melhor performance
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("❌ Erro ao obter SQL DB:", err)
	}
	sqlDB.SetMaxOpenConns(10)         // máximo de conexões abertas
	sqlDB.SetMaxIdleConns(5)          // manter 5 conexões idle
	sqlDB.SetConnMaxLifetime(time.Hour) // renovar conexões a cada hora

	log.Println("✅ Conexão com o banco inicializada com pool otimizado")
	return db
}

// GetDB retorna a conexão padrão.
// Use para rotas públicas (sem token de usuário).
func GetDB() *gorm.DB {
	if db == nil {
		log.Fatal("❌ Banco de dados não inicializado")
	}
	return db
}

// GetUserDB retorna uma sessão com o JWT do usuário injetado,
// ativando auth.uid() nas políticas RLS do Supabase.
// Use em todas as rotas autenticadas.
func GetUserDB(token string) *gorm.DB {
	if db == nil {
		log.Fatal("❌ Banco de dados não inicializado")
	}

	session := db.Session(&gorm.Session{NewDB: true})

	// Injeta o JWT na sessão do Postgres → RLS consegue ler auth.uid()
	session.Exec("SELECT set_config('request.jwt', ?, true)", token)
	session.Exec("SET LOCAL role = authenticated")

	return session
}