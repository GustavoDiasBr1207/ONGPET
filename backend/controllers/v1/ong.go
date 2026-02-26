package v1

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"ongpet/database"
	"ongpet/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateOngInput representa o payload mínimo para criar uma ONG
type CreateOngInput struct {
	Email           string `json:"email" example:"a@b.com"`
	Endereco        string `json:"endereco" example:"Rua A"`
	Nome            string `json:"nome" example:"Minha ONG"`
	NomeResponsavel string `json:"nome_responsavel" example:"Fulano"`
	Telefone        string `json:"telefone" example:"123"`
}

type OngListResponse struct {
	Dados          []models.Ong `json:"dados"`
	TotalRegistros int64        `json:"total_registros"`
	TotalPaginas   int          `json:"total_paginas"`
	ProximaPagina  bool         `json:"proxima_pagina"`
}

// @Summary Lista todas as ONGs
// @Description Retorna todas as ONGs cadastradas
// @Tags ONG
// @Produce json
// @Success 200 {object} v1.OngListResponse
// @Router /api/v1/ongs [get]
func ReadOngs(c *gin.Context) error {
	db := database.GetDB()
	query := db.Model(&models.Ong{})

	if nome := c.Query("nome"); nome != "" {
		query = query.Where("nome ILIKE ?", "%"+nome+"%")
	}

	if email := c.Query("email"); email != "" {
		query = query.Where("email ILIKE ?", "%"+email+"%")
	}

	pageStr := c.DefaultQuery("page", "1")
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

	offset := (page - 1) * limit
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	var ongs []models.Ong
	if err := query.
		Offset(offset).
		Limit(limit + 1).
		Find(&ongs).Error; err != nil {
		return err
	}

	hasNext := false
	if len(ongs) > limit {
		hasNext = true
		ongs = ongs[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"dados":           ongs,
		"total_registros": total,
		"total_paginas":   totalPages,
		"proxima_pagina":  hasNext,
	})

	return nil
}

// @Summary Cria uma nova ONG
// @Description Cria uma ONG no sistema
// @Tags ONG
// @Accept json
// @Produce json
// @Param ong body v1.CreateOngInput true "Nova ONG"
// @Success 201 {object} models.Ong
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/v1/ongs [post]
func CreateOng(c *gin.Context) error {
	var req CreateOngInput

	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	req.Nome = strings.TrimSpace(req.Nome)
	req.Email = strings.TrimSpace(req.Email)

	db := database.GetDB()

	// Verifica duplicidade de email
	if err := db.Where("email = ?", req.Email).
		First(&models.Ong{}).Error; err == nil {
		return gin.Error{
			Err:  errors.New("email já cadastrado"),
			Type: gin.ErrorTypePublic,
		}
	}

	ong := models.Ong{
		Nome:            req.Nome,
		Email:           req.Email,
		Endereco:        req.Endereco,
		Telefone:        req.Telefone,
		NomeResponsavel: req.NomeResponsavel,
	}

	if err := db.Create(&ong).Error; err != nil {
		return err
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "ONG criada com sucesso",
		"ong":     ong,
	})

	return nil
}

// @Summary Atualiza uma ONG existente
// @Description Atualiza os dados de uma ONG pelo ID
// @Tags ONG
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "ID da ONG"
// @Param ong body v1.CreateOngInput true "Dados para atualização"
// @Success 200 {object} object{message=string,ong=models.Ong}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ongs/{id} [put]
func UpdateOng(c *gin.Context) error {
	var req CreateOngInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	db := database.GetDB()
	id := c.Param("id")

	var ong models.Ong
	if err := db.Where("id = ?", id).First(&ong).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("ONG não encontrada")
		}
		return err
	}

	// Atualiza apenas campos enviados
	if strings.TrimSpace(req.Nome) != "" {
		ong.Nome = strings.TrimSpace(req.Nome)
	}
	if strings.TrimSpace(req.Email) != "" {
		ong.Email = strings.TrimSpace(req.Email)
	}
	if req.Endereco != "" {
		ong.Endereco = req.Endereco
	}
	if req.Telefone != "" {
		ong.Telefone = req.Telefone
	}
	if req.NomeResponsavel != "" {
		ong.NomeResponsavel = req.NomeResponsavel
	}

	if err := db.Save(&ong).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ONG atualizada com sucesso",
		"ong":     ong,
	})

	return nil
}

// @Summary Remove uma ONG
// @Description Remove uma ONG pelo ID
// @Tags ONG
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID da ONG"
// @Success 200 {object} object{message=string}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ongs/{id} [delete]
func DeleteOng(c *gin.Context) error {
	db := database.GetDB()
	id := c.Param("id")

	result := db.Where("id = ?", id).Delete(&models.Ong{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("ONG não encontrada")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ONG removida com sucesso",
	})

	return nil
}
