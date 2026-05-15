package v1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ongpet/controllers"
	v1 "ongpet/controllers/v1"
	"ongpet/database"
	"ongpet/models"
	"ongpet/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// ─────────────────────────────────────────────────────────────
// Response types
// ─────────────────────────────────────────────────────────────

type OngListResponse struct {
	Dados          []models.Ong `json:"dados"`
	TotalRegistros int64        `json:"total_registros"`
	TotalPaginas   int          `json:"total_paginas"`
	ProximaPagina  bool         `json:"proxima_pagina"`
}

type PetListResponse struct {
	Dados          []models.Pet `json:"dados"`
	TotalRegistros int64        `json:"total_registros"`
	TotalPaginas   int          `json:"total_paginas"`
	ProximaPagina  bool         `json:"proxima_pagina"`
}

type FormularioListResponse struct {
	Dados          []models.FormularioModelo `json:"dados"`
	TotalRegistros int64                     `json:"total_registros"`
	TotalPaginas   int                       `json:"total_paginas"`
	ProximaPagina  bool                      `json:"proxima_pagina"`
}

type PedidoAdocaoListResponse struct {
	Dados          []models.PedidoAdocaoDTO `json:"dados"`
	TotalRegistros int64                    `json:"total_registros"`
	TotalPaginas   int                      `json:"total_paginas"`
	ProximaPagina  bool                     `json:"proxima_pagina"`
}

type BannerListResponse struct {
	Dados          []models.Banner `json:"dados"`
	TotalRegistros int64           `json:"total_registros"`
	TotalPaginas   int             `json:"total_paginas"`
	ProximaPagina  bool            `json:"proxima_pagina"`
}

// ─────────────────────────────────────────────────────────────
// TestCase
// ─────────────────────────────────────────────────────────────

type TestCase struct {
	Description      string
	Method           string
	URL              string
	Body             []byte
	ExpectedStatus   int
	ExpectedErrorMsg string
	CheckResponse    func(t *testing.T, w *httptest.ResponseRecorder)
}

// ─────────────────────────────────────────────────────────────
// Shared state
// ─────────────────────────────────────────────────────────────

var mockDB *gorm.DB
var testOng models.Ong

// ─────────────────────────────────────────────────────────────
// mockData — banco limpo + ONG base + mocks externos
// ─────────────────────────────────────────────────────────────

func mockData() {
	// Mock do Supabase Auth — evita chamada HTTP real nos testes
	utils.UpdateUserOngMetadataFn = func(userID string, ong utils.OngMetadata, userJWT string) error {
		return nil
	}

	var err error
	mockDB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		panic("falha ao criar banco in-memory: " + err.Error())
	}

	if err := mockDB.AutoMigrate(
		&models.Ong{},
		&models.Pet{},
		&models.PetImage{},
		&models.FormularioModelo{},
		&models.CampoFormulario{},
		&models.RespostaFormulario{},
		&models.PedidoAdocao{},
		&models.Acompanhamento{},
		&models.LogAcompanhamento{},
		&models.Banner{},
	); err != nil {
		panic("falha ao migrar models: " + err.Error())
	}

	mockDB.Exec("DELETE FROM pedido_adocao")
	mockDB.Exec("DELETE FROM resposta_formulario")
	mockDB.Exec("DELETE FROM campo_formulario")
	mockDB.Exec("DELETE FROM formulario_modelo")
	mockDB.Exec("DELETE FROM pet_image")
	mockDB.Exec("DELETE FROM pet")
	mockDB.Exec("DELETE FROM ong")
	mockDB.Exec("DELETE FROM log_acompanhamento")
	mockDB.Exec("DELETE FROM acompanhamento")
	mockDB.Exec("DELETE FROM banner")

	database.SetDB(mockDB)

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

// ─────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────

func check(t *testing.T, result *gorm.DB) {
	t.Helper()
	assert.NoError(t, result.Error)
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	assert.NoError(t, err)
	return b
}

