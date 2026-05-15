package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"ongpet/database"
	"ongpet/models"
	"ongpet/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────
// INPUT TYPES
// ─────────────────────────────────────────────────────────────

type CreateCampoInput struct {
	Nome         string                   `json:"nome"`
	Ordem        int                      `json:"ordem"`
	Configuracao models.CampoConfiguracao `json:"configuracao"`
}

type CreateFormularioModeloInput struct {
	OngID  uuid.UUID          `json:"ong_id"`
	Nome   string             `json:"nome"`
	Campos []CreateCampoInput `json:"campos,omitempty"`
}

type UpdateFormularioModeloInput struct {
	Nome string `json:"nome"`
}

type FormularioListResponse struct {
	Dados          []models.FormularioModelo `json:"dados"`
	TotalRegistros int64                     `json:"total_registros"`
	TotalPaginas   int                       `json:"total_paginas"`
	ProximaPagina  bool                      `json:"proxima_pagina"`
}

// ─────────────────────────────────────────────────────────────
// FORMULARIO MODELO — CRUD
// ─────────────────────────────────────────────────────────────

// @Summary Lista todos os Formulários Modelo
// @Tags FormularioModelo
// @Produce json
// @Param ong_id query string false "Filtrar por ONG"
// @Param page   query int    false "Página (default: 1)"
// @Param limit  query int    false "Itens por página (default: 10)"
// @Success 200 {object} v1.FormularioListResponse
// @Router /api/v1/formularios [get]
func ReadFormularios(c *gin.Context) error {
	// Rota pública — sem token, sem RLS
	db := database.GetDB()
	query := db.Model(&models.FormularioModelo{})

	if ongID := strings.TrimSpace(c.Query("ong_id")); ongID != "" {
		query = query.Where("ong_id = ?", ongID)
	}

	pageStr  := c.DefaultQuery("page",  "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		return errors.New("page inválido")
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return errors.New("limit inválido")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return err
	}

	offset     := (page - 1) * limit
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	var formularios []models.FormularioModelo
	if err := query.
		Preload("Campos", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("ordem ASC")
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit + 1).
		Find(&formularios).Error; err != nil {
		return err
	}

	hasNext := false
	if len(formularios) > limit {
		hasNext = true
		formularios = formularios[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"dados":           formularios,
		"total_registros": total,
		"total_paginas":   totalPages,
		"proxima_pagina":  hasNext,
	})
	return nil
}

// @Summary Busca um Formulário Modelo pelo ID
// @Tags FormularioModelo
// @Produce json
// @Param id path string true "ID do Formulário"
// @Success 200 {object} models.FormularioModelo
// @Router /api/v1/formularios/{id} [get]
func ReadFormulario(c *gin.Context) error {
	// Rota pública — sem token, sem RLS
	db := database.GetDB()

	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	var formulario models.FormularioModelo
	if err := db.
		Preload("Campos", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("ordem ASC")
		}).
		First(&formulario, "id = ?", formularioID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Formulário não encontrado")
		}
		return err
	}

	c.JSON(http.StatusOK, formulario)
	return nil
}

