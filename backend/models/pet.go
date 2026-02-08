package models

type Pet struct {
	ID         uint64  `gorm:"primaryKey"`
	Nome       string  `gorm:"size:100;not null"`
	Sexo       string  `gorm:"size:1;not null;check:sexo IN ('M','F')"`
	Especie    string  `gorm:"size:20;not null;check:especie IN ('Cachorro','Gato','Outros')"`
	Porte      *string `gorm:"size:10;check:porte IN ('pequeno','medio','grande')"`
	Peso       *float64
	Disponivel bool   `gorm:"default:true"`

	OngID uint64 `gorm:"not null;index"`
	Ong   Ong    `gorm:"foreignKey:OngID;constraint:OnDelete:CASCADE"`
}
