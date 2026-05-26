package models

import "github.com/google/uuid"

type OngRegiao string

const (
	OngRegiaoSerra     OngRegiao = "Serra"
	OngRegiaoVitoria   OngRegiao = "Vitória"
	OngRegiaoVilaVelha OngRegiao = "Vila-Velha"
	OngRegiaoCariacica OngRegiao = "Cariacica"
)

type Ong struct {
	BaseModel `swaggerignore:"true"`

	Nome            string    `gorm:"type:text;not null" json:"nome"`
	Endereco        string    `gorm:"type:text" json:"endereco"`
	Telefone        string    `gorm:"type:text" json:"telefone"`
	WhatsappLink    string    `gorm:"type:text" json:"whatsapp"`
	Descricao       string    `gorm:"type:text" json:"descricao"`
	NomeResponsavel string    `gorm:"type:text" json:"nome_responsavel"`
	Email           string    `gorm:"type:text;unique" json:"email"`
	Site			string    `gorm:"type:text" json:"site"`
	Instagram       string    `gorm:"type:text" json:"instagram"`
	Regiao          OngRegiao `gorm:"type:text" json:"area_de_atuacao"`
	Logo            string    `gorm:"type:text" json:"logo"`
	Termo           string    `gorm:"type:text" json:"termo"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
}