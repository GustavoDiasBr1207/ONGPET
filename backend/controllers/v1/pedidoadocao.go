package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
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

type CreateRespostaInput struct {
	CampoFormularioID uuid.UUID `json:"campo_formulario_id"`
	Valor             string    `json:"valor"`
}

type CreatePedidoAdocaoInput struct {
	OngID     uuid.UUID             `json:"ong_id"`
	PetID     uuid.UUID             `json:"pet_id"`
	Respostas []CreateRespostaInput `json:"respostas"`
}

type PedidoAdocaoListResponse struct {
	Dados          []models.PedidoAdocaoDTO `json:"dados"`
	TotalRegistros int64                    `json:"total_registros"`
	TotalPaginas   int                      `json:"total_paginas"`
	ProximaPagina  bool                     `json:"proxima_pagina"`
}

// ─────────────────────────────────────────────────────────────
// PEDIDO ADOCAO — CRUD
// ─────────────────────────────────────────────────────────────

// @Summary Lista todos os Pedidos de Adoção
// @Tags PedidoAdocao
// @Produce json
// @Param ong_id query string false "Filtrar por ONG"
// @Param pet_id query string false "Filtrar por Pet"
// @Param page query int false "Página"
// @Param limit query int false "Itens por página"
// @Success 200 {object} v1.PedidoAdocaoListResponse
// @Security ApiKeyAuth
// @Router /api/v1/pedidos-adocao [get]
func ReadPedidosAdocao(c *gin.Context) error {
	db := database.GetUserDB(c.GetString("token"))
	query := db.Model(&models.PedidoAdocao{})

	if ongID := strings.TrimSpace(c.Query("ong_id")); ongID != "" {
		query = query.Where("ong_id = ?", ongID)
	}
	if petID := strings.TrimSpace(c.Query("pet_id")); petID != "" {
		query = query.Where("pet_id = ?", petID)
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
	if limit > 100 {
		limit = 100
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return err
	}

	offset     := (page - 1) * limit
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	var pedidos []models.PedidoAdocao
	if err := query.
		Preload("Respostas.CampoFormulario").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit + 1).
		Find(&pedidos).Error; err != nil {
		return err
	}

	hasNext := false
	if len(pedidos) > limit {
		hasNext = true
		pedidos = pedidos[:limit]
	}

	dtos := make([]models.PedidoAdocaoDTO, len(pedidos))
	for i, p := range pedidos {
		dtos[i] = mapPedidoToDTO(p)
	}

	c.JSON(http.StatusOK, PedidoAdocaoListResponse{
		Dados:          dtos,
		TotalRegistros: total,
		TotalPaginas:   totalPages,
		ProximaPagina:  hasNext,
	})
	return nil
}

// @Summary Busca um Pedido de Adoção pelo ID
// @Tags PedidoAdocao
// @Produce json
// @Param id path string true "ID do Pedido"
// @Success 200 {object} models.PedidoAdocaoDTO
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/v1/pedidos-adocao/{id} [get]
func ReadPedidoAdocao(c *gin.Context) error {
	db := database.GetUserDB(c.GetString("token"))

	pedidoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do pedido inválido")
	}

	var pedido models.PedidoAdocao
	if err := db.
		Preload("Respostas.CampoFormulario").
		First(&pedido, "id = ?", pedidoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Pedido de adoção não encontrado")
		}
		return err
	}

	c.JSON(http.StatusOK, mapPedidoToDTO(pedido))
	return nil
}

