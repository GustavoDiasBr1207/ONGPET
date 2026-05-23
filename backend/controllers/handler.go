package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AppHandler func(*gin.Context) error

func Handle(h AppHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {
			c.JSON(HttpStatusForError(err.Error()), gin.H{"error": err.Error()})
		}
	}
}

// HttpStatusForError mapeia mensagens de erro ao código HTTP adequado.
// Exportado para reutilização nos testes.
func HttpStatusForError(msg string) int {
	switch msg {
	// 404
	case
		"ONG não encontrada",
		"Pet não encontrado",
		"Imagem não encontrada",
		"Formulário não encontrado",
		"Campo não encontrado",
		"Pedido de adoção não encontrado",
		"Resposta não encontrada",
		"Banner não encontrado":
		return http.StatusNotFound

	// 409
	case
		"email já cadastrado",
		"usuário já possui uma ONG cadastrada":
		return http.StatusConflict

	// 401
	case
		"usuário não autenticado",
		"user_id inválido no token":
		return http.StatusUnauthorized

	// 400 — validações de input
	case
		"ID da ONG inválido", "ONG não possui logo",
		"ID do pet inválido", "ID da imagem inválido",
		"ong_id é obrigatório", "o pet pode ter no máximo 5 imagens",
		"ID do formulário inválido", "ID do campo inválido",
		"nome é obrigatório", "configuracao.label é obrigatório",
		"configuracao.tipo é obrigatório", "o formulário não possui imagem cadastrada",
		"ID do pedido inválido", "ID da resposta inválido",
		"pet_id é obrigatório",
		"status inválido. Use: pendente, aprovado, rejeitado ou cancelado",
		"ID de acompanhamento inválido",
		"ID do banner inválido", "este banner não possui imagem",
		"imagem é obrigatória", "title é obrigatório",
		"instagram_url é obrigatório", "start_at é obrigatório", "end_at é obrigatório",
		"start_at deve estar no formato RFC3339 (ex: 2026-03-11T15:04:05Z)",
		"end_at deve estar no formato RFC3339 (ex: 2026-03-11T15:04:05Z)",
		"start_at deve estar no formato RFC3339",
		"end_at deve estar no formato RFC3339",
		"end_at deve ser posterior a start_at",
		"position não pode ser negativo",
		"title não pode ser vazio",
		"instagram_url não pode ser vazia",
		"limit excede o máximo permitido de 100",
		"page inválido", "limit inválido",
		"nenhuma imagem enviada", "tipo de imagem inválido":
		return http.StatusBadRequest
	}

	if strings.HasPrefix(msg, "body inválido") ||
		strings.HasPrefix(msg, "invalid character") ||
		strings.HasPrefix(msg, "invalid syntax") ||
		strings.HasPrefix(msg, "Key:") ||
		msg == "EOF" {
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}
