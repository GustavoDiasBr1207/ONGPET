package controllers

import "github.com/gin-gonic/gin"

type AppHandler func(*gin.Context) error

func Handle(h AppHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {
			// Aqui você pode centralizar o tratamento de erro
			c.Error(err)
		}
	}
}