// @Summary Cria um novo Formulário Modelo
// @Tags FormularioModelo
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Router /api/v1/formularios [post]
func CreateFormulario(c *gin.Context) error {
	var req CreateFormularioModeloInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return fmt.Errorf("body inválido: %w", err)
	}

	req.Nome = strings.TrimSpace(req.Nome)
	if req.Nome == "" {
		return errors.New("nome é obrigatório")
	}
	if req.OngID == uuid.Nil {
		return errors.New("ong_id é obrigatório")
	}

	var formulario models.FormularioModelo
	err := database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		if err := tx.First(&models.Ong{}, "id = ?", req.OngID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("ONG não encontrada")
			}
			return err
		}

		formulario = models.FormularioModelo{
			OngID: req.OngID,
			Nome:  req.Nome,
		}
		if err := tx.Create(&formulario).Error; err != nil {
			return err
		}

		camposPadroes := []CreateCampoInput{
			{
				Nome:  "Nome",
				Ordem: 0,
				Configuracao: models.CampoConfiguracao{
					Label:       "Nome",
					Placeholder: "Seu nome completo",
					Tipo:        models.TipoCampoTexto,
					Obrigatorio: true,
					Ativo:       true,
				},
			},
			{
				Nome:  "Email",
				Ordem: 1,
				Configuracao: models.CampoConfiguracao{
					Label:       "Email",
					Placeholder: "seu_email@exemplo.com",
					Tipo:        models.TipoCampoEmail,
					Obrigatorio: true,
					Ativo:       true,
				},
			},
			{
				Nome:  "Telefone",
				Ordem: 2,
				Configuracao: models.CampoConfiguracao{
					Label:       "Telefone",
					Placeholder: "(XX) 99999-9999",
					Tipo:        models.TipoCampoTelefone,
					Obrigatorio: true,
					Ativo:       true,
				},
			},
		}

		for _, campoInput := range camposPadroes {
			campo, err := buildCampo(formulario.ID, campoInput)
			if err != nil {
				return err
			}
			if err := tx.Create(&campo).Error; err != nil {
				return err
			}
		}

		for i, campoInput := range req.Campos {
			if campoInput.Ordem == 0 {
				campoInput.Ordem = i + 3
			}
			campo, err := buildCampo(formulario.ID, campoInput)
			if err != nil {
				return err
			}
			if err := tx.Create(&campo).Error; err != nil {
				return err
			}
		}

		return tx.Preload("Campos", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("ordem ASC")
		}).First(&formulario, "id = ?", formulario.ID).Error
	})
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Formulário criado com sucesso",
		"formulario": formulario,
	})
	return nil
}

// @Summary Atualiza o nome de um Formulário Modelo
// @Tags FormularioModelo
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "ID do Formulário"
// @Router /api/v1/formularios/{id} [put]
func UpdateFormulario(c *gin.Context) error {
	var req UpdateFormularioModeloInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return fmt.Errorf("body inválido: %w", err)
	}

	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	var formulario models.FormularioModelo
	err = database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		if err := tx.First(&formulario, "id = ?", formularioID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("Formulário não encontrado")
			}
			return err
		}

		if nome := strings.TrimSpace(req.Nome); nome != "" {
			formulario.Nome = nome
		}

		if err := tx.Save(&formulario).Error; err != nil {
			return err
		}
		return tx.Preload("Campos", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("ordem ASC")
		}).First(&formulario, "id = ?", formularioID).Error
	})
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Formulário atualizado com sucesso",
		"formulario": formulario,
	})
	return nil
}

// @Summary Remove um Formulário Modelo
// @Tags FormularioModelo
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID do Formulário"
// @Router /api/v1/formularios/{id} [delete]
func DeleteFormulario(c *gin.Context) error {
	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	return database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		if err := tx.First(&models.FormularioModelo{}, "id = ?", formularioID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("Formulário não encontrado")
			}
			return err
		}

		if err := tx.Where("campo_formulario_id IN (?)",
			tx.Model(&models.CampoFormulario{}).
				Select("id").
				Where("formulario_modelo_id = ?", formularioID),
		).Delete(&models.RespostaFormulario{}).Error; err != nil {
			return err
		}

		if err := tx.Where("formulario_modelo_id = ?", formularioID).
			Delete(&models.CampoFormulario{}).Error; err != nil {
			return err
		}

		if err := tx.Where("id = ?", formularioID).
			Delete(&models.FormularioModelo{}).Error; err != nil {
			return err
		}

		c.JSON(http.StatusOK, gin.H{"message": "Formulário e campos removidos com sucesso"})
		return nil
	})
}

// ─────────────────────────────────────────────────────────────
// IMAGEM DE CAPA
// ─────────────────────────────────────────────────────────────

