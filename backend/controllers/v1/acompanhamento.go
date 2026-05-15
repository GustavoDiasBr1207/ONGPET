package v1

import (
	"errors"
	"net/http"
	"time"

	"ongpet/database"
	"ongpet/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- INPUT TYPES ---

type CreateAcompanhamentoInput struct {
	PetID            uuid.UUID                `json:"pet_id" binding:"required"`
	AdotanteNome     string                   `json:"adotante_nome" binding:"required"`
	AdotanteTelefone string                   `json:"adotante_telefone" binding:"required"`
	AdotanteEmail    string                   `json:"adotante_email"`
	Frequencia       models.TrackingFrequency `json:"frequencia" binding:"required"`
	ProximaData      *time.Time               `json:"proxima_data"`
}

type CreateLogAcompanhamentoInput struct {
	Notas       string     `json:"notas" binding:"required"`
	DataContato time.Time  `json:"data_contato"`
	ProximaData *time.Time `json:"proxima_data"`
}

// --- HANDLERS ---

// @Summary Cria a configuração de acompanhamento para um pet adotado
// @Tags Acompanhamento
// @Accept json
// @Produce json
// @Param acompanhamento body v1.CreateAcompanhamentoInput true "Dados da adoção"
// @Success 201 {object} models.AcompanhamentoDTO
// @Security ApiKeyAuth
// @Router /api/v1/acompanhamentos [post]
func CreateAcompanhamento(c *gin.Context) error {
	var req CreateAcompanhamentoInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	var result models.Acompanhamento

	err := database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		var pet models.Pet
		if err := tx.First(&pet, "id = ?", req.PetID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("Pet não encontrado")
			}
			return err
		}

		result = models.Acompanhamento{
			PetID:            req.PetID,
			OngID:            pet.OngID,
			AdotanteNome:     req.AdotanteNome,
			AdotanteTelefone: req.AdotanteTelefone,
			AdotanteEmail:    req.AdotanteEmail,
			Frequencia:       req.Frequencia,
			ProximaData:      req.ProximaData,
			Status:           models.AcompanhamentoAtivo,
			LembreteEnviado:  false,
		}

		return tx.Create(&result).Error
	})

	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, mapAcompanhamentoToDTO(result))
	return nil
}

// @Summary Lista todos os acompanhamentos da ONG
// @Tags Acompanhamento
// @Produce json
// @Success 200 {array} models.AcompanhamentoDTO
// @Security ApiKeyAuth
// @Router /api/v1/acompanhamentos [get]
func ListAcompanhamentos(c *gin.Context) error {
	var dtos []models.AcompanhamentoDTO

	err := database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		var acompanhamentos []models.Acompanhamento
		if err := tx.Preload("Pet").Preload("Logs", func(db *gorm.DB) *gorm.DB {
			return db.Order("data_contato DESC")
		}).Order("proxima_data ASC").Find(&acompanhamentos).Error; err != nil {
			return err
		}

		dtos = make([]models.AcompanhamentoDTO, len(acompanhamentos))
		for i, a := range acompanhamentos {
			dtos[i] = mapAcompanhamentoToDTO(a)
		}
		return nil
	})

	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, dtos)
	return nil
}

// @Summary Registra um novo log de contato para um acompanhamento
// @Tags Acompanhamento
// @Accept json
// @Produce json
// @Param id path string true "ID do Acompanhamento (Master)"
// @Param log body v1.CreateLogAcompanhamentoInput true "Notas do contato"
// @Success 201 {object} models.LogAcompanhamentoDTO
// @Security ApiKeyAuth
// @Router /api/v1/acompanhamentos/{id}/logs [post]
func CreateLogAcompanhamento(c *gin.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID de acompanhamento inválido")
	}

	var req CreateLogAcompanhamentoInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	if req.DataContato.IsZero() {
		req.DataContato = time.Now()
	}

	var entry models.LogAcompanhamento

	err = database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		entry = models.LogAcompanhamento{
			AcompanhamentoID: id,
			DataContato:      req.DataContato,
			Notas:            req.Notas,
		}

		if err := tx.Create(&entry).Error; err != nil {
			return err
		}

		if req.ProximaData != nil {
			// Ao atualizar proxima_data, reseta lembrete_enviado para false
			// para que o scheduler envie novamente para a nova data
			return tx.Model(&models.Acompanhamento{}).
				Where("id = ?", id).
				Updates(map[string]interface{}{
					"proxima_data":     req.ProximaData,
					"lembrete_enviado": false,
				}).Error
		}

		return nil
	})

	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, mapLogToDTO(entry))
	return nil
}

// @Summary Busca o histórico de logs de um acompanhamento
// @Tags Acompanhamento
// @Produce json
// @Param id path string true "ID do Acompanhamento"
// @Success 200 {array} models.LogAcompanhamentoDTO
// @Security ApiKeyAuth
// @Router /api/v1/acompanhamentos/{id}/logs [get]
func GetAcompanhamentoLogs(c *gin.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID de acompanhamento inválido")
	}

	var dtos []models.LogAcompanhamentoDTO

	err = database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		var logs []models.LogAcompanhamento
		if err := tx.Where("acompanhamento_id = ?", id).Order("data_contato DESC").Find(&logs).Error; err != nil {
			return err
		}

		dtos = make([]models.LogAcompanhamentoDTO, len(logs))
		for i, l := range logs {
			dtos[i] = mapLogToDTO(l)
		}
		return nil
	})

	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, dtos)
	return nil
}

// --- HELPERS ---

func mapAcompanhamentoToDTO(a models.Acompanhamento) models.AcompanhamentoDTO {
	logsDTO := make([]models.LogAcompanhamentoDTO, len(a.Logs))
	for i, l := range a.Logs {
		logsDTO[i] = mapLogToDTO(l)
	}

	dto := models.AcompanhamentoDTO{
		ID:               a.ID,
		PetID:            a.PetID,
		AdotanteNome:     a.AdotanteNome,
		AdotanteTelefone: a.AdotanteTelefone,
		AdotanteEmail:    a.AdotanteEmail,
		Frequencia:       a.Frequencia,
		ProximaData:      a.ProximaData,
		Status:           a.Status,
		CreatedAt:        a.CreatedAt,
		Logs:             logsDTO,
	}
	if a.Pet.ID != uuid.Nil {
		dto.PetNome = a.Pet.Nome
	}
	return dto
}

func mapLogToDTO(l models.LogAcompanhamento) models.LogAcompanhamentoDTO {
	return models.LogAcompanhamentoDTO{
		ID:               l.ID,
		AcompanhamentoID: l.AcompanhamentoID,
		DataContato:      l.DataContato,
		Notas:            l.Notas,
		CreatedAt:        l.CreatedAt,
	}
}