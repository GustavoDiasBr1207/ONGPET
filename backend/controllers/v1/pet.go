package v1

import (
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

type CreatePetInput struct {
	Nome         string            `json:"nome"          example:"Rex"`
	Especie      models.PetEspecie `json:"especie"       example:"Cachorro"`
	Raca         string            `json:"raca"          example:"Vira-lata"`
	Idade        int               `json:"idade"         example:"1"`
	Descricao    string            `json:"descricao"     example:"Pet dócil e brincalhão"`
	Peso         float64           `json:"peso"          example:"5.5"`
	Porte        models.PetPorte   `json:"porte"         example:"Médio"`
	Regiao       models.PetRegiao  `json:"regiao"        example:"Serra"`
	FormularioID *string           `json:"formulario_id" example:""`
	OngID        uuid.UUID         `json:"ong_id"        example:"6e2c2e7d-4c27-4570-b32e-72799a9e059e"`
	Sexo         models.PetSexo    `json:"sexo"          example:"Macho"`
	Status       models.PetStatus  `json:"status"        example:"draft"`
}

type PetListResponse struct {
	Dados          []models.Pet `json:"dados"`
	TotalRegistros int64        `json:"total_registros"`
	TotalPaginas   int          `json:"total_paginas"`
	ProximaPagina  bool         `json:"proxima_pagina"`
}

// parseFormularioID converte *string para *uuid.UUID ignorando nil e string vazia
func parseFormularioID(s *string) (*uuid.UUID, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(*s)
	if err != nil {
		return nil, errors.New("formulario_id inválido")
	}
	return &parsed, nil
}

// @Summary Lista todos os Pets
// @Description Retorna lista paginada de pets com filtros opcionais
// @Tags Pet
// @Produce json
// @Param nome    query string false "Filtrar por nome"
// @Param especie query string false "Filtrar por espécie"
// @Param porte   query string false "Filtrar por porte"
// @Param regiao  query string false "Filtrar por região"
// @Param ong_id  query string false "Filtrar por ONG"
// @Param page    query int    false "Página (default: 1)"
// @Param limit   query int    false "Itens por página (default: 10)"
// @Success 200 {object} v1.PetListResponse
// @Router /api/v1/pets [get]
func ReadPets(c *gin.Context) error {
	db := database.GetDB()
	query := db.Model(&models.Pet{})

	if nome := strings.TrimSpace(c.Query("nome")); nome != "" {
		query = query.Where("nome LIKE ?", "%"+nome+"%")
	}
	if especie := strings.TrimSpace(c.Query("especie")); especie != "" {
		query = query.Where("especie LIKE ?", "%"+especie+"%")
	}
	if porte := strings.TrimSpace(c.Query("porte")); porte != "" {
		query = query.Where("porte = ?", porte)
	}
	if regiao := strings.TrimSpace(c.Query("regiao")); regiao != "" {
		query = query.Where("regiao LIKE ?", "%"+regiao+"%")
	}
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
// @Description Retorna um pet específico pelo ID
// @Tags Pet
// @Produce json
// @Param id path string true "ID do Pet"
// @Success 200 {object} models.Pet
// @Failure 400 {object} map[string]string
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
// @Description Cria um novo pet vinculado a uma ONG
// @Tags Pet
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param pet body v1.CreatePetInput true "Dados do Pet"
// @Success 201 {object} object{message=string,pet=models.Pet}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/pets [post]
func CreatePet(c *gin.Context) error {
	var req CreatePetInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return fmt.Errorf("body inválido: %w", err)
	}

	req.Nome = strings.TrimSpace(req.Nome)

	if req.OngID == uuid.Nil {
		return errors.New("ong_id é obrigatório")
	}

	formularioID, err := parseFormularioID(req.FormularioID)
	if err != nil {
		return err
	}

	var pet models.Pet
	err = database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		if err := tx.First(&models.Ong{}, "id = ?", req.OngID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("ONG não encontrada")
			}
			return err
		}

		pet = models.Pet{
			Nome:         req.Nome,
			Especie:      req.Especie,
			Raca:         req.Raca,
			Idade:        req.Idade,
			Descricao:    req.Descricao,
			Peso:         req.Peso,
			Porte:        req.Porte,
			Regiao:       req.Regiao,
			Sexo:         req.Sexo,
			OngID:        req.OngID,
			Status:       req.Status,
			FormularioID: formularioID,
		}
		return tx.Create(&pet).Error
	})
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pet criado com sucesso",
		"pet":     pet,
	})
	return nil
}

