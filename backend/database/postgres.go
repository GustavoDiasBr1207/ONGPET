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

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("❌ Erro ao obter SQL DB:", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

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

// SetDB substitui o db global — use apenas em testes.
func SetDB(d *gorm.DB) {
	db = d
}

// isPostgres retorna true se a conexão ativa for Postgres.
// Usado para pular comandos exclusivos do Postgres em testes com SQLite.
func isPostgres() bool {
	if db == nil {
		return false
	}
	return db.Dialector.Name() == "postgres"
}

// WithUserDB executa fn dentro de uma transação com a identidade do usuário
// injetada via request.jwt.claims, ativando o RLS do Supabase corretamente.
//
// Corrige três problemas do GetUserDB anterior:
//   - A: a sessão configurada era descartada e uma nova (limpa) era retornada
//   - C: SET LOCAL vazava para outras conexões do pool — a transação prende
//        todo o bloco na mesma conexão do início ao fim
//   - D: o token agora é o JSON dos claims (não o JWT base64), permitindo
//        que auth.uid() leia o "sub" corretamente
//
// Em testes com SQLite os comandos Postgres são pulados automaticamente.
// Use em todas as rotas autenticadas no lugar de GetUserDB.
func WithUserDB(token string, fn func(tx *gorm.DB) error) error {
	if db == nil {
		log.Fatal("❌ Banco de dados não inicializado")
	}

	pg := isPostgres()

	return db.Transaction(func(tx *gorm.DB) error {
		if pg {
			// Injeta o JSON dos claims na transação.
			// SET LOCAL garante que o valor existe apenas dentro desta transação
			// e é descartado automaticamente ao final — sem risco de vazamento
			// entre conexões do pool.
			if err := tx.Exec(
				"SELECT set_config('request.jwt.claims', ?, true)", token,
			).Error; err != nil {
				return err
			}

			if err := tx.Exec("SET LOCAL ROLE authenticated").Error; err != nil {
				return err
			}
		}

		return fn(tx)
	})
}