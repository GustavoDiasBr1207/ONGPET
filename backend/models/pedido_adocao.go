package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PedidoAdocao struct {
	gorm.Model

	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	OngID uuid.UUID `gorm:"type:uuid;not null" json:"ong_id"`
	Ong   Ong       `gorm:"foreignKey:OngID;references:ID" json:"-" swaggerignore:"true"`

	Respostas []RespostaFormulario `gorm:"foreignKey:PedidoAdocaoID" json:"respostas,omitempty"`
}
