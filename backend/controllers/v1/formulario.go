package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"ongpet/database"
	"ongpet/models"

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

// ─────────────────────────────────────────────────────────────
// FORMULARIO MODELO — CRUD
// ─────────────────────────────────────────────────────────────

// @Summary Lista todos os Formulários Modelo
// @Tags FormularioModelo
// @Security ApiKeyAuth
// @Produce json
// @Param ong_id query string false "Filtrar por ONG"
// @Success 200 {array} models.FormularioModelo
// @Router /api/v1/formularios [get]
func ReadFormularios(c *gin.Context) error {
	db := database.GetDB()
	query := db.Model(&models.FormularioModelo{})

	if ongID := strings.TrimSpace(c.Query("ong_id")); ongID != "" {
		query = query.Where("ong_id = ?", ongID)
	}

	var formularios []models.FormularioModelo
	if err := query.
		Preload("Campos", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("ordem ASC")
		}).
		Find(&formularios).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, formularios)
	return nil
}

// @Summary Busca um Formulário Modelo pelo ID
// @Tags FormularioModelo
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "ID do Formulário"
// @Success 200 {object} models.FormularioModelo
// @Failure 404 {object} map[string]string
// @Router /api/v1/formularios/{id} [get]
func ReadFormulario(c *gin.Context) error {
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
// @Description Cria um formulário com campos configuráveis para uso na adoção de pets
// @Tags FormularioModelo
// @Accept json
// @Produce json
// @Param formulario body v1.CreateFormularioModeloInput true "Novo Formulário"
// @Success 201 {object} models.FormularioModelo
// @Security ApiKeyAuth
// @Router /api/v1/formularios [post]
func CreateFormulario(c *gin.Context) error {
	var req CreateFormularioModeloInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	req.Nome = strings.TrimSpace(req.Nome)
	if req.Nome == "" {
		return errors.New("nome é obrigatório")
	}
	if req.OngID == uuid.Nil {
		return errors.New("ong_id é obrigatório")
	}

	db := database.GetDB()

	if err := db.First(&models.Ong{}, "id = ?", req.OngID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("ONG não encontrada")
		}
		return err
	}

	formulario := models.FormularioModelo{
		OngID: req.OngID,
		Nome:  req.Nome,
	}

	if err := db.Create(&formulario).Error; err != nil {
		return err
	}

	for i, campoInput := range req.Campos {
		if campoInput.Ordem == 0 {
			campoInput.Ordem = i + 1
		}
		campo, err := buildCampo(formulario.ID, campoInput)
		if err != nil {
			return err
		}
		if err := db.Create(&campo).Error; err != nil {
			return err
		}
	}

	// Recarrega com campos ordenados
	if err := db.Preload("Campos", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("ordem ASC")
	}).First(&formulario, "id = ?", formulario.ID).Error; err != nil {
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
// @Param id path string true "ID do Formulário"
// @Param formulario body v1.UpdateFormularioModeloInput true "Dados para atualização"
// @Success 200 {object} models.FormularioModelo
// @Security ApiKeyAuth
// @Router /api/v1/formularios/{id} [put]
func UpdateFormulario(c *gin.Context) error {
	var req UpdateFormularioModeloInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	db := database.GetDB()

	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	var formulario models.FormularioModelo
	if err := db.First(&formulario, "id = ?", formularioID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Formulário não encontrado")
		}
		return err
	}

	if nome := strings.TrimSpace(req.Nome); nome != "" {
		formulario.Nome = nome
	}

	if err := db.Save(&formulario).Error; err != nil {
		return err
	}

	if err := db.Preload("Campos", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("ordem ASC")
	}).First(&formulario, "id = ?", formularioID).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Formulário atualizado com sucesso",
		"formulario": formulario,
	})
	return nil
}

