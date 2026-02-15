package v0

import (
    "ongpet/database"
    "ongpet/models"
    "github.com/gin-gonic/gin"
    "strings"
)

// CreateOngInput representa o payload mínimo para criar uma ONG
type CreateOngInput struct {
    Email           string `json:"email" example:"a@b.com"`
    Endereco        string `json:"endereco" example:"Rua A"`
    Nome            string `json:"nome" example:"Minha ONG"`
    NomeResponsavel string `json:"nome_responsavel" example:"Fulano"`
    Telefone        string `json:"telefone" example:"123"`
}

// @Summary Lista todas as ONGs
// @Description Retorna todas as ONGs cadastradas
// @Tags ONG
// @Produce json
// @Success 200 {array} models.Ong
// @Router /api/v1/ongs [get]
func GetOngs(c *gin.Context) {
    db := database.GetDB()
    var ongs []models.Ong
    if err := db.Find(&ongs).Error; err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, ongs)
}

// @Summary Cria uma nova ONG
// @Description Cria uma ONG no sistema
// @Tags ONG
// @Accept json
// @Produce json
// @Param ong body v0.CreateOngInput true "Nova ONG"
// @Success 201 {object} models.Ong
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/v1/ongs [post]
func CreateOng(c *gin.Context) {
    db := database.GetDB()

    var input CreateOngInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // basic validation
    input.Nome = strings.TrimSpace(input.Nome)
    input.Email = strings.TrimSpace(input.Email)
    if input.Nome == "" || input.Email == "" {
        c.JSON(400, gin.H{"error": "nome and email are required"})
        return
    }

    ong := models.Ong{
        Nome:            input.Nome,
        Email:           input.Email,
        Endereco:        input.Endereco,
        Telefone:        input.Telefone,
        NomeResponsavel: input.NomeResponsavel,
    }

    if err := db.Create(&ong).Error; err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, ong)
}
