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

func TestCriarOng(t *testing.T) {
	mockData()

	sucessoBody := map[string]any{
		"email":            "contato@minhaong.org",
		"endereco":         "Rua das Flores, 123",
		"nome":             "ONG Teste",
		"nome_responsavel": "João Silva",
		"telefone":         "11999999999",
		"whatsapp":         "https://wa.me/5511999999999",
		"descricao":        "ONG dedicada ao resgate de animais",
		"site":             "https://minhaong.org",
		"instagram":        "@minhaong",
		"regiao":           "Serra",
	}

	emailDuplicadoBody := map[string]any{
		"email":            testOng.Email, // "base@ong.org" — já existe via mockData()
		"endereco":         "Rua B, 456",
		"nome":             "Outra ONG",
		"nome_responsavel": "Maria Souza",
		"telefone":         "11988888888",
		"regiao":           "Litoral",
	}

	testCases := []TestCase{
		{
			Description:    "Sucesso ao criar ONG",
			Body:           mustMarshal(t, sucessoBody),
			URL:            "/api/v1/ongs",
			Method:         http.MethodPost,
			ExpectedStatus: http.StatusCreated,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "ONG criada com sucesso", response["message"])
				assert.NotNil(t, response["ong"])

				ongData, _ := json.Marshal(response["ong"])
				var ong models.Ong
				err = json.Unmarshal(ongData, &ong)
				assert.NoError(t, err)

				assert.Equal(t, "ONG Teste", ong.Nome)
				assert.Equal(t, "contato@minhaong.org", ong.Email)
				assert.Equal(t, models.OngRegiao("Serra"), ong.Regiao)

				// Remove a ONG criada pelo user_id fixo para que o próximo test case
				// não caia no check "usuário já possui uma ONG cadastrada".
				mockDB.Unscoped().Delete(&ong)
			},
		},
		{
			Description:      "Erro - e-mail duplicado",
			Body:             mustMarshal(t, emailDuplicadoBody),
			URL:              "/api/v1/ongs",
			Method:           http.MethodPost,
			ExpectedStatus:   http.StatusConflict,
			ExpectedErrorMsg: "email já cadastrado",
		},
		{
			Description:    "Erro - body inválido (JSON malformado)",
			Body:           []byte(`{invalid-json}`),
			URL:            "/api/v1/ongs",
			Method:         http.MethodPost,
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

func TestListarOngs(t *testing.T) {
	mockData()

	ong1 := models.Ong{
		Nome:            "ONG Alpha",
		Email:           "alpha@ong.org",
		Endereco:        "Rua Alpha, 1",
		NomeResponsavel: "Responsável Alpha",
		Telefone:        "11911111111",
		Regiao:          "Serra",
	}

	ong2 := models.Ong{
		Nome:            "ONG Beta",
		Email:           "beta@ong.org",
		Endereco:        "Rua Beta, 2",
		NomeResponsavel: "Responsável Beta",
		Telefone:        "11922222222",
		Regiao:          "Litoral",
	}

	check(t, mockDB.Create(&ong1))
	check(t, mockDB.Create(&ong2))

	testCases := []TestCase{
		{
			Description:    "Sucesso ao listar ONGs - primeira página",
			URL:            "/api/v1/ongs?page=1&limit=10",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response OngListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.GreaterOrEqual(t, len(response.Dados), 2)
				assert.False(t, response.ProximaPagina)
			},
		},
		{
			Description:    "Sucesso ao listar ONGs com paginação - limit 1",
			URL:            "/api/v1/ongs?page=1&limit=1",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response OngListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Len(t, response.Dados, 1)
				assert.True(t, response.ProximaPagina)
			},
		},
		{
			Description:    "Sucesso ao filtrar ONG por nome",
			URL:            "/api/v1/ongs?nome=Alpha",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response OngListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Len(t, response.Dados, 1)
				assert.Equal(t, "ONG Alpha", response.Dados[0].Nome)
			},
		},
		{
			Description:    "Sucesso ao filtrar ONG por email",
			URL:            "/api/v1/ongs?email=beta@ong.org",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response OngListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Len(t, response.Dados, 1)
				assert.Equal(t, "beta@ong.org", response.Dados[0].Email)
			},
		},
		{
			Description:    "Sucesso - filtro sem resultados retorna lista vazia",
			URL:            "/api/v1/ongs?nome=NomeQueNaoExiste",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response OngListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Len(t, response.Dados, 0)
			},
		},
		{
			Description:    "Erro - page inválido (não numérico)",
			URL:            "/api/v1/ongs?page=abc",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusBadRequest,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, fmt.Sprintf("%v", response["error"]), "page inválido")
			},
		},
		{
			Description:    "Erro - limit inválido (não numérico)",
			URL:            "/api/v1/ongs?limit=xyz",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusBadRequest,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, fmt.Sprintf("%v", response["error"]), "limit inválido")
			},
		},
		{
			Description:    "Erro - page zero (não positivo)",
			URL:            "/api/v1/ongs?page=0",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Description:    "Erro - limit zero (não positivo)",
			URL:            "/api/v1/ongs?limit=0",
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

func TestBuscarOngPorID(t *testing.T) {
	mockData()

	ong := models.Ong{
		Nome:            "ONG Específica",
		Email:           "especifica@ong.org",
		Endereco:        "Rua Específica, 99",
		NomeResponsavel: "Responsável",
		Telefone:        "11933333333",
		Regiao:          "Serra",
	}

	check(t, mockDB.Create(&ong))

	testCases := []TestCase{
		{
			Description:    "Sucesso ao buscar ONG por ID",
			URL:            fmt.Sprintf("/api/v1/ongs/%s", ong.ID),
			Method:         http.MethodGet,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.Ong
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, ong.ID, response.ID)
				assert.Equal(t, "ONG Específica", response.Nome)
				assert.Equal(t, "especifica@ong.org", response.Email)
			},
		},
		{
			Description:      "Erro - ONG não encontrada",
			URL:              fmt.Sprintf("/api/v1/ongs/%s", uuid.New()),
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "ONG não encontrada",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/ongs/id-invalido",
			Method:           http.MethodGet,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID da ONG inválido",
		},
	}

	runTestCases(t, testCases)
}

