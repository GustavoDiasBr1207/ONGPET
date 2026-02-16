package models

import "time"

type Pet struct {
	gorm.Model

	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	Nome    string `gorm:"type:text;not null"`
	Especie string `gorm:"type:text"`
	Raca    string `gorm:"type:text"`
	Idade   int    `gorm:"not null"`

	OngID uuid.UUID `gorm:"type:uuid;not null"`
	Ong   Ong       `gorm:"foreignKey:OngID;references:ID"`
}
