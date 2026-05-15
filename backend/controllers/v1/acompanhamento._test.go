package v1_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v1 "ongpet/controllers/v1"
	"ongpet/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────
// helpers locais
// ─────────────────────────────────────────────────────────────

func criarAcompanhamento(t *testing.T, petID uuid.UUID) models.Acompanhamento {
	t.Helper()
	proxima := time.Now().Add(30 * 24 * time.Hour)
	a := models.Acompanhamento{
		PetID:            petID,
		OngID:            testOng.ID,
		AdotanteNome:     "João Silva",
		AdotanteTelefone: "11999999999",
		AdotanteEmail:    "joao@email.com",
		Frequencia:       models.FrequencyMonthly,
		ProximaData:      &proxima,
		Status:           models.AcompanhamentoAtivo,
		LembreteEnviado:  false,
	}
	check(t, mockDB.Create(&a))
	return a
}

// ─────────────────────────────────────────────────────────────
// TestCriarAcompanhamento
// ─────────────────────────────────────────────────────────────

func TestCriarAcompanhamento(t *testing.T) {
	mockData()

	pet := criarPet(t, testOng.ID, "Buddy")
	proxima := time.Now().Add(30 * 24 * time.Hour)

	sucessoBody := map[string]any{
		"pet_id":            pet.ID,
		"adotante_nome":     "João Silva",
		"adotante_telefone": "11999999999",
		"adotante_email":    "joao@email.com",
		"frequencia":        "MONTHLY",
		"proxima_data":      proxima.Format(time.RFC3339),
	}

	testCases := []TestCase{
		{
			Description:    "Sucesso ao criar acompanhamento",
			URL:            "/api/v1/acompanhamentos",
			Method:         http.MethodPost,
			Body:           mustMarshal(t, sucessoBody),
			ExpectedStatus: http.StatusCreated,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp models.AcompanhamentoDTO
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, pet.ID, resp.PetID)
				assert.Equal(t, "João Silva", resp.AdotanteNome)
				assert.Equal(t, models.FrequencyMonthly, resp.Frequencia)
				assert.Equal(t, models.AcompanhamentoAtivo, resp.Status)
			},
		},
		{
			Description: "Erro - pet_id ausente",
			URL:         "/api/v1/acompanhamentos",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"adotante_nome":     "João",
				"adotante_telefone": "11999999999",
				"frequencia":        "MONTHLY",
			}),
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Description: "Erro - adotante_nome ausente",
			URL:         "/api/v1/acompanhamentos",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"pet_id":            pet.ID,
				"adotante_telefone": "11999999999",
				"frequencia":        "MONTHLY",
			}),
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Description: "Erro - adotante_telefone ausente",
			URL:         "/api/v1/acompanhamentos",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"pet_id":        pet.ID,
				"adotante_nome": "João",
				"frequencia":    "MONTHLY",
			}),
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Description: "Erro - frequencia ausente",
			URL:         "/api/v1/acompanhamentos",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"pet_id":            pet.ID,
				"adotante_nome":     "João",
				"adotante_telefone": "11999999999",
			}),
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Description: "Erro - Pet não encontrado",
			URL:         "/api/v1/acompanhamentos",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"pet_id":            uuid.New(),
				"adotante_nome":     "João",
				"adotante_telefone": "11999999999",
				"frequencia":        "MONTHLY",
			}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Pet não encontrado",
		},
		{
			Description:    "Erro - body inválido (JSON malformado)",
			URL:            "/api/v1/acompanhamentos",
			Method:         http.MethodPost,
			Body:           []byte(`{invalid}`),
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestListarAcompanhamentos
// ─────────────────────────────────────────────────────────────

func TestListarAcompanhamentos(t *testing.T) {
	mockData()

	pet1 := criarPet(t, testOng.ID, "Buddy")
	pet2 := criarPet(t, testOng.ID, "Rex")
	criarAcompanhamento(t, pet1.ID)
	criarAcompanhamento(t, pet2.ID)

	testCases := []TestCase{
		{
			Description:    "Sucesso ao listar acompanhamentos",
			URL:            "/api/v1/acompanhamentos",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp v1.AcompanhamentoListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.GreaterOrEqual(t, len(resp.Dados), 2)
			},
		},
		{
			Description:    "Sucesso - lista vazia quando não há acompanhamentos",
			URL:            "/api/v1/acompanhamentos",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp v1.AcompanhamentoListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			},
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestCriarLogAcompanhamento
// ─────────────────────────────────────────────────────────────

func TestCriarLogAcompanhamento(t *testing.T) {
	mockData()

	pet := criarPet(t, testOng.ID, "Mel")
	acomp := criarAcompanhamento(t, pet.ID)
	proxima := time.Now().Add(60 * 24 * time.Hour)

	testCases := []TestCase{
		{
			Description: "Sucesso ao criar log sem atualizar proxima_data",
			URL:         fmt.Sprintf("/api/v1/acompanhamentos/%s/logs", acomp.ID),
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"notas":       "Pet está saudável e feliz.",
				"data_contato": time.Now().Format(time.RFC3339),
			}),
			ExpectedStatus: http.StatusCreated,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp models.LogAcompanhamentoDTO
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, acomp.ID, resp.AcompanhamentoID)
				assert.Equal(t, "Pet está saudável e feliz.", resp.Notas)
			},
		},
		{
			Description: "Sucesso ao criar log atualizando proxima_data",
			URL:         fmt.Sprintf("/api/v1/acompanhamentos/%s/logs", acomp.ID),
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"notas":        "Contato realizado, tudo bem.",
				"proxima_data": proxima.Format(time.RFC3339),
			}),
			ExpectedStatus: http.StatusCreated,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp models.LogAcompanhamentoDTO
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Contato realizado, tudo bem.", resp.Notas)

				// verifica que proxima_data foi atualizada e lembrete_enviado resetado
				var updated models.Acompanhamento
				assert.NoError(t, mockDB.First(&updated, "id = ?", acomp.ID).Error)
				assert.False(t, updated.LembreteEnviado)
				assert.NotNil(t, updated.ProximaData)
			},
		},
		{
			Description: "Erro - notas ausente",
			URL:         fmt.Sprintf("/api/v1/acompanhamentos/%s/logs", acomp.ID),
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"data_contato": time.Now().Format(time.RFC3339),
			}),
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Description:      "Erro - ID de acompanhamento inválido",
			URL:              "/api/v1/acompanhamentos/id-invalido/logs",
			Method:           http.MethodPost,
			Body:             mustMarshal(t, map[string]any{"notas": "teste"}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID de acompanhamento inválido",
		},
		{
			Description:    "Erro - body inválido (JSON malformado)",
			URL:            fmt.Sprintf("/api/v1/acompanhamentos/%s/logs", acomp.ID),
			Method:         http.MethodPost,
			Body:           []byte(`{invalid}`),
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestBuscarLogsAcompanhamento
// ─────────────────────────────────────────────────────────────

func TestBuscarLogsAcompanhamento(t *testing.T) {
	mockData()

	pet := criarPet(t, testOng.ID, "Thor")
	acomp := criarAcompanhamento(t, pet.ID)

	log1 := models.LogAcompanhamento{
		AcompanhamentoID: acomp.ID,
		DataContato:      time.Now().Add(-48 * time.Hour),
		Notas:            "Primeira visita",
	}
	log2 := models.LogAcompanhamento{
		AcompanhamentoID: acomp.ID,
		DataContato:      time.Now(),
		Notas:            "Segunda visita",
	}
	check(t, mockDB.Create(&log1))
	check(t, mockDB.Create(&log2))

	testCases := []TestCase{
		{
			Description:    "Sucesso ao buscar logs do acompanhamento",
			URL:            fmt.Sprintf("/api/v1/acompanhamentos/%s/logs", acomp.ID),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp []models.LogAcompanhamentoDTO
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Len(t, resp, 2)
			},
		},
		{
			Description:    "Sucesso - acompanhamento sem logs retorna lista vazia",
			URL:            fmt.Sprintf("/api/v1/acompanhamentos/%s/logs", uuid.New()),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp []models.LogAcompanhamentoDTO
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Len(t, resp, 0)
			},
		},
		{
			Description:      "Erro - ID de acompanhamento inválido",
			URL:              "/api/v1/acompanhamentos/id-invalido/logs",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID de acompanhamento inválido",
		},
	}

	runTestCases(t, testCases)
}