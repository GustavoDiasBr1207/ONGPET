package models

import "gorm.io/datatypes"

type CampoFormulario struct {
	ID          uint64 `gorm:"primaryKey"`
	NomeCampo   string `gorm:"size:100;not null"`
	Label       string `gorm:"size:150;not null"`
	Tipo        string `gorm:"size:20;not null;check:tipo IN ('text','number','email','select','textarea')"`
	Obrigatorio bool   `gorm:"default:false"`
	Ordem       *int

	Configuracao datatypes.JSON `gorm:"type:jsonb"`

	FormularioModeloID uint64            `gorm:"not null"`
	FormularioModelo   FormularioModelo `gorm:"foreignKey:FormularioModeloID;constraint:OnDelete:CASCADE"`

	Respostas []RespostaFormulario
}
