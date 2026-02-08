package models

type FormularioModelo struct {
	ID    uint64 `gorm:"primaryKey"`
	Nome  string `gorm:"size:100;not null"`
	Ativo bool   `gorm:"default:true"`

	OngID uint64 `gorm:"not null"`
	Ong   Ong    `gorm:"foreignKey:OngID;constraint:OnDelete:CASCADE"`

	Campos []CampoFormulario
}
