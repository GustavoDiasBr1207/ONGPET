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

func criarFormulario(t *testing.T, ongID uuid.UUID, nome string) models.FormularioModelo {
	t.Helper()
	f := models.FormularioModelo{
		OngID: ongID,
		Nome:  nome,
	}
	check(t, mockDB.Create(&f))
	return f
}

func criarCampo(t *testing.T, formularioID uuid.UUID, nome string, ordem int) models.CampoFormulario {
	t.Helper()
	fid := formularioID
	c := models.CampoFormulario{
		FormularioModeloID: &fid,
		Nome:               nome,
		Ordem:              ordem,
		Configuracao:       []byte(`{"label":"` + nome + `","tipo":"texto","obrigatorio":true}`),
	}
	check(t, mockDB.Create(&c))
	return c
}

// ─────────────────────────────────────────────────────────────
// TestCriarFormulario
// ─────────────────────────────────────────────────────────────

func TestCriarFormulario(t *testing.T) {
	mockData()

	testCases := []TestCase{
		{
			Description: "Sucesso ao criar formulário simples",
			URL:         "/api/v1/formularios",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"ong_id": testOng.ID,
				"nome":   "Formulário de Adoção",
			}),
			ExpectedStatus: http.StatusCreated,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Formulário criado com sucesso", resp["message"])
				assert.NotNil(t, resp["formulario"])

				formData, _ := json.Marshal(resp["formulario"])
				var form models.FormularioModelo
				assert.NoError(t, json.Unmarshal(formData, &form))
				assert.Equal(t, "Formulário de Adoção", form.Nome)
				// Deve conter os 3 campos padrão (Nome, Email, Telefone)
				assert.Len(t, form.Campos, 3)
			},
		},
		{
			Description: "Sucesso ao criar formulário com campos customizados",
			URL:         "/api/v1/formularios",
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"ong_id": testOng.ID,
				"nome":   "Formulário Completo",
				"campos": []map[string]any{
					{
						"nome":  "Tem quintal?",
						"ordem": 0,
						"configuracao": map[string]any{
							"label":       "Tem quintal?",
							"tipo":        "radio",
							"obrigatorio": true,
							"ativo":       true,
							"opcoes":      []string{"Sim", "Não"},
						},
					},
				},
			}),
			ExpectedStatus: http.StatusCreated,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				formData, _ := json.Marshal(resp["formulario"])
				var form models.FormularioModelo
				assert.NoError(t, json.Unmarshal(formData, &form))
				// 3 padrão + 1 customizado
				assert.Len(t, form.Campos, 4)
			},
		},
		{
			Description:      "Erro - nome ausente",
			URL:              "/api/v1/formularios",
			Method:           http.MethodPost,
			Body:             mustMarshal(t, map[string]any{"ong_id": testOng.ID}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "nome é obrigatório",
		},
		{
			Description:      "Erro - ong_id ausente",
			URL:              "/api/v1/formularios",
			Method:           http.MethodPost,
			Body:             mustMarshal(t, map[string]any{"nome": "Sem ONG"}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ong_id é obrigatório",
		},
		{
			Description:      "Erro - ONG não encontrada",
			URL:              "/api/v1/formularios",
			Method:           http.MethodPost,
			Body:             mustMarshal(t, map[string]any{"nome": "X", "ong_id": uuid.New()}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "ONG não encontrada",
		},
		{
			Description:    "Erro - body inválido (JSON malformado)",
			URL:            "/api/v1/formularios",
			Method:         http.MethodPost,
			Body:           []byte(`{invalid}`),
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestListarFormularios
// ─────────────────────────────────────────────────────────────

func TestListarFormularios(t *testing.T) {
	mockData()

	criarFormulario(t, testOng.ID, "Form Alpha")
	criarFormulario(t, testOng.ID, "Form Beta")

	outraOng := models.Ong{
		Nome:            "Outra ONG",
		Email:           "outra@ong.org",
		Endereco:        "Rua Outra, 1",
		NomeResponsavel: "Responsável Outra",
		Telefone:        "11900000099",
		Regiao:          models.OngRegiaoSerra,
	}
	check(t, mockDB.Create(&outraOng))
	criarFormulario(t, outraOng.ID, "Form Outra ONG")

	testCases := []TestCase{
		{
			Description:    "Sucesso ao listar formulários",
			URL:            "/api/v1/formularios?page=1&limit=10",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp FormularioListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.GreaterOrEqual(t, len(resp.Dados), 2)
				assert.False(t, resp.ProximaPagina)
			},
		},
		{
			Description:    "Sucesso com paginação - limit 1",
			URL:            "/api/v1/formularios?page=1&limit=1",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp FormularioListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Len(t, resp.Dados, 1)
				assert.True(t, resp.ProximaPagina)
			},
		},
		{
			Description:    "Sucesso ao filtrar por ong_id",
			URL:            fmt.Sprintf("/api/v1/formularios?ong_id=%s", outraOng.ID),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp FormularioListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Len(t, resp.Dados, 1)
				assert.Equal(t, "Form Outra ONG", resp.Dados[0].Nome)
			},
		},
		{
			Description:    "Sucesso - filtro sem resultados retorna lista vazia",
			URL:            fmt.Sprintf("/api/v1/formularios?ong_id=%s", uuid.New()),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp FormularioListResponse
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Len(t, resp.Dados, 0)
			},
		},
		{
			Description:      "Erro - page inválido",
			URL:              "/api/v1/formularios?page=abc",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "page inválido",
		},
		{
			Description:      "Erro - limit inválido",
			URL:              "/api/v1/formularios?limit=xyz",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "limit inválido",
		},
		{
			Description:    "Erro - page zero",
			URL:            "/api/v1/formularios?page=0",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Description:    "Erro - limit zero",
			URL:            "/api/v1/formularios?limit=0",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestBuscarFormularioPorID
// ─────────────────────────────────────────────────────────────

func TestBuscarFormularioPorID(t *testing.T) {
	mockData()

	form := criarFormulario(t, testOng.ID, "Formulário Específico")

	testCases := []TestCase{
		{
			Description:    "Sucesso ao buscar formulário por ID",
			URL:            fmt.Sprintf("/api/v1/formularios/%s", form.ID),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp models.FormularioModelo
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, form.ID, resp.ID)
				assert.Equal(t, "Formulário Específico", resp.Nome)
			},
		},
		{
			Description:      "Erro - formulário não encontrado",
			URL:              fmt.Sprintf("/api/v1/formularios/%s", uuid.New()),
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Formulário não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/formularios/id-invalido",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do formulário inválido",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestAtualizarFormulario
// ─────────────────────────────────────────────────────────────

func TestAtualizarFormulario(t *testing.T) {
	mockData()

	form := criarFormulario(t, testOng.ID, "Nome Original")

	testCases := []TestCase{
		{
			Description: "Sucesso ao atualizar nome do formulário",
			URL:         fmt.Sprintf("/api/v1/formularios/%s", form.ID),
			Method:      http.MethodPut,
			Body:        mustMarshal(t, map[string]any{"nome": "Nome Atualizado"}),
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Formulário atualizado com sucesso", resp["message"])

				var updated models.FormularioModelo
				assert.NoError(t, mockDB.First(&updated, "id = ?", form.ID).Error)
				assert.Equal(t, "Nome Atualizado", updated.Nome)
			},
		},
		{
			Description:    "Sucesso - body sem nome não altera o registro",
			URL:            fmt.Sprintf("/api/v1/formularios/%s", form.ID),
			Method:         http.MethodPut,
			Body:           mustMarshal(t, map[string]any{}),
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Formulário atualizado com sucesso", resp["message"])
			},
		},
		{
			Description:      "Erro - formulário não encontrado",
			URL:              fmt.Sprintf("/api/v1/formularios/%s", uuid.New()),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"nome": "X"}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Formulário não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/formularios/id-invalido",
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"nome": "X"}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do formulário inválido",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestDeletarFormulario
// ─────────────────────────────────────────────────────────────

func TestDeletarFormulario(t *testing.T) {
	mockData()

	form := criarFormulario(t, testOng.ID, "Para Deletar")
	criarCampo(t, form.ID, "Campo A", 1)
	criarCampo(t, form.ID, "Campo B", 2)

	testCases := []TestCase{
		{
			Description:    "Sucesso ao deletar formulário (e seus campos em cascata)",
			URL:            fmt.Sprintf("/api/v1/formularios/%s", form.ID),
			Method:         http.MethodDelete,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Formulário e campos removidos com sucesso", resp["message"])

				var deleted models.FormularioModelo
				err := mockDB.First(&deleted, "id = ?", form.ID).Error
				assert.Error(t, err, "formulário deveria ter sido deletado")

				var campos []models.CampoFormulario
				mockDB.Where("formulario_modelo_id = ?", form.ID).Find(&campos)
				assert.Len(t, campos, 0, "campos deveriam ter sido removidos")
			},
		},
		{
			Description:      "Erro - formulário não encontrado",
			URL:              fmt.Sprintf("/api/v1/formularios/%s", uuid.New()),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Formulário não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/formularios/id-invalido",
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do formulário inválido",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestUploadImagemFormulario
// ─────────────────────────────────────────────────────────────

func TestUploadImagemFormulario(t *testing.T) {
	mockData()

	form := criarFormulario(t, testOng.ID, "Form Upload")

	testCases := []TestCase{
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/formularios/id-invalido/imagem",
			Method:           http.MethodPost,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do formulário inválido",
		},
		{
			Description:      "Erro - formulário não encontrado",
			URL:              fmt.Sprintf("/api/v1/formularios/%s/imagem", uuid.New()),
			Method:           http.MethodPost,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Formulário não encontrado",
		},
		{
			Description:      "Erro - nenhuma imagem enviada",
			URL:              fmt.Sprintf("/api/v1/formularios/%s/imagem", form.ID),
			Method:           http.MethodPost,
			Body:             []byte(``),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "nenhuma imagem enviada",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestDeletarImagemFormulario
// ─────────────────────────────────────────────────────────────

func TestDeletarImagemFormulario(t *testing.T) {
	mockData()

	formSemImagem := criarFormulario(t, testOng.ID, "Form Sem Imagem")

	formComImagem := models.FormularioModelo{
		OngID:     testOng.ID,
		Nome:      "Form Com Imagem",
		ImagemURL: "https://storage.example.com/formularios/capa.webp",
	}
	check(t, mockDB.Create(&formComImagem))

	testCases := []TestCase{
		{
			Description:      "Erro - formulário não possui imagem",
			URL:              fmt.Sprintf("/api/v1/formularios/%s/imagem", formSemImagem.ID),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "o formulário não possui imagem cadastrada",
		},
		{
			Description:      "Erro - formulário não encontrado",
			URL:              fmt.Sprintf("/api/v1/formularios/%s/imagem", uuid.New()),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Formulário não encontrado",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/formularios/id-invalido/imagem",
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do formulário inválido",
		},
		{
			Description:    "Sucesso ao remover imagem do formulário",
			URL:            fmt.Sprintf("/api/v1/formularios/%s/imagem", formComImagem.ID),
			Method:         http.MethodDelete,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Imagem do formulário removida com sucesso", resp["message"])

				var updated models.FormularioModelo
				assert.NoError(t, mockDB.First(&updated, "id = ?", formComImagem.ID).Error)
				assert.Equal(t, "", updated.ImagemURL)
			},
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestCriarCampoFormulario
// ─────────────────────────────────────────────────────────────

func TestCriarCampoFormulario(t *testing.T) {
	mockData()

	form := criarFormulario(t, testOng.ID, "Form Campos")

	campoBody := map[string]any{
		"nome":  "Tem outras pessoas na casa?",
		"ordem": 1,
		"configuracao": map[string]any{
			"label":       "Tem outras pessoas na casa?",
			"tipo":        "radio",
			"obrigatorio": true,
			"ativo":       true,
			"opcoes":      []string{"Sim", "Não"},
		},
	}

	testCases := []TestCase{
		{
			Description:    "Sucesso ao criar campo",
			URL:            fmt.Sprintf("/api/v1/formularios/%s/campos", form.ID),
			Method:         http.MethodPost,
			Body:           mustMarshal(t, campoBody),
			ExpectedStatus: http.StatusCreated,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Campo criado com sucesso", resp["message"])
				assert.NotNil(t, resp["campo"])
			},
		},
		{
			Description: "Erro - nome ausente",
			URL:         fmt.Sprintf("/api/v1/formularios/%s/campos", form.ID),
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"configuracao": map[string]any{"label": "X", "tipo": "texto"},
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "nome é obrigatório",
		},
		{
			Description: "Erro - label ausente na configuração",
			URL:         fmt.Sprintf("/api/v1/formularios/%s/campos", form.ID),
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"nome":         "Campo sem label",
				"configuracao": map[string]any{"tipo": "texto"},
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "configuracao.label é obrigatório",
		},
		{
			Description: "Erro - tipo ausente na configuração",
			URL:         fmt.Sprintf("/api/v1/formularios/%s/campos", form.ID),
			Method:      http.MethodPost,
			Body: mustMarshal(t, map[string]any{
				"nome":         "Campo sem tipo",
				"configuracao": map[string]any{"label": "Campo sem tipo"},
			}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "configuracao.tipo é obrigatório",
		},
		{
			Description:      "Erro - formulário não encontrado",
			URL:              fmt.Sprintf("/api/v1/formularios/%s/campos", uuid.New()),
			Method:           http.MethodPost,
			Body:             mustMarshal(t, campoBody),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Formulário não encontrado",
		},
		{
			Description:      "Erro - ID do formulário inválido",
			URL:              "/api/v1/formularios/id-invalido/campos",
			Method:           http.MethodPost,
			Body:             mustMarshal(t, campoBody),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do formulário inválido",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestAtualizarCampoFormulario
// ─────────────────────────────────────────────────────────────

func TestAtualizarCampoFormulario(t *testing.T) {
	mockData()

	form := criarFormulario(t, testOng.ID, "Form Atualizar Campo")
	campo := criarCampo(t, form.ID, "Campo Original", 1)

	testCases := []TestCase{
		{
			Description: "Sucesso ao atualizar campo",
			URL:         fmt.Sprintf("/api/v1/formularios/%s/campos/%s", form.ID, campo.ID),
			Method:      http.MethodPut,
			Body: mustMarshal(t, map[string]any{
				"nome":  "Campo Atualizado",
				"ordem": 2,
				"configuracao": map[string]any{
					"label":       "Campo Atualizado",
					"tipo":        "textarea",
					"obrigatorio": false,
				},
			}),
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Campo atualizado com sucesso", resp["message"])

				var updated models.CampoFormulario
				assert.NoError(t, mockDB.First(&updated, "id = ?", campo.ID).Error)
				assert.Equal(t, "Campo Atualizado", updated.Nome)
				assert.Equal(t, 2, updated.Ordem)
			},
		},
		{
			Description:      "Erro - campo não encontrado",
			URL:              fmt.Sprintf("/api/v1/formularios/%s/campos/%s", form.ID, uuid.New()),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"nome": "X"}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Campo não encontrado",
		},
		{
			Description:      "Erro - ID do formulário inválido",
			URL:              fmt.Sprintf("/api/v1/formularios/id-invalido/campos/%s", campo.ID),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"nome": "X"}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do formulário inválido",
		},
		{
			Description:      "Erro - ID do campo inválido",
			URL:              fmt.Sprintf("/api/v1/formularios/%s/campos/id-invalido", form.ID),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"nome": "X"}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do campo inválido",
		},
	}

	runTestCases(t, testCases)
}

// ─────────────────────────────────────────────────────────────
// TestDeletarCampoFormulario
// ─────────────────────────────────────────────────────────────

func TestDeletarCampoFormulario(t *testing.T) {
	mockData()

	form := criarFormulario(t, testOng.ID, "Form Deletar Campo")
	campo := criarCampo(t, form.ID, "Campo Para Deletar", 1)

	testCases := []TestCase{
		{
			Description:    "Sucesso ao deletar campo",
			URL:            fmt.Sprintf("/api/v1/formularios/%s/campos/%s", form.ID, campo.ID),
			Method:         http.MethodDelete,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]any
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, "Campo removido com sucesso", resp["message"])

				var deleted models.CampoFormulario
				err := mockDB.First(&deleted, "id = ?", campo.ID).Error
				assert.Error(t, err, "campo deveria ter sido deletado")
			},
		},
		{
			Description:      "Erro - campo não encontrado",
			URL:              fmt.Sprintf("/api/v1/formularios/%s/campos/%s", form.ID, uuid.New()),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "Campo não encontrado",
		},
		{
			Description:      "Erro - ID do formulário inválido",
			URL:              fmt.Sprintf("/api/v1/formularios/id-invalido/campos/%s", campo.ID),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do formulário inválido",
		},
		{
			Description:      "Erro - ID do campo inválido",
			URL:              fmt.Sprintf("/api/v1/formularios/%s/campos/id-invalido", form.ID),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID do campo inválido",
		},
	}

	runTestCases(t, testCases)
}