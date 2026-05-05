package v1

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ongpet/database"
	"ongpet/models"
	"ongpet/utils"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────
// INPUT TYPES
// ─────────────────────────────────────────────────────────────

type CreateBannerInput struct {
	Title        string    `json:"title"`
	InstagramURL string    `json:"instagram_url"`
	StartAt      string    `json:"start_at"`
	EndAt        string    `json:"end_at"`
	Active       *bool     `json:"active"`
	Position     *int      `json:"position"`
	OngID        uuid.UUID `json:"ong_id"`
}

type UpdateBannerInput struct {
	Title        *string    `json:"title"`
	InstagramURL *string    `json:"instagram_url"`
	StartAt      *string    `json:"start_at"`
	EndAt        *string    `json:"end_at"`
	Active       *bool      `json:"active"`
	Position     *int       `json:"position"`
	OngID        *uuid.UUID `json:"ong_id"`
}

type BannerListResponse struct {
	Dados          []models.Banner `json:"dados"`
	TotalRegistros int64           `json:"total_registros"`
	TotalPaginas   int             `json:"total_paginas"`
	ProximaPagina  bool            `json:"proxima_pagina"`
}

// ─────────────────────────────────────────────────────────────
// READ (com paginação e filtros)
// ─────────────────────────────────────────────────────────────

// @Summary Lista todos os Banners
// @Description Retorna banners com filtros e paginação
// @Tags Banner
// @Produce json
// @Param active query bool false "Filtrar por ativo"
// @Param vigente query bool false "Filtrar apenas banners no período vigente (start_at <= now <= end_at)"
// @Param page query int false "Página (default: 1)"
// @Param limit query int false "Itens por página (default: 10)"
// @Success 200 {object} v1.BannerListResponse
// @Router /api/v1/banners [get]
func ReadBanners(c *gin.Context) error {
	db := database.GetDB()
	query := db.Model(&models.Banner{})

	// filtros
	if activeStr := strings.TrimSpace(c.Query("active")); activeStr != "" {
		active := !strings.EqualFold(activeStr, "false") && activeStr != "0"
		query = query.Where("active = ?", active)
	}

	if ongID := strings.TrimSpace(c.Query("ong_id")); ongID != "" {
		query = query.Where("ong_id = ?", ongID)
	}

	if vigente := strings.TrimSpace(c.Query("vigente")); vigente == "true" || vigente == "1" {
		now := time.Now()
		query = query.Where("start_at <= ?", now).Where("end_at >= ?", now)
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

	var banners []models.Banner
	if err := query.
		Order("position ASC").
		Offset(offset).
		Limit(limit + 1).
		Find(&banners).Error; err != nil {
		return err
	}

	hasNext := false
	if len(banners) > limit {
		hasNext = true
		banners = banners[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"dados":           banners,
		"total_registros": total,
		"total_paginas":   totalPages,
		"proxima_pagina":  hasNext,
	})

	return nil
}

// @Summary Busca um Banner pelo ID
// @Description Retorna um Banner específico pelo ID
// @Tags Banner
// @Produce json
// @Param id path string true "ID do Banner"
// @Success 200 {object} models.Banner
// @Failure 404 {object} map[string]string
// @Router /api/v1/banners/{id} [get]
func ReadBanner(c *gin.Context) error {
	db := database.GetDB()

	bannerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do banner inválido")
	}

	var banner models.Banner
	if err := db.First(&banner, "id = ?", bannerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Banner não encontrado")
		}
		return err
	}

	c.JSON(http.StatusOK, banner)
	return nil
}

// ─────────────────────────────────────────────────────────────
// CREATE (apenas metadados via JSON)
// ─────────────────────────────────────────────────────────────