// @Summary Adiciona imagens a um Pet
// @Description Envia até 5 imagens para um pet (máximo 5 no total)
// @Tags Pet
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param id      path     string true "ID do Pet"
// @Param imagens formData file   true "Imagens do Pet"
// @Success 200 {object} object{message=string,pet=models.Pet}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/pets/{id}/imagens [post]
func UploadPetImages(c *gin.Context) error {
	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do pet inválido")
	}

	token := c.GetString("token")

	// Etapa 1: valida que o pet existe (RLS)
	var pet models.Pet
	err = database.WithUserDB(token, func(tx *gorm.DB) error {
		if err := tx.First(&pet, "id = ?", petID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("Pet não encontrado")
			}
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Etapa 2: valida os arquivos enviados
	form, err := c.MultipartForm()
	if err != nil || form.File == nil {
		return errors.New("nenhuma imagem enviada")
	}
	files := form.File["imagens"]
	if len(files) == 0 {
		return errors.New("nenhuma imagem enviada")
	}

	// Etapa 3: verifica limite de imagens e posição atual (RLS)
	var (
		lastPosition int
		totalAtual   int64
	)
	err = database.WithUserDB(token, func(tx *gorm.DB) error {
		if err := tx.Model(&models.PetImage{}).Where("pet_id = ?", petID).Count(&totalAtual).Error; err != nil {
			return err
		}
		if totalAtual+int64(len(files)) > 5 {
			return errors.New("o pet pode ter no máximo 5 imagens")
		}
		return tx.Model(&models.PetImage{}).
			Where("pet_id = ?", petID).
			Select("COALESCE(MAX(position), 0)").
			Scan(&lastPosition).Error
	})
	if err != nil {
		return err
	}

	// Etapa 2: processa e faz upload das imagens
	type uploadedImage struct {
		url      string
		position int
	}
	uploads := make([]uploadedImage, 0, len(files))

	for i, file := range files {
		if !utils.IsValidImage(file) {
			return errors.New("tipo de imagem inválido")
		}
		position := lastPosition + i + 1

		optimized, err := utils.ResizeAndCompressImage(file, 1280, 70)
		if err != nil {
			return err
		}
		url, err := utils.UploadOptimizedFile(
			optimized.Buffer,
			optimized.ContentType,
			optimized.Extension,
			utils.SupabaseBucketPets,
			petID.String(),
			pet.Nome,
			position,
			c.GetString("user_token"),
		)
		if err != nil {
			return err
		}
		uploads = append(uploads, uploadedImage{url: url, position: position})
	}

	// Etapa 3: salva registros de imagem e recarrega pet (RLS)
	err = database.WithUserDB(token, func(tx *gorm.DB) error {
		for _, u := range uploads {
			image := models.PetImage{URL: u.url, PetID: petID, Position: u.position}
			if err := tx.Create(&image).Error; err != nil {
				return err
			}
		}
		return tx.Preload("Imagens", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("position ASC")
		}).First(&pet, "id = ?", petID).Error
	})
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Imagens adicionadas com sucesso",
		"pet":     pet,
	})
	return nil
}

