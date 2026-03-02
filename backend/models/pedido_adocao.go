package models

import "github.com/google/uuid"

type PedidoAdocaoStatus string

const (
	PedidoAdocaoPendente  PedidoAdocaoStatus = "pendente"
	PedidoAdocaoAprovado  PedidoAdocaoStatus = "aprovado"
	PedidoAdocaoRejeitado PedidoAdocaoStatus = "rejeitado"
	PedidoAdocaoCancelado PedidoAdocaoStatus = "cancelado"
)

type PedidoAdocao struct {
	BaseModel `swaggerignore:"true"`

	OngID uuid.UUID `gorm:"type:uuid;not null" json:"ong_id"`
	Ong   Ong       `gorm:"foreignKey:OngID;references:ID" json:"-" swaggerignore:"true"`

	PetID uuid.UUID `gorm:"type:uuid;not null" json:"pet_id"`
	Pet   Pet       `gorm:"foreignKey:PetID;references:ID" json:"-" swaggerignore:"true"`

	Status PedidoAdocaoStatus `gorm:"type:text;default:'pendente'" json:"status"`

	Respostas []RespostaFormulario `gorm:"foreignKey:PedidoAdocaoID" json:"respostas,omitempty"`
}