// @Summary Cria um novo Banner
// @Description Cria um Banner no sistema (sem imagem — use o endpoint de upload)
// @Tags Banner
// @Accept json
// @Produce json
// @Param banner body v1.CreateBannerInput true "Novo Banner"
// @Success 201 {object} models.Banner
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/v1/banners [post]
func CreateBanner(c *gin.Context) error {
	var req CreateBannerInput

	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	req.Title = strings.TrimSpace(req.Title)
	req.InstagramURL = strings.TrimSpace(req.InstagramURL)

	if req.Title == "" {
		return errors.New("title é obrigatório")
	}
	if req.InstagramURL == "" {
		return errors.New("instagram_url é obrigatório")
	}
	if req.StartAt == "" {
		return errors.New("start_at é obrigatório")
	}
	if req.EndAt == "" {
		return errors.New("end_at é obrigatório")
	}

	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		return errors.New("start_at deve estar no formato RFC3339 (ex: 2026-03-11T15:04:05Z)")
	}

	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		return errors.New("end_at deve estar no formato RFC3339 (ex: 2026-03-11T15:04:05Z)")
	}

	if req.OngID == uuid.Nil {
		return errors.New("ong_id é obrigatório")
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	position := 0
	if req.Position != nil {
		position = *req.Position
	}

	db := database.GetDB()

	if err := db.First(&models.Ong{}, "id = ?", req.OngID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("ONG não encontrada")
		}
		return err
	}

	banner := models.Banner{
		Title:        req.Title,
		InstagramURL: req.InstagramURL,
		StartAt:      startAt,
		EndAt:        endAt,
		Active:       active,
		Position:     position,
		OngID:        &req.OngID,
	}

	if err := db.Create(&banner).Error; err != nil {
		return err
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Banner criado com sucesso",
		"banner":  banner,
	})

	return nil
}

// ─────────────────────────────────────────────────────────────
// UPLOAD DE IMAGEM (endpoint dedicado, igual ao UploadPetImages)
// ─────────────────────────────────────────────────────────────

// @Summary Faz upload da imagem de um Banner
// @Description Envia uma imagem para o Banner pelo ID (substitui a anterior)
// @Tags Banner
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "ID do Banner"
// @Param image formData file true "Imagem do Banner"
// @Success 200 {object} object{message=string,banner=models.Banner}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/banners/{id}/imagem [post]
func UploadBannerImage(c *gin.Context) error {
	db := database.GetDB()

	bannerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do banner inválido")
	}

	var banner models.Banner
	if err := db.First(&banner, "id = ?", bannerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Banner não encontrado")
		}
		return err
	}

	file, err := c.FormFile("image")
	if err != nil {
		return errors.New("imagem é obrigatória")
	}

	if !utils.IsValidImage(file) {
		return errors.New("tipo de imagem inválido")
	}

	optimized, err := utils.ResizeAndCompressImage(file, 1280, 70)
	if err != nil {
		return err
	}

	// deleta imagem anterior do storage (se existir)
	if banner.ImageURL != "" {
		if objectPath, pathErr := utils.ExtractObjectPath(banner.ImageURL, utils.SupabaseBucketOngs); pathErr == nil {
			if err := utils.DeleteFile(utils.SupabaseBucketOngs, objectPath); err != nil {
				fmt.Printf("⚠️ Aviso: não foi possível deletar imagem anterior do storage: %s\n", err.Error())
			}
		} else {
			fmt.Printf("⚠️ Aviso: não foi possível extrair path do storage: %s\n", pathErr.Error())
		}
	}

	imageURL, err := utils.UploadOptimizedFile(
		optimized.Buffer,
		optimized.ContentType,
		optimized.Extension,
		utils.SupabaseBucketOngs,
		banner.ID.String(),
		banner.Title,
		banner.Position,
	)
	if err != nil {
		return err
	}

	banner.ImageURL = imageURL

	if err := db.Save(&banner).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Imagem atualizada com sucesso",
		"banner":  banner,
	})

	return nil
}

// ─────────────────────────────────────────────────────────────
// UPDATE (apenas metadados via JSON)
// ─────────────────────────────────────────────────────────────

