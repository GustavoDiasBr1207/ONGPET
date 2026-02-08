package models

import "time"

type PedidoAdocao struct {
	ID        uint                 `gorm:"primaryKey" json:"id"`
	OngID     uint                 `json:"ong_id"`
	Ong       Ong                  `gorm:"foreignKey:OngID" json:"-" swaggerignore:"true"`
	Respostas []RespostaFormulario `gorm:"foreignKey:PedidoAdocaoID" json:"respostas,omitempty"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}
