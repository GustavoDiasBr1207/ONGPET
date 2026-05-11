package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type OngMetadata struct {
	OngID    string `json:"ong_id"`
	OngNome  string `json:"ong_nome"`
	OngEmail string `json:"ong_email"`
}

// UpdateUserOngMetadataFn permite substituir a implementação em testes.
var UpdateUserOngMetadataFn func(userID string, ong OngMetadata, userJWT string) error = updateUserOngMetadataImpl

// UpdateUserOngMetadata é o ponto de entrada público — usa a Fn injetável.
func UpdateUserOngMetadata(userID string, ong OngMetadata, userJWT string) error {
	return UpdateUserOngMetadataFn(userID, ong, userJWT)
}

// updateUserOngMetadataImpl é a implementação real usando o JWT do usuário.
func updateUserOngMetadataImpl(userID string, ong OngMetadata, userJWT string) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_KEY")

	if supabaseURL == "" || anonKey == "" {
		return fmt.Errorf("SUPABASE_URL ou SUPABASE_KEY não configurados")
	}

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"ong_id":    ong.OngID,
			"ong_nome":  ong.OngNome,
			"ong_email": ong.OngEmail,
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar metadata: %w", err)
	}

	url := fmt.Sprintf("%s/auth/v1/user", supabaseURL)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("erro ao criar request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", anonKey)
	req.Header.Set("Authorization", "Bearer "+userJWT)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao chamar Supabase Auth API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("supabase retornou %d: %v", resp.StatusCode, errBody)
	}

	return nil
}