// @Summary Atualiza um Banner existente
// @Description Atualiza os metadados de um Banner pelo ID (não atualiza imagem)
// @Tags Banner
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "ID do Banner"
// @Param banner body v1.UpdateBannerInput true "Dados para atualização"
// @Success 200 {object} object{message=string,banner=models.Banner}
// @Failure 404 {object} map[string]string
// @Router /api/v1/banners/{id} [put]
func UpdateBanner(c *gin.Context) error {
	var req UpdateBannerInput

	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	db := database.GetDB()

	bannerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do banner inválido")
	}

	var banner models.Banner
	if err := db.First(&banner, "id = ?", bannerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Banner não encontrado")
		}
		return err
	}

	if req.Title != nil {
		if t := strings.TrimSpace(*req.Title); t != "" {
			banner.Title = t
		}
	}
	if req.InstagramURL != nil {
		if u := strings.TrimSpace(*req.InstagramURL); u != "" {
			banner.InstagramURL = u
		}
	}
	if req.StartAt != nil {
		t, parseErr := time.Parse(time.RFC3339, *req.StartAt)
		if parseErr != nil {
			return errors.New("start_at deve estar no formato RFC3339")
		}
		banner.StartAt = t
	}
	if req.EndAt != nil {
		t, parseErr := time.Parse(time.RFC3339, *req.EndAt)
		if parseErr != nil {
			return errors.New("end_at deve estar no formato RFC3339")
		}
		banner.EndAt = t
	}
	if req.Active != nil {
		banner.Active = *req.Active
	}
	if req.Position != nil {
		banner.Position = *req.Position
	}
	if req.OngID != nil && *req.OngID != uuid.Nil {
		banner.OngID = req.OngID
	}

	if err := db.Save(&banner).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Banner atualizado com sucesso",
		"banner":  banner,
	})

	return nil
}

// ─────────────────────────────────────────────────────────────
// DELETE IMAGEM DO BANNER
// ─────────────────────────────────────────────────────────────

// @Summary Remove a imagem de um Banner
// @Description Remove a imagem do storage e limpa o image_url do Banner
// @Tags Banner
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID do Banner"
// @Success 200 {object} object{message=string,banner=models.Banner}
// @Failure 404 {object} map[string]string
// @Router /api/v1/banners/{id}/imagem [delete]
func DeleteBannerImage(c *gin.Context) error {
	db := database.GetDB()

	bannerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do banner inválido")
	}

	var banner models.Banner
	if err := db.First(&banner, "id = ?", bannerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Banner não encontrado")
		}
		return err
	}

	if banner.ImageURL == "" {
		return errors.New("este banner não possui imagem")
	}

	if objectPath, pathErr := utils.ExtractObjectPath(banner.ImageURL, utils.SupabaseBucketOngs); pathErr == nil {
		if err := utils.DeleteFile(utils.SupabaseBucketOngs, objectPath); err != nil {
			fmt.Printf("⚠️ Aviso: não foi possível deletar imagem do storage: %s\n", err.Error())
		}
	} else {
		fmt.Printf("⚠️ Aviso: não foi possível extrair path do storage: %s\n", pathErr.Error())
	}

	banner.ImageURL = ""

	if err := db.Save(&banner).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Imagem removida com sucesso",
		"banner":  banner,
	})

	return nil
}

// ─────────────────────────────────────────────────────────────
// DELETE BANNER
// ─────────────────────────────────────────────────────────────

// @Summary Remove um Banner
// @Description Remove um Banner pelo ID (e sua imagem do storage)
// @Tags Banner
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID do Banner"
// @Success 200 {object} object{message=string}
// @Failure 404 {object} map[string]string
// @Router /api/v1/banners/{id} [delete]
func DeleteBanner(c *gin.Context) error {
	db := database.GetDB()

	bannerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do banner inválido")
	}

	var banner models.Banner
	if err := db.First(&banner, "id = ?", bannerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Banner não encontrado")
		}
		return err
	}

	if banner.ImageURL != "" {
		if objectPath, pathErr := utils.ExtractObjectPath(banner.ImageURL, utils.SupabaseBucketOngs); pathErr == nil {
			if err := utils.DeleteFile(utils.SupabaseBucketOngs, objectPath); err != nil {
				fmt.Printf("⚠️ Aviso: não foi possível deletar imagem do storage: %s\n", err.Error())
			}
		}
	}

	result := db.Where("id = ?", bannerID).Delete(&models.Banner{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("Banner não encontrado")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Banner removido com sucesso",
	})

	return nil
}
