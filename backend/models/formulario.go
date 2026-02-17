package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type FormularioModelo struct {
	BaseModel `swaggerignore:"true"`

	Nome   string            `gorm:"type:text;not null" json:"nome"`
	Campos []CampoFormulario `gorm:"foreignKey:FormularioModeloID" json:"campos,omitempty"`
}

type CampoFormulario struct {
	BaseModel `swaggerignore:"true"`

	FormularioModeloID uuid.UUID `gorm:"type:uuid;not null" json:"formularioModeloId"`

	Nome         string         `gorm:"type:text;not null" json:"nome"`
	Configuracao datatypes.JSON `gorm:"type:jsonb" json:"-" swaggerignore:"true"`

	FormularioModelo FormularioModelo `gorm:"foreignKey:FormularioModeloID" json:"-" swaggerignore:"true"`
}

type RespostaFormulario struct {
	BaseModel `swaggerignore:"true"`

	PedidoAdocaoID    uuid.UUID `gorm:"type:uuid;not null" json:"pedidoAdocaoId"`
	CampoFormularioID uuid.UUID `gorm:"type:uuid;not null" json:"campoFormularioId"`

	CampoFormulario CampoFormulario `gorm:"foreignKey:CampoFormularioID" json:"-" swaggerignore:"true"`

	Valor string `gorm:"type:text;not null" json:"valor"`
}
