package models

import (
	"time"

	"gorm.io/datatypes"
)

type FormularioModelo struct {
	ID        uint              `gorm:"primaryKey" json:"id"`
	Nome      string            `json:"nome"`
	Campos    []CampoFormulario `gorm:"foreignKey:FormularioModeloID" json:"campos,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type CampoFormulario struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	FormularioModeloID uint           `json:"formularioModeloId"`
	Nome               string         `json:"nome"`
	Configuracao       datatypes.JSON `json:"-" swaggerignore:"true"` // ignorado no Swagger
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type RespostaFormulario struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	PedidoAdocaoID   uint           `json:"pedidoAdocaoId"`
	CampoFormularioID uint          `json:"campoFormularioId"`
	CampoFormulario  CampoFormulario `gorm:"foreignKey:CampoFormularioID" json:"-" swaggerignore:"true"`
	Valor            string         `json:"valor"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}
