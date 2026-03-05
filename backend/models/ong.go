package models

import "github.com/google/uuid"

type Ong struct {
	BaseModel `swaggerignore:"true"`

	Nome            string `gorm:"type:text;not null" json:"nome"`
	Endereco        string `gorm:"type:text" json:"endereco"`
	Telefone        string `gorm:"type:text" json:"telefone"`
	NomeResponsavel string `gorm:"type:text" json:"nome_responsavel"`
	Email           string `gorm:"type:text;unique" json:"email"`
	//logo 		  	string `gorm:"type:text" json:"logo"`
	

	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
}