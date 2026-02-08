package models

import "time"

type PedidoAdocao struct {
	ID           uint64    `gorm:"primaryKey"`
	Status       string    `gorm:"size:20;not null;default:'pendente';check:status IN ('pendente','aprovado','rejeitado')"`
	DataCriacao  time.Time `gorm:"autoCreateTime"`

	PetID uint64 `gorm:"not null;index"`
	Pet   Pet    `gorm:"foreignKey:PetID"`

	OngID uint64 `gorm:"not null"`
	Ong   Ong    `gorm:"foreignKey:OngID"`

	Respostas []RespostaFormulario
}