func TestAtualizarOng(t *testing.T) {
	mockData()

	ong := models.Ong{
		Nome:            "ONG Original",
		Email:           "original@ong.org",
		Endereco:        "Rua Original, 1",
		NomeResponsavel: "Responsável Original",
		Telefone:        "11944444444",
		Descricao:       "Descrição original",
		Regiao:          "Serra",
	}

	check(t, mockDB.Create(&ong))

	novoNome := "ONG Atualizada"
	novoEmail := "atualizado@ong.org"
	novaDescricao := "Descrição atualizada"
	novoEndereco := "Rua Nova, 200"
	novoTelefone := "11955555555"
	novoInstagram := "@ong_atualizada"
	novoSite := "https://ongAtualizada.org"
	novaRegiao := "Litoral"

	testCases := []TestCase{
		{
			Description: "Sucesso ao atualizar ONG",
			URL:         fmt.Sprintf("/api/v1/ongs/%s", ong.ID),
			Method:      http.MethodPut,
			Body: mustMarshal(t, map[string]any{
				"nome":      novoNome,
				"email":     novoEmail,
				"descricao": novaDescricao,
				"endereco":  novoEndereco,
				"telefone":  novoTelefone,
				"instagram": novoInstagram,
				"site":      novoSite,
				"regiao":    novaRegiao,
			}),
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "ONG atualizada com sucesso", response["message"])

				var updated models.Ong
				err = mockDB.First(&updated, "id = ?", ong.ID).Error
				assert.NoError(t, err)

				assert.Equal(t, novoNome, updated.Nome)
				assert.Equal(t, novoEmail, updated.Email)
				assert.Equal(t, novaDescricao, updated.Descricao)
				assert.Equal(t, novoEndereco, updated.Endereco)
				assert.Equal(t, novoTelefone, updated.Telefone)
				assert.Equal(t, novoInstagram, updated.Instagram)
				assert.Equal(t, novoSite, updated.Site)
				assert.Equal(t, models.OngRegiao(novaRegiao), updated.Regiao)
			},
		},
		{
			Description: "Sucesso - atualização parcial (apenas nome)",
			URL:         fmt.Sprintf("/api/v1/ongs/%s", ong.ID),
			Method:      http.MethodPut,
			Body: mustMarshal(t, map[string]any{
				"nome": "Apenas Nome Alterado",
			}),
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "ONG atualizada com sucesso", response["message"])

				var updated models.Ong
				err = mockDB.First(&updated, "id = ?", ong.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, "Apenas Nome Alterado", updated.Nome)
			},
		},
		{
			Description:      "Erro - ONG não encontrada",
			URL:              fmt.Sprintf("/api/v1/ongs/%s", uuid.New()),
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"nome": "Qualquer"}),
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "ONG não encontrada",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/ongs/id-invalido",
			Method:           http.MethodPut,
			Body:             mustMarshal(t, map[string]any{"nome": "Qualquer"}),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID da ONG inválido",
		},
		{
			Description:    "Erro - body inválido (JSON malformado)",
			URL:            fmt.Sprintf("/api/v1/ongs/%s", ong.ID),
			Method:         http.MethodPut,
			Body:           []byte(`{invalid}`),
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	runTestCases(t, testCases)
}