// @Summary Faz upload da imagem de capa de um Formulário
// @Tags FormularioModelo
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "ID do Formulário"
// @Router /api/v1/formularios/{id}/imagem [post]
func UploadFormularioImagem(c *gin.Context) error {
	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	token := c.GetString("token")

	// Valida ownership via RLS antes de processar o arquivo
	var formulario models.FormularioModelo
	err = database.WithUserDB(token, func(tx *gorm.DB) error {
		return tx.First(&formulario, "id = ?", formularioID).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Formulário não encontrado")
		}
		return err
	}

	file, err := c.FormFile("imagem")
	if err != nil {
		return errors.New("nenhuma imagem enviada")
	}
	if !utils.IsValidImage(file) {
		return errors.New("tipo de imagem inválido")
	}

	// Remove imagem anterior (se existir)
	if formulario.ImagemURL != "" {
		if objectPath, pathErr := utils.ExtractObjectPath(formulario.ImagemURL, utils.SupabaseBucketFormularios); pathErr == nil {
			if delErr := utils.DeleteFile(utils.SupabaseBucketFormularios, objectPath); delErr != nil {
				fmt.Printf("⚠️ Aviso: não foi possível deletar imagem anterior do storage: %s\n", delErr.Error())
			}
		}
	}

	optimized, err := utils.ResizeAndCompressImage(file, 1280, 70)
	if err != nil {
		return err
	}

	url, err := utils.UploadOptimizedFile(
		optimized.Buffer,
		optimized.ContentType,
		optimized.Extension,
		utils.SupabaseBucketFormularios,
		formularioID.String(),
		formulario.Nome,
		1,
		c.GetString("user_token"),
	)
	if err != nil {
		return err
	}

	// Salva nova URL via RLS
	err = database.WithUserDB(token, func(tx *gorm.DB) error {
		formulario.ImagemURL = url
		if err := tx.Save(&formulario).Error; err != nil {
			return err
		}
		return tx.Preload("Campos", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("ordem ASC")
		}).First(&formulario, "id = ?", formularioID).Error
	})
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Imagem do formulário atualizada com sucesso",
		"formulario": formulario,
	})
	return nil
}

// @Summary Remove a imagem de capa de um Formulário
// @Tags FormularioModelo
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID do Formulário"
// @Router /api/v1/formularios/{id}/imagem [delete]
func DeleteFormularioImagem(c *gin.Context) error {
	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	token := c.GetString("token")

	var (
		formulario models.FormularioModelo
		objectPath string
	)
	err = database.WithUserDB(token, func(tx *gorm.DB) error {
		if err := tx.First(&formulario, "id = ?", formularioID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("Formulário não encontrado")
			}
			return err
		}
		if formulario.ImagemURL == "" {
			return errors.New("o formulário não possui imagem cadastrada")
		}
		if path, pathErr := utils.ExtractObjectPath(formulario.ImagemURL, utils.SupabaseBucketFormularios); pathErr == nil {
			objectPath = path
		} else {
			fmt.Printf("⚠️ Aviso: não foi possível extrair path do storage: %s\n", pathErr.Error())
		}
		formulario.ImagemURL = ""
		if err := tx.Save(&formulario).Error; err != nil {
			return err
		}
		return tx.Preload("Campos", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("ordem ASC")
		}).First(&formulario, "id = ?", formularioID).Error
	})
	if err != nil {
		return err
	}

	// Remove do storage após commit do banco
	if objectPath != "" {
		if err := utils.DeleteFile(utils.SupabaseBucketFormularios, objectPath); err != nil {
			fmt.Printf("⚠️ Aviso: não foi possível deletar arquivo do storage: %s\n", err.Error())
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Imagem do formulário removida com sucesso",
		"formulario": formulario,
	})
	return nil
}

// ─────────────────────────────────────────────────────────────
// CAMPOS — CRUD
// ─────────────────────────────────────────────────────────────