// @Summary Cria um novo Pedido de Adoção
// @Description Cria o pedido com as respostas ao formulário vinculado ao Pet.
// @Description Para campos do tipo 'imagem', envie o campo_formulario_id com valor vazio ("").
// @Description Após criar o pedido, use POST /pedidos-adocao/:pedidoId/respostas/:respostaId/imagem
// @Description para enviar a imagem de cada campo do tipo imagem.
// @Tags PedidoAdocao
// @Accept json
// @Produce json
// @Param pedido body v1.CreatePedidoAdocaoInput true "Novo Pedido"
// @Success 201 {object} models.PedidoAdocaoDTO
// @Failure 400 {object} map[string]string
// @Router /api/v1/pedidos-adocao [post]
func CreatePedidoAdocao(c *gin.Context) error {
	var req CreatePedidoAdocaoInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	if req.OngID == uuid.Nil {
		return errors.New("ong_id é obrigatório")
	}
	if req.PetID == uuid.Nil {
		return errors.New("pet_id é obrigatório")
	}

	db := database.GetDB() // público — sem token, sem RLS

	if err := db.First(&models.Ong{}, "id = ?", req.OngID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("ONG não encontrada")
		}
		return err
	}

	var pet models.Pet
	if err := db.First(&pet, "id = ?", req.PetID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Pet não encontrado")
		}
		return err
	}

	if pet.FormularioID != nil && *pet.FormularioID != uuid.Nil {
		var campos []models.CampoFormulario
		if err := db.
			Where("formulario_modelo_id = ?", *pet.FormularioID).
			Order("ordem ASC").
			Find(&campos).Error; err != nil {
			return err
		}
		if err := validarRespostas(req.Respostas, campos); err != nil {
			return err
		}
	}

	pedido := models.PedidoAdocao{
		OngID:  req.OngID,
		PetID:  req.PetID,
		Status: models.PedidoAdocaoPendente,
	}

	if err := db.Create(&pedido).Error; err != nil {
		return err
	}

	for _, r := range req.Respostas {
		if r.CampoFormularioID == uuid.Nil {
			continue
		}
		resposta := models.RespostaFormulario{
			PedidoAdocaoID:    pedido.ID,
			CampoFormularioID: r.CampoFormularioID,
			Valor:             r.Valor,
		}
		if err := db.Create(&resposta).Error; err != nil {
			return err
		}
	}

	if err := db.
		Preload("Respostas.CampoFormulario").
		First(&pedido, "id = ?", pedido.ID).Error; err != nil {
		return err
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pedido de adoção criado com sucesso",
		"pedido":  mapPedidoToDTO(pedido),
	})
	return nil
}

// @Summary Remove um Pedido de Adoção
// @Tags PedidoAdocao
// @Produce json
// @Param id path string true "ID do Pedido"
// @Success 200 {object} object{message=string}
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/v1/pedidos-adocao/{id} [delete]
func DeletePedidoAdocao(c *gin.Context) error {
	db := database.GetUserDB(c.GetString("token"))

	pedidoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do pedido inválido")
	}

	if err := db.Where("pedido_adocao_id = ?", pedidoID).
		Delete(&models.RespostaFormulario{}).Error; err != nil {
		return err
	}

	result := db.Where("id = ?", pedidoID).Delete(&models.PedidoAdocao{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("Pedido de adoção não encontrado")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pedido de adoção removido com sucesso"})
	return nil
}

// ─────────────────────────────────────────────────────────────
// UPLOAD DE IMAGEM EM RESPOSTA DE CAMPO TIPO "imagem"
// ─────────────────────────────────────────────────────────────

// @Summary Faz upload de imagem em uma resposta de campo do tipo 'imagem'
// @Description Endpoint público (sem auth). Após criar o pedido, chame este endpoint
// @Description para cada campo do tipo 'imagem'. A URL gerada é salva em RespostaFormulario.Valor.
// @Description Se a resposta já possuía uma imagem anterior, ela é removida do storage antes do novo upload.
// @Tags PedidoAdocao
// @Accept multipart/form-data
// @Produce json
// @Param pedidoId   path     string true "ID do Pedido de Adoção"
// @Param respostaId path     string true "ID da Resposta (campo do tipo imagem)"
// @Param imagem     formData file   true "Arquivo de imagem"
// @Success 200 {object} object{message=string,resposta=models.RespostaFormulario}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/pedidos-adocao/{pedidoId}/respostas/{respostaId}/imagem [post]
func UploadRespostaImagem(c *gin.Context) error {
	db := database.GetDB() // público — sem token, sem RLS

	pedidoID, err := uuid.Parse(c.Param("pedidoId"))
	if err != nil {
		return errors.New("ID do pedido inválido")
	}

	respostaID, err := uuid.Parse(c.Param("respostaId"))
	if err != nil {
		return errors.New("ID da resposta inválido")
	}

	var resposta models.RespostaFormulario
	if err := db.
		Preload("CampoFormulario").
		First(&resposta, "id = ? AND pedido_adocao_id = ?", respostaID, pedidoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Resposta não encontrada")
		}
		return err
	}

	var cfg models.CampoConfiguracao
	if err := json.Unmarshal(resposta.CampoFormulario.Configuracao, &cfg); err != nil {
		return errors.New("erro ao ler configuração do campo")
	}
	if cfg.Tipo != models.TipoCampoImagem {
		return fmt.Errorf("campo '%s' não é do tipo imagem", cfg.Label)
	}

	file, err := c.FormFile("imagem")
	if err != nil {
		return errors.New("nenhuma imagem enviada")
	}

	if !utils.IsValidImage(file) {
		return errors.New("tipo de imagem inválido")
	}

	if resposta.Valor != "" {
		if objectPath, pathErr := utils.ExtractObjectPath(resposta.Valor, utils.SupabaseBucketFormularios); pathErr == nil {
			if delErr := utils.DeleteFile(utils.SupabaseBucketFormularios, objectPath); delErr != nil {
				fmt.Printf("⚠️ Aviso: não foi possível deletar imagem anterior do storage: %s\n", delErr.Error())
			}
		} else {
			fmt.Printf("⚠️ Aviso: não foi possível extrair path do storage: %s\n", pathErr.Error())
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
		pedidoID.String(),
		cfg.Label,
		1,
	)
	if err != nil {
		return err
	}

	resposta.Valor = url
	if err := db.Save(&resposta).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Imagem enviada com sucesso",
		"resposta": resposta,
	})
	return nil
}

