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

// CreateOngInput representa o payload mínimo para criar uma ONG
type CreateOngInput struct {
	Email           string           `json:"email"            example:"a@b.com"`
	Endereco        string           `json:"endereco"         example:"Rua A"`
	Nome            string           `json:"nome"             example:"Minha ONG"`
	NomeResponsavel string           `json:"nome_responsavel" example:"Fulano"`
	Telefone        string           `json:"telefone"         example:"123"`
	WhatsappLink    string           `json:"whatsapp"         example:"https://wa.me/5527999999999"`
	Descricao       string           `json:"descricao"        example:"ONG dedicada ao resgate de animais"`
	Site            string           `json:"site"             example:"https://minhaong.org"`
	Instagram       string           `json:"instagram"        example:"@minhaong"`
	Regiao          models.OngRegiao `json:"regiao"           example:"Serra"`
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
// @Param nome  query string false "Filtrar por nome"
// @Param email query string false "Filtrar por e-mail"
// @Param page  query int    false "Página (default: 1)"
// @Param limit query int    false "Itens por página (default: 10)"
// @Success 200 {object} v1.OngListResponse
// @Router /api/v1/ongs [get]
func ReadOngs(c *gin.Context) error {
	db := database.GetDB() // público — sem token
	query := db.Model(&models.Ong{})

	if nome := strings.TrimSpace(c.Query("nome")); nome != "" {
		query = query.Where("nome ILIKE ?", "%"+nome+"%")
	}
	if email := strings.TrimSpace(c.Query("email")); email != "" {
		query = query.Where("email ILIKE ?", "%"+email+"%")
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

// @Summary Busca uma ONG pelo ID
// @Description Retorna uma ONG específica pelo ID
// @Tags ONG
// @Produce json
// @Param id path string true "ID da ONG"
// @Success 200 {object} models.Ong
// @Failure 404 {object} map[string]string
// @Router /api/v1/ongs/{id} [get]
func ReadOng(c *gin.Context) error {
	db := database.GetDB() // público — sem token

	ongID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID da ONG inválido")
	}

	var ong models.Ong
	if err := db.First(&ong, "id = ?", ongID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("ONG não encontrada")
		}
		return err
	}

	c.JSON(http.StatusOK, ong)
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

	req.Nome  = strings.TrimSpace(req.Nome)
	req.Email = strings.TrimSpace(req.Email)

	db := database.GetUserDB(c.GetString("token")) // autenticado — RLS ativo

	if err := db.Where("email = ?", req.Email).First(&models.Ong{}).Error; err == nil {
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
		WhatsappLink:    req.WhatsappLink,
		Descricao:       req.Descricao,
		Site:            req.Site,
		NomeResponsavel: req.NomeResponsavel,
		Instagram:       req.Instagram,
		Regiao:          req.Regiao,
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

// @Summary Faz upload do logo de uma ONG
// @Description Envia uma imagem de logo para a ONG pelo ID (usando Supabase)
// @Tags ONG
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param id   path     string true "ID da ONG"
// @Param logo formData file   true "Logo da ONG"
// @Success 200 {object} object{message=string,ong=models.Ong}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/ongs/{id}/logo [post]
func UploadOngLogo(c *gin.Context) error {
	db := database.GetUserDB(c.GetString("token")) // autenticado — RLS ativo

	ongID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID da ONG inválido")
	}

	var ong models.Ong
	if err := db.First(&ong, "id = ?", ongID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("ONG não encontrada")
		}
		return err
	}

	file, err := c.FormFile("logo")
	if err != nil || file == nil {
		return errors.New("nenhuma imagem enviada")
	}

	if !utils.IsValidImage(file) {
		return errors.New("tipo de imagem inválido")
	}

	// Remove logo anterior do storage, se existir
	if ong.Logo != "" {
		if objectPath, pathErr := utils.ExtractObjectPath(ong.Logo); pathErr == nil {
			if delErr := utils.DeleteFile(objectPath); delErr != nil {
				fmt.Printf("⚠️ Aviso: não foi possível deletar logo anterior do storage: %s\n", delErr.Error())
			}
		}
	}

	optimized, err := utils.ResizeAndCompressImage(file, 512, 80)
	if err != nil {
		return err
	}

	url, err := utils.UploadOptimizedFile(
		optimized.Buffer,
		optimized.ContentType,
		optimized.Extension,
		ongID.String(),
		ong.Nome,
		0, // posição fixa — logo único
	)
	if err != nil {
		return err
	}

	ong.Logo = url
	if err := db.Save(&ong).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logo enviado com sucesso",
		"ong":     ong,
	})
	return nil
}