func TestDeletarOng(t *testing.T) {
	mockData()

	ong := models.Ong{
		Nome:            "ONG Para Deletar",
		Email:           "deletar@ong.org",
		Endereco:        "Rua Temporária, 0",
		NomeResponsavel: "Responsável Temp",
		Telefone:        "11900000000",
		Regiao:          "Serra",
	}

	check(t, mockDB.Create(&ong))

	testCases := []TestCase{
		{
			Description:    "Sucesso ao deletar ONG",
			URL:            fmt.Sprintf("/api/v1/ongs/%s", ong.ID),
			Method:         http.MethodDelete,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "ONG removida com sucesso", response["message"])

				var deleted models.Ong
				err = mockDB.First(&deleted, "id = ?", ong.ID).Error
				assert.Error(t, err, "ONG deveria ter sido deletada do banco")
			},
		},
		{
			Description:      "Erro - ONG não encontrada",
			URL:              fmt.Sprintf("/api/v1/ongs/%s", uuid.New()),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "ONG não encontrada",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/ongs/id-invalido",
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID da ONG inválido",
		},
	}

	runTestCases(t, testCases)
}

func TestUploadLogoOng(t *testing.T) {
	mockData()

	ong := models.Ong{
		Nome:            "ONG Logo Test",
		Email:           "logo@ong.org",
		Endereco:        "Rua Logo, 1",
		NomeResponsavel: "Responsável Logo",
		Telefone:        "11966666666",
		Regiao:          "Serra",
	}

	check(t, mockDB.Create(&ong))

	testCases := []TestCase{
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/ongs/id-invalido/logo",
			Method:           http.MethodPost,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID da ONG inválido",
		},
		{
			Description:      "Erro - ONG não encontrada",
			URL:              fmt.Sprintf("/api/v1/ongs/%s/logo", uuid.New()),
			Method:           http.MethodPost,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "ONG não encontrada",
		},
		{
			Description:      "Erro - nenhuma imagem enviada",
			URL:              fmt.Sprintf("/api/v1/ongs/%s/logo", ong.ID),
			Method:           http.MethodPost,
			Body:             []byte(``),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "nenhuma imagem enviada",
		},
	}

	runTestCases(t, testCases)
}

func TestDeletarLogoOng(t *testing.T) {
	mockData()

	ongSemLogo := models.Ong{
		Nome:            "ONG Sem Logo",
		Email:           "semlogo@ong.org",
		Endereco:        "Rua Sem Logo, 1",
		NomeResponsavel: "Responsável Sem Logo",
		Telefone:        "11977777777",
		Regiao:          "Serra",
		Logo:            "",
	}

	ongComLogo := models.Ong{
		Nome:            "ONG Com Logo",
		Email:           "comlogo@ong.org",
		Endereco:        "Rua Com Logo, 2",
		NomeResponsavel: "Responsável Com Logo",
		Telefone:        "11988888888",
		Regiao:          "Litoral",
		Logo:            "https://storage.example.com/ongs/logo-existente.webp",
	}

	check(t, mockDB.Create(&ongSemLogo))
	check(t, mockDB.Create(&ongComLogo))

	testCases := []TestCase{
		{
			Description:      "Erro - ONG não possui logo",
			URL:              fmt.Sprintf("/api/v1/ongs/%s/logo", ongSemLogo.ID),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ONG não possui logo",
		},
		{
			Description:      "Erro - ONG não encontrada",
			URL:              fmt.Sprintf("/api/v1/ongs/%s/logo", uuid.New()),
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusNotFound,
			ExpectedErrorMsg: "ONG não encontrada",
		},
		{
			Description:      "Erro - ID inválido",
			URL:              "/api/v1/ongs/id-invalido/logo",
			Method:           http.MethodDelete,
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedErrorMsg: "ID da ONG inválido",
		},
		{
			Description:    "Sucesso ao remover logo da ONG",
			URL:            fmt.Sprintf("/api/v1/ongs/%s/logo", ongComLogo.ID),
			Method:         http.MethodDelete,
			ExpectedStatus: http.StatusOK,
			CheckResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "Logo removido com sucesso", response["message"])

				var updated models.Ong
				err = mockDB.First(&updated, "id = ?", ongComLogo.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, "", updated.Logo)
			},
		},
	}

	runTestCases(t, testCases)
}