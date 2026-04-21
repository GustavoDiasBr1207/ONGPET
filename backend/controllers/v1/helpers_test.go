package v1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v1 "ongpet/controllers/v1"
	"ongpet/database"
	"ongpet/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// OngListResponse espelha a estrutura retornada por ReadOngs.
type OngListResponse struct {
	Dados          []models.Ong `json:"dados"`
	TotalRegistros int64        `json:"total_registros"`
	TotalPaginas   int          `json:"total_paginas"`
	ProximaPagina  bool         `json:"proxima_pagina"`
}

// PetListResponse espelha a estrutura retornada por ReadPets.
type PetListResponse struct {
	Dados          []models.Pet `json:"dados"`
	TotalRegistros int64        `json:"total_registros"`
	TotalPaginas   int          `json:"total_paginas"`
	ProximaPagina  bool         `json:"proxima_pagina"`
}

// TestCase descreve um caso de teste HTTP.
type TestCase struct {
	Description      string
	Method           string
	URL              string
	Body             []byte
	ExpectedStatus   int
	ExpectedErrorMsg string
	CheckResponse    func(t *testing.T, w *httptest.ResponseRecorder)
}

// mockDB é a conexão SQLite in-memory compartilhada entre os testes.
var mockDB *gorm.DB

// testOng é uma ONG pré-inserida por mockData() para ser usada nos testes.
var testOng models.Ong

// mockData reinicia o banco in-memory e insere dados base para os testes.
// Deve ser chamado no início de cada TestXxx que precisar de estado limpo.
func mockData() {
	var err error
	mockDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic("falha ao criar banco in-memory: " + err.Error())
	}

	if err := mockDB.AutoMigrate(&models.Ong{}, &models.Pet{}, &models.PetImage{}); err != nil {
		panic("falha ao migrar models: " + err.Error())
	}

	// Substitui o db global pelo mock para que GetDB() e GetUserDB() o usem.
	database.SetDB(mockDB)

	// ONG base disponível em todos os testes via testOng.
	testOng = models.Ong{
		Nome:            "ONG Base",
		Email:           "base@ong.org",
		Endereco:        "Rua Base, 1",
		NomeResponsavel: "Responsável Base",
		Telefone:        "11900000001",
		Regiao:          models.OngRegiaoSerra,
	}
	if err := mockDB.Create(&testOng).Error; err != nil {
		panic("falha ao criar testOng: " + err.Error())
	}
}

// check interrompe o teste imediatamente se result.Error não for nil.
func check(t *testing.T, result *gorm.DB) {
	t.Helper()
	assert.NoError(t, result.Error)
}

// mustMarshal serializa v para JSON, falhando o teste em caso de erro.
func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	assert.NoError(t, err)
	return b
}

// newRouter monta um gin.Engine com todas as rotas do servidor real,
// mas sem middleware de autenticação (token injetado diretamente).
func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	wrap := func(fn func(*gin.Context) error) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("token", "test-token")
			if err := fn(c); err != nil {
				status := http.StatusInternalServerError

				msg := err.Error()
				switch msg {
				case "ID da ONG inválido",
					"page inválido",
					"limit inválido",
					"nenhuma imagem enviada",
					"tipo de imagem inválido",
					"ONG não possui logo",
					"ID do pet inválido",
					"ID da imagem inválido",
					"ong_id é obrigatório",
					"o pet pode ter no máximo 5 imagens":
					status = http.StatusBadRequest
				case "ONG não encontrada",
					"Pet não encontrado",
					"Imagem não encontrada":
					status = http.StatusNotFound
				case "email já cadastrado":
					status = http.StatusConflict
				default:
					if strings.HasPrefix(msg, "body inválido") {
						status = http.StatusBadRequest
					}
				}

				if !c.Writer.Written() {
					c.JSON(status, gin.H{"error": msg})
				}
			}
		}
	}

	api := r.Group("/api/v1")
	{
		ongs := api.Group("/ongs")
		ongs.GET("", wrap(v1.ReadOngs))
		ongs.GET("/:id", wrap(v1.ReadOng))
		ongs.POST("", wrap(v1.CreateOng))
		ongs.PUT("/:id", wrap(v1.UpdateOng))
		ongs.DELETE("/:id", wrap(v1.DeleteOng))
		ongs.POST("/:id/logo", wrap(v1.UploadOngLogo))
		ongs.DELETE("/:id/logo", wrap(v1.DeleteOngLogo))

		pets := api.Group("/pets")
		pets.GET("", wrap(v1.ReadPets))
		pets.GET("/:id", wrap(v1.ReadPet))
		pets.POST("", wrap(v1.CreatePet))
		pets.PUT("/:id", wrap(v1.UpdatePet))
		pets.DELETE("/:id", wrap(v1.DeletePet))
		pets.POST("/:id/imagens", wrap(v1.UploadPetImages))
		pets.DELETE("/:id/imagens/:imageId", wrap(v1.DeletePetImage))
	}

	return r
}

// runTestCases executa uma lista de TestCase contra o router de testes.
func runTestCases(t *testing.T, cases []TestCase) {
	t.Helper()
	r := newRouter()

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Description, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tc.Body != nil {
				bodyReader = bytes.NewReader(tc.Body)
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}

			req, err := http.NewRequest(tc.Method, tc.URL, bodyReader)
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.ExpectedStatus, w.Code, "status inesperado para: %s", tc.Description)

			if tc.ExpectedErrorMsg != "" {
				var body map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &body)
				assert.NoError(t, err)
				assert.Contains(t, body["error"], tc.ExpectedErrorMsg)
			}

			if tc.CheckResponse != nil {
				tc.CheckResponse(t, w)
			}
		})
	}
}