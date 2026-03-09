package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppHandler func(*gin.Context) error

func Handle(h AppHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {
			// Retorna erro HTTP em vez de apenas registrar
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
	}
}
