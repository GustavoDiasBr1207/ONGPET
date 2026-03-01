package v1

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"ongpet/database"
	"ongpet/models"
    "ongpet/utils"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
    FormularioID *uuid.UUID `json:"formulario_id"`
    OngID        uuid.UUID  `json:"ong_id"`
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
// @Success 200 {object} v1.PetListResponse
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
		Preload("Imagens", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("position ASC")
		}).
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

// @Summary Busca um Pet pelo ID
// @Description Retorna um Pet específico pelo ID
// @Tags Pet
// @Produce json
// @Param id path string true "ID do Pet"
// @Success 200 {object} models.Pet
// @Failure 404 {object} map[string]string
// @Router /api/v1/pets/{id} [get]
func ReadPet(c *gin.Context) error {
	db := database.GetDB()

	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do pet inválido")
	}

	var pet models.Pet
	if err := db.
		Preload("Imagens", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("position ASC")
		}).
		First(&pet, "id = ?", petID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Pet não encontrado")
		}
		return err
	}

	c.JSON(http.StatusOK, pet)
	return nil
}

// @Summary Cria um novo Pet
// @Description Cria um Pet no sistema
// @Tags Pet
// @Accept json
// @Produce json
// @Param pet body v1.CreatePetInput true "Novo Pet"
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

		pet := models.Pet{
		Nome:      req.Nome,
		Especie:   req.Especie,
		Raca:      req.Raca,
		Idade:     req.Idade,
		Descricao: req.Descricao,
		Peso:      req.Peso,
		Porte:     req.Porte,
		Regiao:    req.Regiao,
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

// @Summary Adiciona imagens a um Pet existente
// @Description Faz upload de até 5 imagens para o Pet pelo ID (usando Supabase)
// @Tags Pet
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "ID do Pet"
// @Param imagens formData file true "Imagens do Pet (até 5)"
// @Success 200 {object} object{message=string,pet=models.Pet}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/pets/{id}/imagens [post]
func UploadPetImages(c *gin.Context) error {
	db := database.GetDB()

	// 🔎 PET ID
	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do pet inválido")
	}

	// 🔎 BUSCA PET
	var pet models.Pet
	if err := db.First(&pet, "id = ?", petID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Pet não encontrado")
		}
		return err
	}

	// 📸 MULTIPART
	form, err := c.MultipartForm()
	if err != nil || form.File == nil {
		return errors.New("nenhuma imagem enviada")
	}

	files := form.File["imagens"]
	if len(files) == 0 {
		return errors.New("nenhuma imagem enviada")
	}

	// 🔢 CONTA IMAGENS ATUAIS
	var totalAtual int64
	if err := db.Model(&models.PetImage{}).
		Where("pet_id = ?", petID).
		Count(&totalAtual).Error; err != nil {
		return err
	}

	if totalAtual+int64(len(files)) > 5 {
		return errors.New("o pet pode ter no máximo 5 imagens")
	}

	// 🔢 ÚLTIMA POSIÇÃO
	var lastPosition int
	if err := db.Model(&models.PetImage{}).
		Where("pet_id = ?", petID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&lastPosition).Error; err != nil {
		return err
	}

	// 🔁 UPLOAD + CREATE
	for i, file := range files {
		if !utils.IsValidImage(file) {
			return errors.New("tipo de imagem inválido")
		}

		position := lastPosition + i + 1

		url, err := utils.UploadFile(
			file,
			petID.String(),
			pet.Nome,
			position,
		)
		if err != nil {
			return err
		}

		image := models.PetImage{
			URL:      url,
			PetID:    petID,
			Position: position,
		}

		if err := db.Create(&image).Error; err != nil {
			return err
		}
	}

	// 🔄 RECARREGA PET COM IMAGENS ORDENADAS
	if err := db.Preload("Imagens", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("position ASC")
	}).First(&pet, "id = ?", petID).Error; err != nil {
		return err
	}

	// 📤 RESPONSE
	c.JSON(http.StatusOK, gin.H{
		"message": "Imagens adicionadas com sucesso",
		"pet":     pet,
	})

	return nil
}

// @Summary Atualiza um Pet existente
// @Description Atualiza os dados de um Pet pelo ID (não atualiza imagens)
// @Tags Pet
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "ID do Pet"
// @Param pet body v1.CreatePetInput true "Dados para atualização"
// @Success 200 {object} object{message=string,pet=models.Pet}
// @Failure 404 {object} map[string]string
// @Router /api/v1/pets/{id} [put]
func UpdatePet(c *gin.Context) error {
	var req CreatePetInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	db := database.GetDB()

	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do pet inválido")
	}

	var pet models.Pet
	if err := db.First(&pet, "id = ?", petID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Pet não encontrado")
		}
		return err
	}

	// 🔄 UPDATE PARCIAL (somente campos enviados)
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

	// 🔐 Atualiza FormularioID com segurança
	if req.FormularioID != nil {
		if *req.FormularioID == uuid.Nil {
			pet.FormularioID = nil // permite limpar
		} else {
			pet.FormularioID = req.FormularioID
		}
	}

	// 🔐 Atualiza ONG
	if req.OngID != uuid.Nil {
		pet.OngID = req.OngID
	}

	// 💾 SALVA
	if err := db.Save(&pet).Error; err != nil {
		return err
	}

	// 🔄 Recarrega com imagens ordenadas (para o frontend)
	if err := db.Preload("Imagens", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("position ASC")
	}).First(&pet, "id = ?", petID).Error; err != nil {
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

// @Summary Remove uma imagem de um Pet
// @Description Remove uma imagem específica de um Pet pelo ID
// @Tags Pet
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID do Pet"
// @Param imageId path string true "ID da Imagem"
// @Success 200 {object} object{message=string,pet=models.Pet}
// @Failure 404 {object} map[string]string
// @Router /api/v1/pets/{id}/imagens/{imageId} [delete]
func DeletePetImage(c *gin.Context) error {
	db := database.GetDB()

	// 🔎 VALIDA IDs
	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do pet inválido")
	}

	imageID, err := uuid.Parse(c.Param("imageId"))
	if err != nil {
		return errors.New("ID da imagem inválido")
	}

	// 🔍 BUSCA IMAGEM NO BANCO
	var image models.PetImage
	if err := db.First(&image, "id = ? AND pet_id = ?", imageID, petID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Imagem não encontrada")
		}
		return err
	}

	// 🗑️ EXTRAI PATH RELATIVO E DELETA DO SUPABASE
	objectPath, pathErr := utils.ExtractObjectPath(image.URL)
	if pathErr == nil {
		if err := utils.DeleteFile(objectPath); err != nil {
			// Loga o erro mas continua (arquivo pode já estar deletado)
			fmt.Printf("⚠️ Aviso: não foi possível deletar arquivo do storage: %s\n", err.Error())
		}
	} else {
		fmt.Printf("⚠️ Aviso: não foi possível extrair path do storage: %s\n", pathErr.Error())
	}

	// 💾 DELETA DO BANCO
	if err := db.Delete(&image).Error; err != nil {
		return err
	}

	// 🔄 RETORNA PET ATUALIZADO COM IMAGENS ORDENADAS
	var pet models.Pet
	if err := db.Preload("Imagens", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("position ASC")
	}).First(&pet, "id = ?", petID).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Imagem removida com sucesso",
		"pet":     pet,
	})

	return nil
}