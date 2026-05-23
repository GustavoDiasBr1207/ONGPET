package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func debugEnv() {
	if os.Getenv("DEBUG") != "true" {
		return
	}

	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️ .env não encontrado")
	}

	url   := strings.TrimSpace(os.Getenv("SUPABASE_URL"))
	key   := strings.TrimSpace(os.Getenv("SUPABASE_KEY"))
	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))

	fmt.Println("\n=== DEBUG: VARIÁVEIS DE AMBIENTE ===")
	fmt.Printf("DATABASE_URL:        %s\n", truncate(dbURL, 40))
	fmt.Printf("SUPABASE_URL:        %s\n", url)
	fmt.Printf("SUPABASE_KEY:        %s...\n", truncate(key, 30))
	fmt.Printf("WHATSAPP_ENABLED:    %s\n", os.Getenv("WHATSAPP_ENABLED"))
	fmt.Printf("SWAGGER_ENABLED:     %s\n", os.Getenv("SWAGGER_ENABLED"))
	fmt.Printf("GIN_MODE:            %s\n", os.Getenv("GIN_MODE"))

	missing := false
	for _, v := range []struct{ name, val string }{
		{"DATABASE_URL", dbURL},
		{"SUPABASE_URL", url},
		{"SUPABASE_KEY", key},
	} {
		if v.val == "" {
			fmt.Printf("❌ ERRO: %s está vazia!\n", v.name)
			missing = true
		}
	}
	if !missing {
		fmt.Println("✅ Variáveis obrigatórias presentes!")
	}
	fmt.Println("=====================================")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}