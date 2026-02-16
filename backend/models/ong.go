package models

import "time"

type Ong struct {
	gorm.Model

	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	Nome             string `gorm:"type:text;not null"`
	Endereco         string `gorm:"type:text"`
	Telefone         string `gorm:"type:text"`
	NomeResponsavel  string `gorm:"type:text"`
	Email            string `gorm:"type:text;unique"`
}
