package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FormularioModelo struct {
	gorm.Model

	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Nome   string    `gorm:"type:text;not null" json:"nome"`
	Campos []CampoFormulario `gorm:"foreignKey:FormularioModeloID" json:"campos,omitempty"`
}

type CampoFormulario struct {
	gorm.Model

	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FormularioModeloID uuid.UUID `gorm:"type:uuid;not null" json:"formularioModeloId"`

	Nome         string         `gorm:"type:text;not null" json:"nome"`
	Configuracao datatypes.JSON `gorm:"type:jsonb" json:"-" swaggerignore:"true"`

	FormularioModelo FormularioModelo `gorm:"foreignKey:FormularioModeloID" json:"-" swaggerignore:"true"`
}

type RespostaFormulario struct {
	gorm.Model

	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	PedidoAdocaoID    uuid.UUID `gorm:"type:uuid;not null" json:"pedidoAdocaoId"`
	CampoFormularioID uuid.UUID `gorm:"type:uuid;not null" json:"campoFormularioId"`

	CampoFormulario CampoFormulario `gorm:"foreignKey:CampoFormularioID" json:"-" swaggerignore:"true"`

	Valor string `gorm:"type:text;not null" json:"valor"`
}