// ─────────────────────────────────────────────────────────────
// Router
// ─────────────────────────────────────────────────────────────

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	wrap := func(fn func(*gin.Context) error) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("token", "test-token")
			c.Set("user_id", "00000000-0000-0000-0000-000000000001")
			if err := fn(c); err != nil {
				if !c.Writer.Written() {
					c.JSON(controllers.HttpStatusForError(err.Error()), gin.H{"error": err.Error()})
				}
			}
		}
	}

	api := r.Group("/api/v1")
	{
		// ONGs
		ongs := api.Group("/ongs")
		ongs.GET("", wrap(v1.ReadOngs))
		ongs.GET("/:id", wrap(v1.ReadOng))
		ongs.POST("", wrap(v1.CreateOng))
		ongs.PUT("/:id", wrap(v1.UpdateOng))
		ongs.DELETE("/:id", wrap(v1.DeleteOng))
		ongs.POST("/:id/logo", wrap(v1.UploadOngLogo))
		ongs.DELETE("/:id/logo", wrap(v1.DeleteOngLogo))

		// Pets
		pets := api.Group("/pets")
		pets.GET("", wrap(v1.ReadPets))
		pets.GET("/:id", wrap(v1.ReadPet))
		pets.POST("", wrap(v1.CreatePet))
		pets.PUT("/:id", wrap(v1.UpdatePet))
		pets.DELETE("/:id", wrap(v1.DeletePet))
		pets.POST("/:id/imagens", wrap(v1.UploadPetImages))
		pets.DELETE("/:id/imagens/:imageId", wrap(v1.DeletePetImage))

		// Formulários
		forms := api.Group("/formularios")
		forms.GET("", wrap(v1.ReadFormularios))
		forms.GET("/:id", wrap(v1.ReadFormulario))
		forms.POST("", wrap(v1.CreateFormulario))
		forms.PUT("/:id", wrap(v1.UpdateFormulario))
		forms.DELETE("/:id", wrap(v1.DeleteFormulario))
		forms.POST("/:id/imagem", wrap(v1.UploadFormularioImagem))
		forms.DELETE("/:id/imagem", wrap(v1.DeleteFormularioImagem))
		forms.POST("/:id/campos", wrap(v1.CreateCampoFormulario))
		forms.PUT("/:id/campos/:campoId", wrap(v1.UpdateCampoFormulario))
		forms.DELETE("/:id/campos/:campoId", wrap(v1.DeleteCampoFormulario))

		// Pedidos de Adoção
		pedidos := api.Group("/pedidos-adocao")
		pedidos.GET("", wrap(v1.ReadPedidosAdocao))
		pedidos.GET("/:id", wrap(v1.ReadPedidoAdocao))
		pedidos.POST("", wrap(v1.CreatePedidoAdocao))
		pedidos.DELETE("/:id", wrap(v1.DeletePedidoAdocao))
		pedidos.PUT("/:id/status", wrap(v1.UpdateStatusPedidoAdocao))
		pedidos.POST("/:pedidoId/respostas/:respostaId/imagem", wrap(v1.UploadRespostaImagem))

		// Acompanhamentos
		acomps := api.Group("/acompanhamentos")
		acomps.POST("", wrap(v1.CreateAcompanhamento))
		acomps.GET("", wrap(v1.ListAcompanhamentos))
		acomps.POST("/:id/logs", wrap(v1.CreateLogAcompanhamento))
		acomps.GET("/:id/logs", wrap(v1.GetAcompanhamentoLogs))

		// Banners
		banners := api.Group("/banners")
		banners.GET("", wrap(v1.ReadBanners))
		banners.GET("/:id", wrap(v1.ReadBanner))
		banners.POST("", wrap(v1.CreateBanner))
		banners.PUT("/:id", wrap(v1.UpdateBanner))
		banners.DELETE("/:id", wrap(v1.DeleteBanner))
		banners.POST("/:id/imagem", wrap(v1.UploadBannerImage))
		banners.DELETE("/:id/imagem", wrap(v1.DeleteBannerImage))
	}

	return r
}

// ─────────────────────────────────────────────────────────────
// runTestCases
// ─────────────────────────────────────────────────────────────

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
				assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
				assert.Contains(t, body["error"], tc.ExpectedErrorMsg)
			}

			if tc.CheckResponse != nil {
				tc.CheckResponse(t, w)
			}
		})
	}
}