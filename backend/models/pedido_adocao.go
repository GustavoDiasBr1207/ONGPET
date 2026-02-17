package models

import "github.com/google/uuid"

type PedidoAdocao struct {
	BaseModel `swaggerignore:"true"`

	OngID uuid.UUID `gorm:"type:uuid;not null" json:"ong_id"`
	Ong   Ong       `gorm:"foreignKey:OngID;references:ID" json:"-" swaggerignore:"true"`

	Respostas []RespostaFormulario `gorm:"foreignKey:PedidoAdocaoID" json:"respostas,omitempty"`
}