// @Summary Remove um Formulário Modelo
// @Description Remove o formulário e todos os campos em cascata
// @Tags FormularioModelo
// @Produce json
// @Param id path string true "ID do Formulário"
// @Success 200 {object} object{message=string}
// @Security ApiKeyAuth
// @Router /api/v1/formularios/{id} [delete]
func DeleteFormulario(c *gin.Context) error {
	db := database.GetDB()

	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	if err := db.First(&models.FormularioModelo{}, "id = ?", formularioID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Formulário não encontrado")
		}
		return err
	}

	// Soft delete em cascata: respostas → campos → formulário
	if err := db.Where("campo_formulario_id IN (?)",
		db.Model(&models.CampoFormulario{}).
			Select("id").
			Where("formulario_modelo_id = ?", formularioID),
	).Delete(&models.RespostaFormulario{}).Error; err != nil {
		return err
	}

	if err := db.Where("formulario_modelo_id = ?", formularioID).
		Delete(&models.CampoFormulario{}).Error; err != nil {
		return err
	}

	if err := db.Where("id = ?", formularioID).
		Delete(&models.FormularioModelo{}).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "Formulário e campos removidos com sucesso"})
	return nil
}

// ─────────────────────────────────────────────────────────────
// CAMPOS — CRUD
// ─────────────────────────────────────────────────────────────

// @Summary Adiciona um Campo ao Formulário
// @Tags FormularioModelo
// @Accept json
// @Produce json
// @Param id path string true "ID do Formulário"
// @Param campo body v1.CreateCampoInput true "Novo Campo"
// @Success 201 {object} models.CampoFormulario
// @Security ApiKeyAuth
// @Router /api/v1/formularios/{id}/campos [post]
func CreateCampoFormulario(c *gin.Context) error {
	var req CreateCampoInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
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

	db := database.GetDB()

	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	if err := db.First(&models.FormularioModelo{}, "id = ?", formularioID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Formulário não encontrado")
		}
		return err
	}

	// Auto ordem se não informada
	if req.Ordem == 0 {
		var lastOrdem int
		db.Model(&models.CampoFormulario{}).
			Where("formulario_modelo_id = ?", formularioID).
			Select("COALESCE(MAX(ordem), 0)").
			Scan(&lastOrdem)
		req.Ordem = lastOrdem + 1
	}

	campo, err := buildCampo(formularioID, req)
	if err != nil {
		return err
	}

	if err := db.Create(&campo).Error; err != nil {
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
// @Param id path string true "ID do Formulário"
// @Param campoId path string true "ID do Campo"
// @Param campo body v1.CreateCampoInput true "Dados para atualização"
// @Success 200 {object} models.CampoFormulario
// @Security ApiKeyAuth
// @Router /api/v1/formularios/{id}/campos/{campoId} [put]
func UpdateCampoFormulario(c *gin.Context) error {
	var req CreateCampoInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	db := database.GetDB()

	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	campoID, err := uuid.Parse(c.Param("campoId"))
	if err != nil {
		return errors.New("ID do campo inválido")
	}

	var campo models.CampoFormulario
	if err := db.First(&campo, "id = ? AND formulario_modelo_id = ?", campoID, formularioID).Error; err != nil {
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

	if err := db.Save(&campo).Error; err != nil {
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
// @Produce json
// @Param id path string true "ID do Formulário"
// @Param campoId path string true "ID do Campo"
// @Success 200 {object} object{message=string}
// @Security ApiKeyAuth
// @Router /api/v1/formularios/{id}/campos/{campoId} [delete]
func DeleteCampoFormulario(c *gin.Context) error {
	db := database.GetDB()

	formularioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do formulário inválido")
	}

	campoID, err := uuid.Parse(c.Param("campoId"))
	if err != nil {
		return errors.New("ID do campo inválido")
	}

	if err := db.First(&models.CampoFormulario{}, "id = ? AND formulario_modelo_id = ?", campoID, formularioID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Campo não encontrado")
		}
		return err
	}

	if err := db.Exec("DELETE FROM resposta_formulario WHERE campo_formulario_id = ?", campoID).Error; err != nil {
		return err
	}

	if err := db.Exec("DELETE FROM campo_formulario WHERE id = ? AND formulario_modelo_id = ?", campoID, formularioID).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "Campo removido com sucesso"})
	return nil
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
		FormularioModeloID: &formularioID, // ponteiro
		Nome:               req.Nome,
		Ordem:              req.Ordem,
		Configuracao:       configJSON,
	}, nil
}