// ─────────────────────────────────────────────────────────────
// STATUS
// ─────────────────────────────────────────────────────────────

type UpdateStatusPedidoAdocaoInput struct {
	Status models.PedidoAdocaoStatus `json:"status"`
}

// @Summary Atualiza o status de um Pedido de Adoção
// @Description Status válidos: pendente, aprovado, rejeitado, cancelado
// @Tags PedidoAdocao
// @Accept json
// @Produce json
// @Param id path string true "ID do Pedido"
// @Param status body v1.UpdateStatusPedidoAdocaoInput true "Novo status"
// @Success 200 {object} models.PedidoAdocaoDTO
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/v1/pedidos-adocao/{id}/status [put]
func UpdateStatusPedidoAdocao(c *gin.Context) error {
	var req UpdateStatusPedidoAdocaoInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	switch req.Status {
	case models.PedidoAdocaoPendente,
		models.PedidoAdocaoAprovado,
		models.PedidoAdocaoRejeitado,
		models.PedidoAdocaoCancelado:
		// ok
	default:
		return errors.New("status inválido. Use: pendente, aprovado, rejeitado ou cancelado")
	}

	db := database.GetUserDB(c.GetString("token"))

	pedidoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.New("ID do pedido inválido")
	}

	var pedido models.PedidoAdocao
	if err := db.First(&pedido, "id = ?", pedidoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Pedido de adoção não encontrado")
		}
		return err
	}

	pedido.Status = req.Status

	if err := db.Save(&pedido).Error; err != nil {
		return err
	}

	if err := db.
		Preload("Respostas.CampoFormulario").
		First(&pedido, "id = ?", pedidoID).Error; err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status atualizado com sucesso",
		"pedido":  mapPedidoToDTO(pedido),
	})
	return nil
}

// ─────────────────────────────────────────────────────────────
// VALIDAÇÃO DE RESPOSTAS
// ─────────────────────────────────────────────────────────────

