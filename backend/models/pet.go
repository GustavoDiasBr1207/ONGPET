package models

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Pet struct {
    BaseModel `swaggerignore:"true"`

    Nome      string  `gorm:"type:text;not null" json:"nome"`
    Especie   string  `gorm:"type:text" json:"especie"`
    Raca      string  `gorm:"type:text" json:"raca"`
    Idade     int     `gorm:"not null" json:"idade"`
    Descricao string  `gorm:"type:text" json:"descricao"`
    Peso      float64 `gorm:"not null" json:"peso"`
    Porte     string  `gorm:"type:text" json:"porte"`
    Regiao    string  `gorm:"type:text" json:"regiao"`

    // ARRAY PostgreSQL (text[])
    Imagens pq.StringArray `gorm:"type:text[]" json:"imagens" swaggertype:"array,string"`

    // OPCIONAL
    FormularioID *uuid.UUID `gorm:"type:uuid" json:"formulario_id"`

    OngID uuid.UUID `gorm:"type:uuid;not null" json:"ong_id"`
    Ong   Ong       `gorm:"foreignKey:OngID;references:ID" json:"-" swaggerignore:"true"`
}





