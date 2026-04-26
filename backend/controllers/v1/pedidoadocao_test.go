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

func criarPet(t *testing.T, ongID uuid.UUID, nome string) models.Pet {
	t.Helper()
	pet := models.Pet{
		Nome:    nome,
		Especie: "Cachorro",
		Status:  models.PetAvailable,
		OngID:   ongID,
	}
	check(t, mockDB.Create(&pet))
	return pet
}

func criarPedido(t *testing.T, ongID uuid.UUID, petID uuid.UUID) models.PedidoAdocao {
	t.Helper()
	pedido := models.PedidoAdocao{
		OngID:  ongID,
		PetID:  petID,
		Status: models.PedidoAdocaoPendente,
	}
	check(t, mockDB.Create(&pedido))
	return pedido
}

// ─────────────────────────────────────────────────────────────
// TestCriarPedidoAdocao
// ─────────────────────────────────────────────────────────────

func TestCriarPedidoAdocao(t *testing.T) {
	mockData()

	pet := criarPet(t, testOng.ID, "Bolinha")

	testCases := []TestCase{
		{
			Description: "Sucesso ao criar pedido de adoção",
			URL:         "/api/v1/pedidos-adocao",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"ong_id":    testOng.ID,
				"pet_id":    pet.ID,
				"respostas": []any{},
			}),
			ExpectedStatus: http.StatusCreated,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Pedido de adoção criado com sucesso", resp["message"])
				assert.NotNil(t, resp["pedido"])
			},
		},
		{
			Description: "Erro - ong_id ausente (zero UUID)",
			URL:         "/api/v1/pedidos-adocao",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"ong_id": uuid.Nil,
				"pet_id": pet.ID,
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ong_id é obrigatório",
		},
		{
			Description: "Erro - pet_id ausente (zero UUID)",
			URL:         "/api/v1/pedidos-adocao",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"ong_id": testOng.ID,
				"pet_id": uuid.Nil,
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "pet_id é obrigatório",
		},
		{
			Description: "Erro - ONG não encontrada",
			URL:         "/api/v1/pedidos-adocao",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"ong_id": uuid.New(),
				"pet_id": pet.ID,
			}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "ONG não encontrada",
		},
		{
			Description: "Erro - Pet não encontrado",
			URL:         "/api/v1/pedidos-adocao",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"ong_id": testOng.ID,
				"pet_id": uuid.New(),
			}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Pet não encontrado",
		},
		{
			Description:    "Erro - body inválido (JSON malformado)",
			URL:            "/api/v1/pedidos-adocao",
			Method:         http.MethodPost,
			Body:           []byte(`{invalid}`),
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestListarPedidosAdocao
// ─────────────────────────────────────────────────────────────

func TestListarPedidosAdocao(t *testing.T) {
	mockData()

	pet := criarPet(t, testOng.ID, "Bolinha")
	criarPedido(t, testOng.ID, pet.ID)
	criarPedido(t, testOng.ID, pet.ID)

	testCases := []TestCase{
		{
			Description:    "Sucesso ao listar pedidos - primeira página",
			URL:            "/api/v1/pedidos-adocao?page=1&limit=10",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp PedidoAdocaoListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.GreaterOrEqual(t, len(resp.Dados), 2)
				assert.False(t, resp.ProximaPagina)
			},
		},
		{
			Description:    "Sucesso com paginação - limit 1",
			URL:            "/api/v1/pedidos-adocao?page=1&limit=1",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp PedidoAdocaoListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Len(t, resp.Dados, 1)
				assert.True(t, resp.ProximaPagina)
			},
		},
		{
			Description:    "Sucesso ao filtrar por ong_id",
			URL:            fmt.Sprintf("/api/v1/pedidos-adocao?ong_id=%s", testOng.ID),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp PedidoAdocaoListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.GreaterOrEqual(t, len(resp.Dados), 2)
			},
		},
		{
			Description:    "Sucesso ao filtrar por pet_id",
			URL:            fmt.Sprintf("/api/v1/pedidos-adocao?pet_id=%s", pet.ID),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp PedidoAdocaoListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.GreaterOrEqual(t, len(resp.Dados), 2)
			},
		},
		{
			Description:    "Sucesso - filtro por ong inexistente retorna lista vazia",
			URL:            fmt.Sprintf("/api/v1/pedidos-adocao?ong_id=%s", uuid.New()),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp PedidoAdocaoListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Len(t, resp.Dados, 0)
			},
		},
		{
			Description:      "Erro - page inválido",
			URL:              "/api/v1/pedidos-adocao?page=abc",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "page inválido",
		},
		{
			Description:      "Erro - limit inválido",
			URL:              "/api/v1/pedidos-adocao?limit=xyz",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "limit inválido",
		},
		{
			Description:    "Erro - page zero",
			URL:            "/api/v1/pedidos-adocao?page=0",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Description:    "Erro - limit zero",
			URL:            "/api/v1/pedidos-adocao?limit=0",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestBuscarPedidoAdocaoPorID
// ─────────────────────────────────────────────────────────────

func TestBuscarPedidoAdocaoPorID(t *testing.T) {
	mockData()

	pet := criarPet(t, testOng.ID, "Toby")
	pedido := criarPedido(t, testOng.ID, pet.ID)

	testCases := []TestCase{
		{
			Description:    "Sucesso ao buscar pedido por ID",
			URL:            fmt.Sprintf("/api/v1/pedidos-adocao/%s", pedido.ID),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp models.PedidoAdocaoDTO
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, pedido.ID, resp.ID)
				assert.Equal(t, models.PedidoAdocaoPendente, resp.Status)
			},
		},
		{
			Description:      "Erro - pedido não encontrado",
			URL:              fmt.Sprintf("/api/v1/pedidos-adocao/%s", uuid.New()),
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Pedido de adoção não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/pedidos-adocao/id-invalido",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do pedido inválido",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestDeletarPedidoAdocao
// ─────────────────────────────────────────────────────────────

func TestDeletarPedidoAdocao(t *testing.T) {
	mockData()

	pet := criarPet(t, testOng.ID, "Para Deletar")
	pedido := criarPedido(t, testOng.ID, pet.ID)

	testCases := []TestCase{
		{
			Description:    "Sucesso ao deletar pedido de adoção",
			URL:            fmt.Sprintf("/api/v1/pedidos-adocao/%s", pedido.ID),
			Method:         http.MethodDelete,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Pedido de adoção removido com sucesso", resp["message"])

				var deleted models.PedidoAdocao
				err := mockDB.First(&deleted, "id = ?", pedido.ID).Error
				assert.Error(t, err, "pedido deveria ter sido deletado")
			},
		},
		{
			Description:      "Erro - pedido não encontrado",
			URL:              fmt.Sprintf("/api/v1/pedidos-adocao/%s", uuid.New()),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Pedido de adoção não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/pedidos-adocao/id-invalido",
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do pedido inválido",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestAtualizarStatusPedidoAdocao
// ─────────────────────────────────────────────────────────────

func TestAtualizarStatusPedidoAdocao(t *testing.T) {
	mockData()

	pet := criarPet(t, testOng.ID, "Rex")
	pedido := criarPedido(t, testOng.ID, pet.ID)

	testCases := []TestCase{
		{
			Description:    "Sucesso ao aprovar pedido",
			URL:            fmt.Sprintf("/api/v1/pedidos-adocao/%s/status", pedido.ID),
			Method:         http.MethodPut,
			Body:           mustMarshal(t, map[string]any{"status": "aprovado"}),
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Status atualizado com sucesso", resp["message"])

				var updated models.PedidoAdocao
				assert.NoError(t, mockDB.First(&updated, "id = ?", pedido.ID).Error)
				assert.Equal(t, models.PedidoAdocaoAprovado, updated.Status)
			},
		},
		{
			Description:    "Sucesso ao rejeitar pedido",
			URL:            fmt.Sprintf("/api/v1/pedidos-adocao/%s/status", pedido.ID),
			Method:         http.MethodPut,
			Body:           mustMarshal(t, map[string]any{"status": "rejeitado"}),
			ExpectedStatus: http.StatusOK,
		},
		{
			Description:    "Sucesso ao cancelar pedido",
			URL:            fmt.Sprintf("/api/v1/pedidos-adocao/%s/status", pedido.ID),
			Method:         http.MethodPut,
			Body:           mustMarshal(t, map[string]any{"status": "cancelado"}),
			ExpectedStatus: http.StatusOK,
		},
		{
			Description:    "Sucesso ao marcar como pendente",
			URL:            fmt.Sprintf("/api/v1/pedidos-adocao/%s/status", pedido.ID),
			Method:         http.MethodPut,
			Body:           mustMarshal(t, map[string]any{"status": "pendente"}),
			ExpectedStatus: http.StatusOK,
		},
		{
			Description:      "Erro - status inválido",
			URL:              fmt.Sprintf("/api/v1/pedidos-adocao/%s/status", pedido.ID),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"status": "em_andamento"}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "status inválido",
		},
		{
			Description:      "Erro - pedido não encontrado",
			URL:              fmt.Sprintf("/api/v1/pedidos-adocao/%s/status", uuid.New()),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"status": "aprovado"}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Pedido de adoção não encontrado",
		},
		{
			// body válido para não travar no ShouldBindJSON antes do uuid.Parse
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/pedidos-adocao/id-invalido/status",
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"status": "aprovado"}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do pedido inválido",
		},
		{
			Description:    "Erro - body inválido (JSON malformado)",
			URL:            fmt.Sprintf("/api/v1/pedidos-adocao/%s/status", pedido.ID),
			Method:         http.MethodPut,
			Body:           []byte(`{invalid}`),
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestUploadRespostaImagem
// ─────────────────────────────────────────────────────────────

func TestUploadRespostaImagem(t *testing.T) {
	mockData()

	pet := criarPet(t, testOng.ID, "Pet Upload")
	pedido := criarPedido(t, testOng.ID, pet.ID)

	testCases := []TestCase{
		{
			Description:      "Erro - ID do pedido inválido",
			URL:              fmt.Sprintf("/api/v1/pedidos-adocao/id-invalido/respostas/%s/imagem", uuid.New()),
			Method:           http.MethodPost,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do pedido inválido",
		},
		{
			Description:      "Erro - ID da resposta inválido",
			URL:              fmt.Sprintf("/api/v1/pedidos-adocao/%s/respostas/id-invalido/imagem", pedido.ID),
			Method:           http.MethodPost,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID da resposta inválido",
		},
		{
			Description:      "Erro - resposta não encontrada",
			URL:              fmt.Sprintf("/api/v1/pedidos-adocao/%s/respostas/%s/imagem", pedido.ID, uuid.New()),
			Method:           http.MethodPost,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Resposta não encontrada",
		},
	}

	runTestCases(t, testCases)
}