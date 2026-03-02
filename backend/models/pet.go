package models

import (
	"github.com/google/uuid"
)

type PetStatus string 

const (
	PetAdopted   PetStatus = "adopted"
	PetAvailable PetStatus = "available"
	PetDraft     PetStatus = "draft"
)

type Pet struct {
	BaseModel `swaggerignore:"true"`

	Nome      string    `gorm:"type:text" json:"nome"`
	Especie   string    `gorm:"type:text" json:"especie"`
	Raca      string    `gorm:"type:text" json:"raca"`
	Idade     int       `gorm:"default:0" json:"idade"`
	Descricao string    `gorm:"type:text" json:"descricao"`
	Peso      float64   `gorm:"default:0.0" json:"peso"`
	Porte     string    `gorm:"type:text" json:"porte"`
	Regiao    string    `gorm:"type:text" json:"regiao"`
	Status    PetStatus `gorm:"type:petstatus;default:'draft'" json:"status"`

	Imagens      []PetImage `gorm:"foreignKey:PetID" json:"imagens"`
	FormularioID *uuid.UUID `gorm:"type:uuid" json:"formulario_id"`

	OngID uuid.UUID `gorm:"type:uuid;not null" json:"ong_id"`
	Ong   Ong       `gorm:"foreignKey:OngID;references:ID" json:"-" swaggerignore:"true"`
}