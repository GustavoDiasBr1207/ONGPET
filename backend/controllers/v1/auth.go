package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"ongpet/utils"

	"github.com/gin-gonic/gin"
)

// LoginRequest representa o corpo esperado para login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// @Summary Login
// @Description Faz login via Supabase e retorna tokens (access_token, etc.)
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Credenciais"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
	var reqBody LoginRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	if supabaseURL == "" || supabaseKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "supabase configuration missing"})
		return
	}

	payload := map[string]string{
		"email":    reqBody.Email,
		"password": reqBody.Password,
	}
	b, _ := json.Marshal(payload)

	url := supabaseURL + "/auth/v1/token?grant_type=password"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", supabaseKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to call supabase"})
		return
	}
	defer resp.Body.Close()

	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse supabase response"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, respBody)
		return
	}

	c.JSON(http.StatusOK, respBody)
}

// ─────────────────────────────────────────────────────────────
// EMAIL HANDLERS
// ─────────────────────────────────────────────────────────────

type SendTestEmailRequest struct {
	To       string            `json:"to" binding:"required,email"`
	Template string            `json:"template" binding:"required"`
	Data     map[string]string `json:"data" binding:"required"`
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
			"status":  "error",
			"message": "Mailer não inicializado",
		})
		return nil
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Email configurado corretamente",
		"from":    "amigofieladocao@gmail.com",
	})
	return nil
}

// @Summary Envia email de teste
// @Description Envia um email de teste usando os templates configurados
// @Description Templates disponíveis: adoption_request, pet_registered, adoption_confirmed, contact
// @Tags Email
// @Accept json
// @Produce json
// @Param request body SendTestEmailRequest true "Dados do email de teste"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/test/email [post]
func SendTestEmail(c *gin.Context) error {
	var req SendTestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	mailer := utils.GetMailer()
	if mailer == nil {
		return nil
	}

	var subject, body string

	switch req.Template {
	case "pet_registered":
		subject, body = utils.NewPetRegisteredEmail(
			req.Data["pet_name"],
			req.Data["owner_name"],
		)
	case "adoption_request":
		subject, body = utils.NewAdoptionRequestEmail(
			req.Data["pet_name"],
			req.Data["requester_name"],
			req.Data["ong_name"],
		)
	case "adoption_confirmed":
		subject, body = utils.NewAdoptionConfirmedEmail(
			req.Data["pet_name"],
			req.Data["requester_name"],
		)
	case "contact":
		subject, body = utils.NewContactEmail(
			req.Data["name"],
			req.Data["email"],
			req.Data["message"],
		)
	default:
		subject = req.Data["subject"]
		body = req.Data["body"]
	}

	if err := mailer.Send([]string{req.To}, subject, body); err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email de teste enviado com sucesso",
		"to":      req.To,
		"template": req.Template,
	})
	return nil
}