// @Summary Adiciona um Campo ao Formulário
// @Tags FormularioModelo
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "ID do Formulário"
// @Router /api/v1/formularios/{id}/campos [post]
func CreateCampoFormulario(c *gin.Context) error {
	var req CreateCampoInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return fmt.Errorf("body inválido: %w", err)
	}

	req.Nome = strings.TrimSpace(req.Nome)
	if req.Nome == "" {
		return errors.New("nome é obrigatório")
	}
	if req.Configuracao.Label == "" {
		return errors.New("configuracao.label é obrigatório")
	}
	if req.Configuracao.Tipo == "" {
		return errors.New("configuracao.tipo é obrigatório")
	}

	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	var campo models.CampoFormulario
	err = database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		if err := tx.First(&models.FormularioModelo{}, "id = ?", formularioID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("Formulário não encontrado")
			}
			return err
		}

		if req.Ordem == 0 {
			var lastOrdem int
			tx.Model(&models.CampoFormulario{}).
				Where("formulario_modelo_id = ?", formularioID).
				Select("COALESCE(MAX(ordem), 0)").
				Scan(&lastOrdem)
			req.Ordem = lastOrdem + 1
		}

		var err error
		campo, err = buildCampo(formularioID, req)
		if err != nil {
			return err
		}
		return tx.Create(&campo).Error
	})
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Campo criado com sucesso",
		"campo":   campo,
	})
	return nil
}

// @Summary Atualiza um Campo do Formulário
// @Tags FormularioModelo
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "ID do Formulário"
// @Param campoId path string true "ID do Campo"
// @Router /api/v1/formularios/{id}/campos/{campoId} [put]
func UpdateCampoFormulario(c *gin.Context) error {
	var req CreateCampoInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return fmt.Errorf("body inválido: %w", err)
	}

	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	campoID, err := uuid.Parse(c.Param("campoId"))
	if err != nil {
		return errors.New("ID do campo inválido")
	}

	var campo models.CampoFormulario
	err = database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		if err := tx.First(&campo, "id = ? AND formulario_modelo_id = ?", campoID, formularioID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("Campo não encontrado")
			}
			return err
		}

		if nome := strings.TrimSpace(req.Nome); nome != "" {
			campo.Nome = nome
		}
		if req.Ordem > 0 {
			campo.Ordem = req.Ordem
		}
		if req.Configuracao.Tipo != "" || req.Configuracao.Label != "" {
			configJSON, err := json.Marshal(req.Configuracao)
			if err != nil {
				return errors.New("erro ao serializar configuração")
			}
			campo.Configuracao = configJSON
		}

		return tx.Save(&campo).Error
	})
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Campo atualizado com sucesso",
		"campo":   campo,
	})
	return nil
}

// @Summary Remove um Campo do Formulário
// @Tags FormularioModelo
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID do Formulário"
// @Param campoId path string true "ID do Campo"
// @Router /api/v1/formularios/{id}/campos/{campoId} [delete]
func DeleteCampoFormulario(c *gin.Context) error {
	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	campoID, err := uuid.Parse(c.Param("campoId"))
	if err != nil {
		return errors.New("ID do campo inválido")
	}

	return database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		if err := tx.First(&models.CampoFormulario{}, "id = ? AND formulario_modelo_id = ?", campoID, formularioID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("Campo não encontrado")
			}
			return err
		}

		if err := tx.Exec("DELETE FROM resposta_formulario WHERE campo_formulario_id = ?", campoID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM campo_formulario WHERE id = ? AND formulario_modelo_id = ?", campoID, formularioID).Error; err != nil {
			return err
		}

		c.JSON(http.StatusOK, gin.H{"message": "Campo removido com sucesso"})
		return nil
	})
}

// ─────────────────────────────────────────────────────────────
// HELPER
// ─────────────────────────────────────────────────────────────

func buildCampo(formularioID uuid.UUID, req CreateCampoInput) (models.CampoFormulario, error) {
	configJSON, err := json.Marshal(req.Configuracao)
	if err != nil {
		return models.CampoFormulario{}, errors.New("erro ao serializar configuração")
	}
	return models.CampoFormulario{
		FormularioModeloID: &formularioID,
		Nome:               req.Nome,
		Ordem:              req.Ordem,
		Configuracao:       configJSON,
	}, nil
}