func validarRespostas(respostas []CreateRespostaInput, campos []models.CampoFormulario) error {
	respostaMap := make(map[uuid.UUID]string, len(respostas))
	for _, r := range respostas {
		respostaMap[r.CampoFormularioID] = r.Valor
	}

	for _, campo := range campos {
		var cfg models.CampoConfiguracao
		if err := json.Unmarshal(campo.Configuracao, &cfg); err != nil {
			continue
		}

		valor, respondido := respostaMap[campo.ID]

		if cfg.Tipo == models.TipoCampoImagem {
			if cfg.Obrigatorio && !respondido {
				return fmt.Errorf("campo '%s' é obrigatório — inclua-o nas respostas com valor vazio e envie a imagem via POST .../respostas/:id/imagem", cfg.Label)
			}
			continue
		}

		if cfg.Obrigatorio && (!respondido || strings.TrimSpace(valor) == "") {
			return fmt.Errorf("campo '%s' é obrigatório", cfg.Label)
		}
		if !respondido || strings.TrimSpace(valor) == "" {
			continue
		}

		switch cfg.Tipo {
		case models.TipoCampoTexto, models.TipoCampoTextarea, models.TipoCampoEmail, models.TipoCampoTelefone:
			if cfg.Min != nil && len(valor) < *cfg.Min {
				return fmt.Errorf("campo '%s' deve ter no mínimo %d caracteres", cfg.Label, *cfg.Min)
			}
			if cfg.Max != nil && len(valor) > *cfg.Max {
				return fmt.Errorf("campo '%s' deve ter no máximo %d caracteres", cfg.Label, *cfg.Max)
			}
			if cfg.Regex != "" {
				matched, err := regexp.MatchString(cfg.Regex, valor)
				if err != nil || !matched {
					return fmt.Errorf("campo '%s' tem formato inválido", cfg.Label)
				}
			}

		case models.TipoCampoNumero:
			num, err := strconv.ParseFloat(valor, 64)
			if err != nil {
				return fmt.Errorf("campo '%s' deve ser um número", cfg.Label)
			}
			if cfg.MinValor != nil && num < *cfg.MinValor {
				return fmt.Errorf("campo '%s' deve ser no mínimo %.2f", cfg.Label, *cfg.MinValor)
			}
			if cfg.MaxValor != nil && num > *cfg.MaxValor {
				return fmt.Errorf("campo '%s' deve ser no máximo %.2f", cfg.Label, *cfg.MaxValor)
			}

		case models.TipoCampoSelect, models.TipoCampoRadio:
			if len(cfg.Opcoes) > 0 && !contains(cfg.Opcoes, valor) {
				return fmt.Errorf("campo '%s': opção '%s' inválida", cfg.Label, valor)
			}

		case models.TipoCampoCheckbox:
			for _, s := range strings.Split(valor, ",") {
				s = strings.TrimSpace(s)
				if len(cfg.Opcoes) > 0 && !contains(cfg.Opcoes, s) {
					return fmt.Errorf("campo '%s': opção '%s' inválida", cfg.Label, s)
				}
			}
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ─────────────────────────────────────────────────────────────
// HELPER: PedidoAdocao → DTO
// ─────────────────────────────────────────────────────────────

func mapPedidoToDTO(p models.PedidoAdocao) models.PedidoAdocaoDTO {
	respostasDTO := make([]models.RespostaFormularioDTO, len(p.Respostas))
	for i, r := range p.Respostas {
		var cfg models.CampoConfiguracao
		_ = json.Unmarshal(r.CampoFormulario.Configuracao, &cfg)
		respostasDTO[i] = models.RespostaFormularioDTO{
			ID: r.ID,
			Campo: models.CampoFormularioDTO{
				ID:           r.CampoFormulario.ID,
				Nome:         r.CampoFormulario.Nome,
				Ordem:        r.CampoFormulario.Ordem,
				Configuracao: cfg,
			},
			Valor: r.Valor,
		}
	}
	return models.PedidoAdocaoDTO{
		ID:        p.ID,
		OngID:     p.OngID,
		PetID:     p.PetID,
		Status:    p.Status,
		CreatedAt: p.CreatedAt,
		Respostas: respostasDTO,
	}
}