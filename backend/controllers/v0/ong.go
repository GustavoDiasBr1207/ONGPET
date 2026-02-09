package v0

import (
    "ongpet/database"
    "ongpet/models"
    "github.com/gin-gonic/gin"
)

// @Summary Lista todas as ONGs
// @Description Retorna todas as ONGs cadastradas
// @Tags ONG
// @Produce json
// @Success 200 {array} models.Ong
// @Router /api/v1/ongs [get]
func GetOngs(c *gin.Context) {
    db := database.GetDB()
    var ongs []models.Ong
    db.Find(&ongs)
    c.JSON(200, ongs)
}

// @Summary Cria uma nova ONG
// @Description Cria uma ONG no sistema
// @Tags ONG
// @Accept json
// @Produce json
// @Param ong body models.Ong true "Nova ONG"
// @Success 201 {object} models.Ong
// @Failure 400 {object} map[string]string
// @Router /api/v1/ongs [post]
func CreateOng(c *gin.Context) {
    db := database.GetDB()
    var ong models.Ong
    if err := c.ShouldBindJSON(&ong); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    db.Create(&ong)
    c.JSON(201, ong)
}
