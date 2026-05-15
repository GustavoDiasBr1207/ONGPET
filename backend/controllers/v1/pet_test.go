package v1_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"ongpet/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────
// helpers locais
// ─────────────────────────────────────────────────────────────

func setupPetRouter(t *testing.T) (*models.Ong, func()) {
	t.Helper()
	mockData() // banco limpo + testOng inserida

	// registra rotas de pet no mesmo router
	return &testOng, func() {}
}

// ─────────────────────────────────────────────────────────────
// TestCriarPet
// ─────────────────────────────────────────────────────────────

func TestCriarPet(t *testing.T) {
	ong, _ := setupPetRouter(t)

	sucessoBody := map[string]any{
		"nome":      "Rex",
		"especie":   "Cachorro",
		"raca":      "Vira-lata",
		"idade":     2,
		"descricao": "Cachorro alegre",
		"peso":      8.5,
		"porte":     "Médio",
		"regiao":    "Serra",
		"status":    "available",
		"ong_id":    ong.ID,
	}

	testCases := []TestCase{
		{
			Description:    "Sucesso ao criar pet",
			URL:            "/api/v1/pets",
			Method:         http.MethodPost,
			Body:           mustMarshal(t, sucessoBody),
			ExpectedStatus: http.StatusCreated,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Pet criado com sucesso", resp["message"])
				assert.NotNil(t, resp["pet"])
			},
		},
		{
			Description: "Erro - ong_id ausente (zero UUID)",
			URL:         "/api/v1/pets",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"nome":    "Sem ONG",
				"especie": "Gato",
				"ong_id":  uuid.Nil,
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ong_id é obrigatório",
		},
		{
			Description: "Erro - ONG não encontrada",
			URL:         "/api/v1/pets",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"nome":    "Fantasma",
				"especie": "Gato",
				"ong_id":  uuid.New(),
			}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "ONG não encontrada",
		},
		{
			Description:    "Erro - body inválido (JSON malformado)",
			URL:            "/api/v1/pets",
			Method:         http.MethodPost,
			Body:           []byte(`{invalid}`),
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestListarPets
// ─────────────────────────────────────────────────────────────

func TestListarPets(t *testing.T) {
	ong, _ := setupPetRouter(t)

	pet1 := models.Pet{
		Nome:    "Bolinha",
		Especie: "Cachorro",
		Porte:   "Pequeno",
		Regiao:  "Serra",
		Status:  models.PetAvailable,
		OngID:   ong.ID,
	}
	pet2 := models.Pet{
		Nome:    "Mingau",
		Especie: "Gato",
		Porte:   "Pequeno",
		Regiao:  "Litoral",
		Status:  models.PetAvailable,
		OngID:   ong.ID,
	}
	check(t, mockDB.Create(&pet1))
	check(t, mockDB.Create(&pet2))

	testCases := []TestCase{
		{
			Description:    "Sucesso ao listar pets - primeira página",
			URL:            "/api/v1/pets?page=1&limit=10",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp PetListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.GreaterOrEqual(t, len(resp.Dados), 2)
				assert.False(t, resp.ProximaPagina)
			},
		},
		{
			Description:    "Sucesso com paginação - limit 1",
			URL:            "/api/v1/pets?page=1&limit=1",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp PetListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Len(t, resp.Dados, 1)
				assert.True(t, resp.ProximaPagina)
			},
		},
		{
			Description:    "Sucesso ao filtrar por nome",
			URL:            "/api/v1/pets?nome=Bolinha",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp PetListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Len(t, resp.Dados, 1)
				assert.Equal(t, "Bolinha", resp.Dados[0].Nome)
			},
		},
		{
			Description:    "Sucesso ao filtrar por especie",
			URL:            "/api/v1/pets?especie=Gato",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp PetListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Len(t, resp.Dados, 1)
				assert.Equal(t, "Mingau", resp.Dados[0].Nome)
			},
		},
		{
			Description:    "Sucesso ao filtrar por ong_id",
			URL:            fmt.Sprintf("/api/v1/pets?ong_id=%s", ong.ID),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp PetListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.GreaterOrEqual(t, len(resp.Dados), 2)
			},
		},
		{
			Description:    "Sucesso - filtro sem resultados retorna lista vazia",
			URL:            "/api/v1/pets?nome=NomeQueNaoExiste",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp PetListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Len(t, resp.Dados, 0)
			},
		},
		{
			Description:      "Erro - page inválido",
			URL:              "/api/v1/pets?page=abc",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "page inválido",
		},
		{
			Description:      "Erro - limit inválido",
			URL:              "/api/v1/pets?limit=xyz",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "limit inválido",
		},
		{
			Description:    "Erro - page zero",
			URL:            "/api/v1/pets?page=0",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Description:    "Erro - limit zero",
			URL:            "/api/v1/pets?limit=0",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestBuscarPetPorID
// ─────────────────────────────────────────────────────────────

func TestBuscarPetPorID(t *testing.T) {
	ong, _ := setupPetRouter(t)

	pet := models.Pet{
		Nome:    "Toby",
		Especie: "Cachorro",
		Status:  models.PetAvailable,
		OngID:   ong.ID,
	}
	check(t, mockDB.Create(&pet))

	testCases := []TestCase{
		{
			Description:    "Sucesso ao buscar pet por ID",
			URL:            fmt.Sprintf("/api/v1/pets/%s", pet.ID),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp models.Pet
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, pet.ID, resp.ID)
				assert.Equal(t, "Toby", resp.Nome)
			},
		},
		{
			Description:      "Erro - pet não encontrado",
			URL:              fmt.Sprintf("/api/v1/pets/%s", uuid.New()),
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Pet não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/pets/id-invalido",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do pet inválido",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestAtualizarPet
// ─────────────────────────────────────────────────────────────

func TestAtualizarPet(t *testing.T) {
	ong, _ := setupPetRouter(t)

	pet := models.Pet{
		Nome:    "Original",
		Especie: "Cachorro",
		Porte:   "Pequeno",
		Regiao:  "Serra",
		Status:  models.PetDraft,
		OngID:   ong.ID,
	}
	check(t, mockDB.Create(&pet))

	testCases := []TestCase{
		{
			Description: "Sucesso ao atualizar pet",
			URL:         fmt.Sprintf("/api/v1/pets/%s", pet.ID),
			Method:      http.MethodPut,
			Body: mustMarshal(t, map[string]any{
				"nome":      "Atualizado",
				"especie":   "Gato",
				"porte":     "Grande",
				"regiao":    "Vitória",
				"status":    "available",
				"descricao": "Nova descrição",
				"peso":      5.0,
				"idade":     3,
			}),
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Pet atualizado com sucesso", resp["message"])

				var updated models.Pet
				assert.NoError(t, mockDB.First(&updated, "id = ?", pet.ID).Error)
				assert.Equal(t, "Atualizado", updated.Nome)
				assert.Equal(t, models.PetEspecieGato, updated.Especie)
				assert.Equal(t, models.PetPorteGrande, updated.Porte)
				assert.Equal(t, models.PetRegiaoVitoria, updated.Regiao)
				assert.Equal(t, models.PetAvailable, updated.Status)
			},
		},
		{
			Description: "Sucesso - atualização parcial (apenas nome)",
			URL:         fmt.Sprintf("/api/v1/pets/%s", pet.ID),
			Method:      http.MethodPut,
			Body: mustMarshal(t, map[string]any{
				"nome": "Só Nome",
			}),
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Pet atualizado com sucesso", resp["message"])

				var updated models.Pet
				assert.NoError(t, mockDB.First(&updated, "id = ?", pet.ID).Error)
				assert.Equal(t, "Só Nome", updated.Nome)
			},
		},
		{
			Description:      "Erro - pet não encontrado",
			URL:              fmt.Sprintf("/api/v1/pets/%s", uuid.New()),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"nome": "X"}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Pet não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/pets/id-invalido",
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"nome": "X"}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do pet inválido",
		},
		{
			Description:    "Erro - body inválido (JSON malformado)",
			URL:            fmt.Sprintf("/api/v1/pets/%s", pet.ID),
			Method:         http.MethodPut,
			Body:           []byte(`{invalid}`),
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestDeletarPet
// ─────────────────────────────────────────────────────────────

func TestDeletarPet(t *testing.T) {
	ong, _ := setupPetRouter(t)

	pet := models.Pet{
		Nome:   "Para Deletar",
		OngID:  ong.ID,
		Status: models.PetDraft,
	}
	check(t, mockDB.Create(&pet))

	testCases := []TestCase{
		{
			Description:    "Sucesso ao deletar pet",
			URL:            fmt.Sprintf("/api/v1/pets/%s", pet.ID),
			Method:         http.MethodDelete,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Pet removido com sucesso", resp["message"])

				var deleted models.Pet
				err := mockDB.First(&deleted, "id = ?", pet.ID).Error
				assert.Error(t, err, "pet deveria ter sido deletado")
			},
		},
		{
			Description:      "Erro - pet não encontrado",
			URL:              fmt.Sprintf("/api/v1/pets/%s", uuid.New()),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Pet não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/pets/id-invalido",
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do pet inválido",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestUploadImagensPet
// ─────────────────────────────────────────────────────────────

func TestUploadImagensPet(t *testing.T) {
	ong, _ := setupPetRouter(t)

	pet := models.Pet{
		Nome:   "Pet Upload",
		OngID:  ong.ID,
		Status: models.PetDraft,
	}
	check(t, mockDB.Create(&pet))

	testCases := []TestCase{
		{
			Description:      "Erro - ID do pet inválido",
			URL:              "/api/v1/pets/id-invalido/imagens",
			Method:           http.MethodPost,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do pet inválido",
		},
		{
			Description:      "Erro - pet não encontrado",
			URL:              fmt.Sprintf("/api/v1/pets/%s/imagens", uuid.New()),
			Method:           http.MethodPost,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Pet não encontrado",
		},
		{
			Description:      "Erro - nenhuma imagem enviada",
			URL:              fmt.Sprintf("/api/v1/pets/%s/imagens", pet.ID),
			Method:           http.MethodPost,
			Body:             []byte(``),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "nenhuma imagem enviada",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestDeletarImagemPet
// ─────────────────────────────────────────────────────────────

func TestDeletarImagemPet(t *testing.T) {
	ong, _ := setupPetRouter(t)

	pet := models.Pet{
		Nome:   "Pet Com Imagem",
		OngID:  ong.ID,
		Status: models.PetAvailable,
	}
	check(t, mockDB.Create(&pet))

	imagem := models.PetImage{
		URL:      "https://storage.example.com/pets/imagem.webp",
		PetID:    pet.ID,
		Position: 1,
	}
	check(t, mockDB.Create(&imagem))

	testCases := []TestCase{
		{
			Description:      "Erro - ID do pet inválido",
			URL:              fmt.Sprintf("/api/v1/pets/id-invalido/imagens/%s", imagem.ID),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do pet inválido",
		},
		{
			Description:      "Erro - ID da imagem inválido",
			URL:              fmt.Sprintf("/api/v1/pets/%s/imagens/id-invalido", pet.ID),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID da imagem inválido",
		},
		{
			Description:      "Erro - imagem não encontrada",
			URL:              fmt.Sprintf("/api/v1/pets/%s/imagens/%s", pet.ID, uuid.New()),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Imagem não encontrada",
		},
		{
			Description:    "Sucesso ao remover imagem do pet",
			URL:            fmt.Sprintf("/api/v1/pets/%s/imagens/%s", pet.ID, imagem.ID),
			Method:         http.MethodDelete,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Imagem removida com sucesso", resp["message"])

				var deletada models.PetImage
				err := mockDB.First(&deletada, "id = ?", imagem.ID).Error
				assert.Error(t, err, "imagem deveria ter sido deletada")
			},
		},
	}

	runTestCases(t, testCases)
}