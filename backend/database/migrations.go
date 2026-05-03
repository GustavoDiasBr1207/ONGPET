package database

import (
	"fmt"
	"log"

	"ongpet/models"
)

// RunMigrations executa todas as migrações de criação de tabelas e índices
func RunMigrations() {
	if db == nil {
		log.Fatal("❌ Banco de dados não inicializado")
	}

	// Auto migrate das tabelas
	if err := db.AutoMigrate(
		&models.Ong{},
		&models.Pet{},
		&models.PetImage{},
		&models.FormularioModelo{},
		&models.CampoFormulario{},
		&models.PedidoAdocao{},
		&models.RespostaFormulario{},
		&models.Acompanhamento{},
		&models.LogAcompanhamento{},
	); err != nil {
		log.Fatal("❌ Erro na migração de tabelas:", err)
	}

	log.Println("✅ Migrações de tabelas finalizadas")

	// ─── Corrige índices únicos da tabela ONG para respeitar soft delete ───
	// O GORM recria esses índices simples a cada AutoMigrate, então
	// derrubamos e recriamos como índices parciais (WHERE deleted_at IS NULL)
	fixIndices := []string{
		`DROP INDEX IF EXISTS public.unique_ong_owner`,
		`DROP INDEX IF EXISTS public.uni_ong_email`,

		`CREATE UNIQUE INDEX IF NOT EXISTS unique_ong_owner_active 
		 ON public.ong (user_id) WHERE deleted_at IS NULL`,

		`CREATE UNIQUE INDEX IF NOT EXISTS uni_ong_email_active 
		 ON public.ong (email) WHERE deleted_at IS NULL`,
	}

	for _, sql := range fixIndices {
		if err := db.Exec(sql).Error; err != nil {
			fmt.Printf("⚠️ Aviso ao corrigir índice ONG: %v\n", err)
		}
	}

	log.Println("✅ Índices únicos da ONG corrigidos para soft delete")

	// ─── Índices de performance (partial index — soft delete) ───
	indices := []string{
		// PedidoAdocao
		`CREATE INDEX IF NOT EXISTS idx_pedido_adocao_ong_id 
		 ON pedido_adocao(ong_id) 
		 WHERE deleted_at IS NULL`,

		`CREATE INDEX IF NOT EXISTS idx_pedido_adocao_ong_created 
		 ON pedido_adocao(ong_id, created_at DESC) 
		 WHERE deleted_at IS NULL`,

		// CampoFormulario
		`CREATE INDEX IF NOT EXISTS idx_campo_formulario_modelo_id 
		 ON campo_formulario(formulario_modelo_id, ordem) 
		 WHERE deleted_at IS NULL`,

		// RespostaFormulario
		`CREATE INDEX IF NOT EXISTS idx_resposta_formulario_pedido_id 
		 ON resposta_formulario(pedido_adocao_id) 
		 WHERE deleted_at IS NULL`,

		// Pet
		`CREATE INDEX IF NOT EXISTS idx_pet_ong_id 
		 ON pet(ong_id) 
		 WHERE deleted_at IS NULL`,

		// PetImagem
		`CREATE INDEX IF NOT EXISTS idx_pet_image_pet_id 
		 ON pet_imagem(pet_id) 
		 WHERE deleted_at IS NULL`,

		// FormularioModelo
		`CREATE INDEX IF NOT EXISTS idx_formulario_modelo_ong_id 
		 ON formulario_modelo(ong_id) 
		 WHERE deleted_at IS NULL`,

		// ONG
		`CREATE INDEX IF NOT EXISTS idx_ong_email 
		 ON ong(email) 
		 WHERE deleted_at IS NULL`,
	}

	created := 0
	for _, indexSQL := range indices {
		if err := db.Exec(indexSQL).Error; err != nil {
			fmt.Printf("⚠️ Aviso ao criar índice: %v\n", err)
		} else {
			created++
		}
	}

	log.Printf("✅ %d índices criados/verificados para otimização de queries\n", created)
}