// @Summary Atualiza um Pet existente
// @Description Atualiza os dados de um pet pelo ID
// @Tags Pet
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id  path string           true "ID do Pet"
// @Param pet body v1.CreatePetInput true "Dados para atualização"
// @Success 200 {object} object{message=string,pet=models.Pet}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/pets/{id} [put]
func UpdatePet(c *gin.Context) error {
	var req CreatePetInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return fmt.Errorf("body inválido: %w", err)
	}

	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do pet inválido")
	}

	formularioID, err := parseFormularioID(req.FormularioID)
	if err != nil {
		return err
	}

	var pet models.Pet
	err = database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		if err := tx.First(&pet, "id = ?", petID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("Pet não encontrado")
			}
			return err
		}

		if nome := strings.TrimSpace(req.Nome); nome != "" {
			pet.Nome = nome
		}
		if req.Especie != "" {
			pet.Especie = models.PetEspecie(strings.TrimSpace(string(req.Especie)))
		}
		if req.Raca != "" {
			pet.Raca = strings.TrimSpace(req.Raca)
		}
		pet.Idade = req.Idade
		if req.Descricao != "" {
			pet.Descricao = strings.TrimSpace(req.Descricao)
		}
		pet.Peso = req.Peso
		if porte := strings.TrimSpace(string(req.Porte)); porte != "" {
			pet.Porte = models.PetPorte(porte)
		}
		if regiao := strings.TrimSpace(string(req.Regiao)); regiao != "" {
			pet.Regiao = models.PetRegiao(regiao)
		}
		if req.Sexo != "" {
			pet.Sexo = models.PetSexo(strings.TrimSpace(string(req.Sexo)))
		}
		if req.Status != "" {
			pet.Status = models.PetStatus(req.Status)
		}
		if req.OngID != uuid.Nil {
			pet.OngID = req.OngID
		}
		pet.FormularioID = formularioID

		if err := tx.Save(&pet).Error; err != nil {
			return err
		}
		return tx.Preload("Imagens", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("position ASC")
		}).First(&pet, "id = ?", petID).Error
	})
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pet atualizado com sucesso",
		"pet":     pet,
	})
	return nil
}

// @Summary Remove um Pet
// @Description Remove um pet pelo ID
// @Tags Pet
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID do Pet"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/pets/{id} [delete]
func DeletePet(c *gin.Context) error {
	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do pet inválido")
	}

	return database.WithUserDB(c.GetString("token"), func(tx *gorm.DB) error {
		result := tx.Where("id = ?", petID).Delete(&models.Pet{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("Pet não encontrado")
		}
		c.JSON(http.StatusOK, gin.H{"message": "Pet removido com sucesso"})
		return nil
	})
}

// @Summary Remove uma imagem de um Pet
// @Description Remove uma imagem específica de um pet
// @Tags Pet
// @Security ApiKeyAuth
// @Produce json
// @Param id      path string true "ID do Pet"
// @Param imageId path string true "ID da Imagem"
// @Success 200 {object} object{message=string,pet=models.Pet}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/pets/{id}/imagens/{imageId} [delete]
func DeletePetImage(c *gin.Context) error {
	petID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do pet inválido")
	}

	imageID, err := uuid.Parse(c.Param("imageId"))
	if err != nil {
		return errors.New("ID da imagem inválido")
	}

	token := c.GetString("token")

	var (
		image      models.PetImage
		objectPath string
	)
	err = database.WithUserDB(token, func(tx *gorm.DB) error {
		if err := tx.First(&image, "id = ? AND pet_id = ?", imageID, petID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("Imagem não encontrada")
			}
			return err
		}
		if path, pathErr := utils.ExtractObjectPath(image.URL, utils.SupabaseBucketPets); pathErr == nil {
			objectPath = path
		} else {
			fmt.Printf("⚠️ Aviso: não foi possível extrair path do storage: %s\n", pathErr.Error())
		}
		return tx.Delete(&image).Error
	})
	if err != nil {
		return err
	}

	// Remove do storage após commit do banco
	if objectPath != "" {
		if err := utils.DeleteFile(utils.SupabaseBucketPets, objectPath, c.GetString("user_token")); err != nil {
			fmt.Printf("⚠️ Aviso: não foi possível deletar arquivo do storage: %s\n", err.Error())
		}
	}

	var pet models.Pet
	if err := database.WithUserDB(token, func(tx *gorm.DB) error {
		return tx.Preload("Imagens", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("position ASC")
		}).First(&pet, "id = ?", petID).Error
	}); err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Imagem removida com sucesso",
		"pet":     pet,
	})
	return nil
}
