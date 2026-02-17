package v0

import (
	"github.com/gin-gonic/gin"
	"ongpet/database"
	"ongpet/models"
)

// @Summary Lista todos os Pets
// @Description Retorna todos os pets cadastrados
// @Tags Pet
// @Produce json
// @Success 200 {array} models.Pet
// @Router /api/v1/pets [get]
func GetPets(c *gin.Context) {
	db := database.GetDB()
	var pets []models.Pet
	db.Find(&pets)
	c.JSON(200, pets)
}

// @Summary Cria um novo Pet
// @Description Cria um Pet no sistema
// @Tags Pet
// @Accept json
// @Produce json
// @Param pet body models.Pet true "Novo Pet"
// @Success 201 {object} models.Pet
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/v1/pets [post]
func CreatePet(c *gin.Context) {
	db := database.GetDB()
	var pet models.Pet
	if err := c.ShouldBindJSON(&pet); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	db.Create(&pet)
	c.JSON(201, pet)
}
