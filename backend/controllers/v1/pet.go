package v0

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"ongpet/database"
	"ongpet/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreatePetInput representa o payload para criação/atualização de Pet
type CreatePetInput struct {
	Nome    string    `json:"nome" example:"Rex"`
	Especie string    `json:"especie" example:"Cachorro"`
	Raca    string    `json:"raca" example:"Vira-lata"`
	Idade   int       `json:"idade" example:"3"`
	OngID   uuid.UUID `json:"ong_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type PetListResponse struct {
	Dados          []models.Pet `json:"dados"`
	TotalRegistros int64        `json:"total_registros"`
	TotalPaginas   int          `json:"total_paginas"`
	ProximaPagina  bool         `json:"proxima_pagina"`
}

// @Summary Lista todos os Pets
// @Description Retorna todos os pets cadastrados
// @Tags Pet
// @Produce json
// @Success 200 {object} v0.PetListResponse
// @Router /api/v1/pets [get]
func ReadPets(c *gin.Context) error {
	db := database.GetDB()
	query := db.Model(&models.Pet{})

	if nome := c.Query("nome"); nome != "" {
		query = query.Where("nome ILIKE ?", "%"+nome+"%")
	}

	if especie := c.Query("especie"); especie != "" {
		query = query.Where("especie ILIKE ?", "%"+especie+"%")
	}

	if ongID := c.Query("ong_id"); ongID != "" {
		query = query.Where("ong_id = ?", ongID)
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

	var pets []models.Pet
	if err := query.
		Offset(offset).
		Limit(limit + 1).
		Find(&pets).Error; err != nil {
		return err
	}

	hasNext := false
	if len(pets) > limit {
		hasNext = true
		pets = pets[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"dados":           pets,
		"total_registros": total,
		"total_paginas":   totalPages,
		"proxima_pagina":  hasNext,
	})

	return nil
}

// @Summary Cria um novo Pet
// @Description Cria um Pet no sistema
// @Tags Pet
// @Accept json
// @Produce json
// @Param pet body v0.CreatePetInput true "Novo Pet"
// @Success 201 {object} models.Pet
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/v1/pets [post]
func CreatePet(c *gin.Context) error {
	var req CreatePetInput

	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	req.Nome = strings.TrimSpace(req.Nome)

	if req.Nome == "" {
		return errors.New("nome é obrigatório")
	}

	db := database.GetDB()

	// valida se a ONG existe
	if err := db.Where("id = ?", req.OngID).First(&models.Ong{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("ONG não encontrada")
		}
		return err
	}

	pet := models.Pet{
		Nome:    req.Nome,
		Especie: req.Especie,
		Raca:    req.Raca,
		Idade:   req.Idade,
		OngID:   req.OngID,
	}

	if err := db.Create(&pet).Error; err != nil {
		return err
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pet criado com sucesso",
		"pet":     pet,
	})

	return nil
}

// @Summary Atualiza um Pet existente
// @Description Atualiza os dados de um Pet pelo ID
// @Tags Pet
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "ID do Pet"
// @Param pet body v0.CreatePetInput true "Dados para atualização"
// @Success 200 {object} object{message=string,pet=models.Pet}
// @Failure 404 {object} map[string]string
// @Router /api/v1/pets/{id} [put]
func UpdatePet(c *gin.Context) error {
	var req CreatePetInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	db := database.GetDB()
	id := c.Param("id")

	var pet models.Pet
	if err := db.Where("id = ?", id).First(&pet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Pet não encontrado")
		}
		return err
	}

	if strings.TrimSpace(req.Nome) != "" {
		pet.Nome = strings.TrimSpace(req.Nome)
	}
	if req.Especie != "" {
		pet.Especie = req.Especie
	}
	if req.Raca != "" {
		pet.Raca = req.Raca
	}
	if req.Idade > 0 {
		pet.Idade = req.Idade
	}
	if req.OngID != uuid.Nil {
		pet.OngID = req.OngID
	}

	if err := db.Save(&pet).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pet atualizado com sucesso",
		"pet":     pet,
	})

	return nil
}

// @Summary Remove um Pet
// @Description Remove um Pet pelo ID
// @Tags Pet
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID do Pet"
// @Success 200 {object} object{message=string}
// @Failure 404 {object} map[string]string
// @Router /api/v1/pets/{id} [delete]
func DeletePet(c *gin.Context) error {
	db := database.GetDB()
	id := c.Param("id")

	result := db.Where("id = ?", id).Delete(&models.Pet{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("Pet não encontrado")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pet removido com sucesso",
	})

	return nil
}
