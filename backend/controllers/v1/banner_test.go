package v1_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ongpet/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────
// helpers locais
// ─────────────────────────────────────────────────────────────

func criarBanner(t *testing.T, title string) models.Banner {
	t.Helper()
	b := models.Banner{
		Title:        title,
		InstagramURL: "https://instagram.com/ongpet",
		StartAt:      time.Now().Add(-24 * time.Hour),
		EndAt:        time.Now().Add(30 * 24 * time.Hour),
		Active:       true,
		Position:     0,
	}
	check(t, mockDB.Create(&b))
	return b
}

// ─────────────────────────────────────────────────────────────
// TestReadBanners
// ─────────────────────────────────────────────────────────────

func TestReadBanners(t *testing.T) {
	mockData()
	criarBanner(t, "Banner A")
	criarBanner(t, "Banner B")

	testCases := []TestCase{
		{
			Description:    "Sucesso ao listar banners",
			URL:            "/api/v1/banners",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp BannerListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.GreaterOrEqual(t, len(resp.Dados), 2)
			},
		},
		{
			Description:    "Filtro active=true retorna apenas ativos",
			URL:            "/api/v1/banners?active=true",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp BannerListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				for _, b := range resp.Dados {
					assert.True(t, b.Active)
				}
			},
		},
		{
			Description:    "Filtro vigente=true retorna apenas banners no período",
			URL:            "/api/v1/banners?vigente=true",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp BannerListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				now := time.Now()
				for _, b := range resp.Dados {
					assert.False(t, b.StartAt.After(now))
					assert.False(t, b.EndAt.Before(now))
				}
			},
		},
		{
			Description:    "Paginação - limit=1 retorna no máximo 1 item",
			URL:            "/api/v1/banners?page=1&limit=1",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp BannerListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.LessOrEqual(t, len(resp.Dados), 1)
				assert.True(t, resp.ProximaPagina)
			},
		},
		{
			Description:      "Erro - page inválido",
			URL:              "/api/v1/banners?page=abc",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "page inválido",
		},
		{
			Description:      "Erro - limit inválido",
			URL:              "/api/v1/banners?limit=abc",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "limit inválido",
		},
		{
			Description:      "Erro - limit acima do máximo (100)",
			URL:              "/api/v1/banners?limit=101",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "limit excede o máximo permitido de 100",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestReadBanner
// ─────────────────────────────────────────────────────────────

func TestReadBanner(t *testing.T) {
	mockData()
	banner := criarBanner(t, "Banner Único")

	testCases := []TestCase{
		{
			Description:    "Sucesso ao buscar banner por ID",
			URL:            fmt.Sprintf("/api/v1/banners/%s", banner.ID),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp models.Banner
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, banner.ID, resp.ID)
				assert.Equal(t, "Banner Único", resp.Title)
			},
		},
		{
			Description:      "Erro - Banner não encontrado",
			URL:              fmt.Sprintf("/api/v1/banners/%s", uuid.New()),
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Banner não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/banners/id-invalido",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do banner inválido",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestCreateBanner
// ─────────────────────────────────────────────────────────────

func TestCreateBanner(t *testing.T) {
	mockData()

	validDates := map[string]any{
		"start_at": time.Now().Add(-time.Hour).Format(time.RFC3339),
		"end_at":   time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
	}

	sucessoBody := map[string]any{
		"title":         "Banner Novo",
		"instagram_url": "https://instagram.com/ongpet",
		"start_at":      validDates["start_at"],
		"end_at":        validDates["end_at"],
		"active":        true,
		"position":      1,
		"ong_id":        testOng.ID,
	}

	testCases := []TestCase{
		{
			Description:    "Sucesso ao criar banner",
			URL:            "/api/v1/banners",
			Method:         http.MethodPost,
			Body:           mustMarshal(t, sucessoBody),
			ExpectedStatus: http.StatusCreated,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Banner criado com sucesso", resp["message"])
				b, ok := resp["banner"].(map[string]any)
				assert.True(t, ok)
				assert.Equal(t, "Banner Novo", b["title"])
			},
		},
		{
			Description: "Erro - title ausente",
			URL:         "/api/v1/banners",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"instagram_url": "https://instagram.com/ongpet",
				"start_at":      validDates["start_at"],
				"end_at":        validDates["end_at"],
				"ong_id":        testOng.ID,
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "title é obrigatório",
		},
		{
			Description: "Erro - instagram_url ausente",
			URL:         "/api/v1/banners",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"title":   "Banner X",
				"start_at": validDates["start_at"],
				"end_at":  validDates["end_at"],
				"ong_id":  testOng.ID,
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "instagram_url é obrigatório",
		},
		{
			Description: "Erro - start_at ausente",
			URL:         "/api/v1/banners",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"title":         "Banner X",
				"instagram_url": "https://instagram.com/ongpet",
				"end_at":        validDates["end_at"],
				"ong_id":        testOng.ID,
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "start_at é obrigatório",
		},
		{
			Description: "Erro - end_at ausente",
			URL:         "/api/v1/banners",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"title":         "Banner X",
				"instagram_url": "https://instagram.com/ongpet",
				"start_at":      validDates["start_at"],
				"ong_id":        testOng.ID,
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "end_at é obrigatório",
		},
		{
			Description: "Erro - ong_id ausente",
			URL:         "/api/v1/banners",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"title":         "Banner X",
				"instagram_url": "https://instagram.com/ongpet",
				"start_at":      validDates["start_at"],
				"end_at":        validDates["end_at"],
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ong_id é obrigatório",
		},
		{
			Description: "Erro - start_at formato inválido",
			URL:         "/api/v1/banners",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"title":         "Banner X",
				"instagram_url": "https://instagram.com/ongpet",
				"start_at":      "01/01/2026",
				"end_at":        validDates["end_at"],
				"ong_id":        testOng.ID,
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "start_at deve estar no formato RFC3339",
		},
		{
			Description: "Erro - end_at formato inválido",
			URL:         "/api/v1/banners",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"title":         "Banner X",
				"instagram_url": "https://instagram.com/ongpet",
				"start_at":      validDates["start_at"],
				"end_at":        "01/01/2026",
				"ong_id":        testOng.ID,
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "end_at deve estar no formato RFC3339",
		},
		{
			Description:    "Erro - body inválido (JSON malformado)",
			URL:            "/api/v1/banners",
			Method:         http.MethodPost,
			Body:           []byte(`{invalid}`),
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Description: "Erro - end_at anterior a start_at",
			URL:         "/api/v1/banners",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"title":         "Banner X",
				"instagram_url": "https://instagram.com/ongpet",
				"start_at":      time.Now().Add(48 * time.Hour).Format(time.RFC3339),
				"end_at":        time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"ong_id":        testOng.ID,
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "end_at deve ser posterior a start_at",
		},
		{
			Description: "Erro - position negativo",
			URL:         "/api/v1/banners",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"title":         "Banner X",
				"instagram_url": "https://instagram.com/ongpet",
				"start_at":      validDates["start_at"],
				"end_at":        validDates["end_at"],
				"ong_id":        testOng.ID,
				"position":      -1,
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "position não pode ser negativo",
		},
		{
			Description: "Erro - ONG não encontrada",
			URL:         "/api/v1/banners",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"title":         "Banner X",
				"instagram_url": "https://instagram.com/ongpet",
				"start_at":      validDates["start_at"],
				"end_at":        validDates["end_at"],
				"ong_id":        "00000000-0000-0000-0000-000000000099",
			}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "ONG não encontrada",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestUpdateBanner
// ─────────────────────────────────────────────────────────────

func TestUpdateBanner(t *testing.T) {
	mockData()
	banner := criarBanner(t, "Banner Original")

	testCases := []TestCase{
		{
			Description:    "Sucesso ao atualizar título",
			URL:            fmt.Sprintf("/api/v1/banners/%s", banner.ID),
			Method:         http.MethodPut,
			Body:           mustMarshal(t, map[string]any{"title": "Banner Atualizado"}),
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Banner atualizado com sucesso", resp["message"])
				b, ok := resp["banner"].(map[string]any)
				assert.True(t, ok)
				assert.Equal(t, "Banner Atualizado", b["title"])
			},
		},
		{
			Description:    "Sucesso ao atualizar active e position",
			URL:            fmt.Sprintf("/api/v1/banners/%s", banner.ID),
			Method:         http.MethodPut,
			Body:           mustMarshal(t, map[string]any{"active": false, "position": 5}),
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				b, ok := resp["banner"].(map[string]any)
				assert.True(t, ok)
				assert.Equal(t, false, b["active"])
				assert.Equal(t, float64(5), b["position"])
			},
		},
		{
			Description:      "Erro - start_at formato inválido",
			URL:              fmt.Sprintf("/api/v1/banners/%s", banner.ID),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"start_at": "nao-e-data"}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "start_at deve estar no formato RFC3339",
		},
		{
			Description:      "Erro - end_at formato inválido",
			URL:              fmt.Sprintf("/api/v1/banners/%s", banner.ID),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"end_at": "nao-e-data"}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "end_at deve estar no formato RFC3339",
		},
		{
			Description:      "Erro - Banner não encontrado",
			URL:              fmt.Sprintf("/api/v1/banners/%s", uuid.New()),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"title": "X"}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Banner não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/banners/id-invalido",
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"title": "X"}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do banner inválido",
		},
		{
			Description:    "Erro - body inválido (JSON malformado)",
			URL:            fmt.Sprintf("/api/v1/banners/%s", banner.ID),
			Method:         http.MethodPut,
			Body:           []byte(`{invalid}`),
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Description:      "Erro - title vazio",
			URL:              fmt.Sprintf("/api/v1/banners/%s", banner.ID),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"title": ""}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "title não pode ser vazio",
		},
		{
			Description:      "Erro - instagram_url vazia",
			URL:              fmt.Sprintf("/api/v1/banners/%s", banner.ID),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"instagram_url": ""}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "instagram_url não pode ser vazia",
		},
		{
			Description:      "Erro - position negativo",
			URL:              fmt.Sprintf("/api/v1/banners/%s", banner.ID),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"position": -3}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "position não pode ser negativo",
		},
		{
			Description: "Erro - end_at anterior a start_at",
			URL:         fmt.Sprintf("/api/v1/banners/%s", banner.ID),
			Method:      http.MethodPut,
			Body: mustMarshal(t, map[string]any{
				"start_at": time.Now().Add(48 * time.Hour).Format(time.RFC3339),
				"end_at":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "end_at deve ser posterior a start_at",
		},
		{
			Description:      "Erro - ONG não encontrada ao atualizar ong_id",
			URL:              fmt.Sprintf("/api/v1/banners/%s", banner.ID),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"ong_id": "00000000-0000-0000-0000-000000000099"}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "ONG não encontrada",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestDeleteBanner
// ─────────────────────────────────────────────────────────────

func TestDeleteBanner(t *testing.T) {
	mockData()
	banner := criarBanner(t, "Para Deletar")

	testCases := []TestCase{
		{
			Description:    "Sucesso ao deletar banner",
			URL:            fmt.Sprintf("/api/v1/banners/%s", banner.ID),
			Method:         http.MethodDelete,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Banner removido com sucesso", resp["message"])
			},
		},
		{
			Description:      "Erro - Banner não encontrado",
			URL:              fmt.Sprintf("/api/v1/banners/%s", uuid.New()),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Banner não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/banners/id-invalido",
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do banner inválido",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestDeleteBannerImage
// ─────────────────────────────────────────────────────────────

func TestDeleteBannerImage(t *testing.T) {
	mockData()

	bannerSemImagem := criarBanner(t, "Sem Imagem")

	bannerComImagem := criarBanner(t, "Com Imagem")
	check(t, mockDB.Model(&bannerComImagem).Update("image_url", "https://storage.example.com/banners/capa.webp"))

	testCases := []TestCase{
		{
			Description:      "Erro - banner sem imagem",
			URL:              fmt.Sprintf("/api/v1/banners/%s/imagem", bannerSemImagem.ID),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "este banner não possui imagem",
		},
		{
			Description:    "Sucesso ao deletar imagem do banner",
			URL:            fmt.Sprintf("/api/v1/banners/%s/imagem", bannerComImagem.ID),
			Method:         http.MethodDelete,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Imagem removida com sucesso", resp["message"])
				b, ok := resp["banner"].(map[string]any)
				assert.True(t, ok)
				assert.Equal(t, "", b["image_url"])
			},
		},
		{
			Description:      "Erro - Banner não encontrado",
			URL:              fmt.Sprintf("/api/v1/banners/%s/imagem", uuid.New()),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Banner não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/banners/id-invalido/imagem",
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do banner inválido",
		},
	}

	runTestCases(t, testCases)
}