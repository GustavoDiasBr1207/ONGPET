package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func debugEnv() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️ .env não encontrado")
	}

	url := strings.TrimSpace(os.Getenv("SUPABASE_URL"))
	key := strings.TrimSpace(os.Getenv("SUPABASE_SERVICE_ROLE_KEY"))
	bucket := strings.TrimSpace(os.Getenv("SUPABASE_BUCKET"))

	fmt.Println("\n=== DEBUG: VARIÁVEIS DE AMBIENTE ===")
	fmt.Printf("SUPABASE_URL: %s\n", url)
	fmt.Printf("SUPABASE_SERVICE_ROLE_KEY: %s...\n", truncate(key, 30))
	fmt.Printf("SUPABASE_BUCKET: %s\n", bucket)
	fmt.Printf("Comprimento da chave: %d caracteres\n", len(key))

	if url == "" {
		fmt.Println("❌ ERRO: SUPABASE_URL está vazia!")
	}
	if key == "" {
		fmt.Println("❌ ERRO: SUPABASE_SERVICE_ROLE_KEY está vazia!")
	}
	if bucket == "" {
		fmt.Println("❌ ERRO: SUPABASE_BUCKET está vazia!")
	}

	if url != "" && key != "" && bucket != "" {
		fmt.Println("✅ Todas as variáveis obrigatórias estão presentes!")
	}
	fmt.Println("=====================================")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