// @Summary Remove o logo de uma ONG
// @Description Remove o logo de uma ONG pelo ID
// @Tags ONG
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID da ONG"
// @Success 200 {object} object{message=string,ong=models.Ong}
// @Failure 404 {object} map[string]string
// @Router /api/v1/ongs/{id}/logo [delete]
func DeleteOngLogo(c *gin.Context) error {
	db := database.GetUserDB(c.GetString("token")) // autenticado — RLS ativo

	ongID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID da ONG inválido")
	}

	var ong models.Ong
	if err := db.First(&ong, "id = ?", ongID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("ONG não encontrada")
		}
		return err
	}

	if ong.Logo == "" {
		return errors.New("ONG não possui logo")
	}

	objectPath, pathErr := utils.ExtractObjectPath(ong.Logo)
	if pathErr == nil {
		if err := utils.DeleteFile(objectPath); err != nil {
			fmt.Printf("⚠️ Aviso: não foi possível deletar logo do storage: %s\n", err.Error())
		}
	} else {
		fmt.Printf("⚠️ Aviso: não foi possível extrair path do storage: %s\n", pathErr.Error())
	}

	ong.Logo = ""
	if err := db.Save(&ong).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logo removido com sucesso",
		"ong":     ong,
	})
	return nil
}

// @Summary Atualiza uma ONG existente
// @Description Atualiza os dados de uma ONG pelo ID (não atualiza logo)
// @Tags ONG
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id  path string            true "ID da ONG"
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

	db := database.GetUserDB(c.GetString("token")) // autenticado — RLS ativo

	ongID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID da ONG inválido")
	}

	var ong models.Ong
	if err := db.First(&ong, "id = ?", ongID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("ONG não encontrada")
		}
		return err
	}

	if nome := strings.TrimSpace(req.Nome); nome != "" {
		ong.Nome = nome
	}
	if email := strings.TrimSpace(req.Email); email != "" {
		ong.Email = email
	}
	if req.Endereco != "" {
		ong.Endereco = strings.TrimSpace(req.Endereco)
	}
	if req.Telefone != "" {
		ong.Telefone = strings.TrimSpace(req.Telefone)
	}
	if req.WhatsappLink != "" {
		ong.WhatsappLink = strings.TrimSpace(req.WhatsappLink)
	}
	if req.Descricao != "" {
		ong.Descricao = strings.TrimSpace(req.Descricao)
	}
	if req.Site != "" {
		ong.Site = strings.TrimSpace(req.Site)
	}
	if req.NomeResponsavel != "" {
		ong.NomeResponsavel = strings.TrimSpace(req.NomeResponsavel)
	}
	if req.Instagram != "" {
		ong.Instagram = strings.TrimSpace(req.Instagram)
	}
	if req.Regiao != "" {
		ong.Regiao = req.Regiao
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
	db := database.GetUserDB(c.GetString("token")) // autenticado — RLS ativo

	ongID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID da ONG inválido")
	}

	result := db.Where("id = ?", ongID).Delete(&models.Ong{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("ONG não encontrada")
	}

	c.JSON(http.StatusOK, gin.H{"message": "ONG removida com sucesso"})
	return nil
}