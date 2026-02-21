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
	"github.com/lib/pq" // 👈 FALTAVA ISSO
	"gorm.io/gorm"
)

type CreatePetInput struct {
    Nome         string     `json:"nome"`
    Especie      string     `json:"especie"`
    Raca         string     `json:"raca"`
    Idade        int        `json:"idade"`
    Descricao    string     `json:"descricao"`
    Peso         float64    `json:"peso"`
    Porte        string     `json:"porte"`
    Regiao       string     `json:"regiao"`
    Imagens      []string   `json:"imagens"`
    FormularioID *uuid.UUID `json:"formulario_id"` // opcional
    OngID        uuid.UUID  `json:"ong_id"`        // obrigatório
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

	// filtros
	if nome := strings.TrimSpace(c.Query("nome")); nome != "" {
		query = query.Where("nome ILIKE ?", "%"+nome+"%")
	}

	if especie := strings.TrimSpace(c.Query("especie")); especie != "" {
		query = query.Where("especie ILIKE ?", "%"+especie+"%")
	}

	if porte := strings.TrimSpace(c.Query("porte")); porte != "" {
		query = query.Where("porte = ?", porte)
	}

	if regiao := strings.TrimSpace(c.Query("regiao")); regiao != "" {
		query = query.Where("regiao ILIKE ?", "%"+regiao+"%")
	}

	if ongID := strings.TrimSpace(c.Query("ong_id")); ongID != "" {
		query = query.Where("ong_id = ?", ongID)
	}

	// paginação
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
    if req.Idade <= 0 {
        return errors.New("idade inválida")
    }
    if req.Peso <= 0 {
        return errors.New("peso inválido")
    }
    if req.OngID == uuid.Nil {
        return errors.New("ong_id é obrigatório")
    }
    if len(req.Imagens) > 5 {
        return errors.New("máximo de 5 imagens")
    }

    db := database.GetDB()

    if err := db.First(&models.Ong{}, "id = ?", req.OngID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errors.New("ONG não encontrada")
        }
        return err
    }

    pet := models.Pet{
        Nome:      req.Nome,
        Especie:   req.Especie,
        Raca:      req.Raca,
        Idade:     req.Idade,
        Descricao: req.Descricao,
        Peso:      req.Peso,
        Porte:     req.Porte,
        Regiao:    req.Regiao,
        Imagens:   pq.StringArray(req.Imagens),
        OngID:     req.OngID,
    }

    // ✅ CORRETO: verifica se o ponteiro não é nil e se não é uuid.Nil
    if req.FormularioID != nil && *req.FormularioID != uuid.Nil {
        pet.FormularioID = req.FormularioID
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

    // update parcial
    if nome := strings.TrimSpace(req.Nome); nome != "" {
        pet.Nome = nome
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
    if req.Descricao != "" {
        pet.Descricao = req.Descricao
    }
    if req.Peso > 0 {
        pet.Peso = req.Peso
    }
    if req.Porte != "" {
        pet.Porte = req.Porte
    }
    if req.Regiao != "" {
        pet.Regiao = req.Regiao
    }

    // permite atualizar ou limpar imagens
    if req.Imagens != nil {
        if len(req.Imagens) > 5 {
            return errors.New("máximo de 5 imagens")
        }
        pet.Imagens = req.Imagens
    }

    // permite atualizar formulario_id e ong_id de forma segura
    if req.FormularioID != nil && *req.FormularioID != uuid.Nil {
        pet.FormularioID = req.FormularioID
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

	idParam := c.Param("id")
	petID, err := uuid.Parse(idParam)
	if err != nil {
		return errors.New("ID do pet inválido")
	}

	result := db.Where("id = ?", petID).Delete(&models.Pet{})
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
