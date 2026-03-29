package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"ongpet/utils"

	"github.com/gin-gonic/gin"
)

type SendTestEmailRequest struct {
	To       string            `json:"to" binding:"required,email"`
	Template string            `json:"template" binding:"required"`
	Data     map[string]string `json:"data" binding:"required"`
}

// @Summary Envia email de teste
// @Description Envia um email de teste usando os templates configurados
// @Description Templates disponíveis: adoption_request, pet_registered, adoption_confirmed, contact
// @Tags Email
// @Accept json
// @Produce json
// @Param request body v1.SendTestEmailRequest true "Dados do email de teste"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/test/email [post]
func SendTestEmail(c *gin.Context) error {
	var req SendTestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	testData := utils.EmailTestData{
		To:       req.To,
		Template: req.Template,
		Data:     req.Data,
	}

	if err := utils.SendTestEmail(testData); err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email de teste enviado com sucesso",
		"to":      req.To,
		"template": req.Template,
	})
	return nil
}

// @Summary Valida configuração de email
// @Description Verifica se o mailer foi inicializado corretamente
// @Tags Email
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/test/email-config [get]
func CheckEmailConfig(c *gin.Context) error {
	mailer := utils.GetMailer()
	
	if mailer == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Mailer não inicializado",
		})
		return nil
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"message": "Email configurado corretamente",
		"from": "gustavodl.gdl33@gmail.com",
	})
	return nil
}
