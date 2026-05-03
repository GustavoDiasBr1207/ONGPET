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

// UpdateUserOngMetadata atualiza os user_metadata do usuário no Supabase Auth
// adicionando os dados da ONG criada. Usa a Admin API (service role key).
func UpdateUserOngMetadata(userID string, ong OngMetadata) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || serviceKey == "" {
		return fmt.Errorf("SUPABASE_URL ou SUPABASE_SERVICE_ROLE_KEY não configurados")
	}

	payload := map[string]interface{}{
		"user_metadata": map[string]interface{}{
			"ong_id":    ong.OngID,
			"ong_nome":  ong.OngNome,
			"ong_email": ong.OngEmail,
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar metadata: %w", err)
	}

	url := fmt.Sprintf("%s/auth/v1/admin/users/%s", supabaseURL, userID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("erro ao criar request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao chamar Supabase Admin API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		return fmt.Errorf("supabase retornou %d: %v", resp.StatusCode, errBody)
	}

	return nil
}