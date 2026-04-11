package models

import (
	"time"

	"github.com/google/uuid"
)

type AcompanhamentoStatus string

const (
	AcompanhamentoAtivo      AcompanhamentoStatus = "ativo"
	AcompanhamentoFinalizado AcompanhamentoStatus = "finalizado"
)

type TrackingFrequency string

const (
	FrequencyUnique    TrackingFrequency = "UNIQUE"
	FrequencyMonthly   TrackingFrequency = "MONTHLY"
	FrequencyQuarterly TrackingFrequency = "QUARTERLY"
	FrequencyBiannual  TrackingFrequency = "BIANNUAL"
)

// Acompanhamento representa a configuração master do pós-adoção de um pet
type Acompanhamento struct {
	BaseModel `swaggerignore:"true"`

	PetID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"pet_id"`
	Pet   Pet       `gorm:"foreignKey:PetID;references:ID" json:"-" swaggerignore:"true"`

	OngID uuid.UUID `gorm:"type:uuid;not null;index" json:"ong_id"`
	Ong   Ong       `gorm:"foreignKey:OngID;references:ID" json:"-" swaggerignore:"true"`

	AdotanteNome     string            `gorm:"not null" json:"adotante_nome"`
	AdotanteTelefone string            `gorm:"not null" json:"adotante_telefone"`
	AdotanteEmail    string            `json:"adotante_email"`
	
	Frequencia  TrackingFrequency    `gorm:"type:text;not null" json:"frequencia"`
	ProximaData *time.Time           `json:"proxima_data"`
	Status      AcompanhamentoStatus `gorm:"type:text;default:'ativo'" json:"status"`

	Logs []LogAcompanhamento `gorm:"foreignKey:AcompanhamentoID" json:"logs,omitempty"`
}

// LogAcompanhamento representa um registro individual de contato realizado
type LogAcompanhamento struct {
	BaseModel `swaggerignore:"true"`

	AcompanhamentoID uuid.UUID      `gorm:"type:uuid;not null;index" json:"acompanhamento_id"`
	Acompanhamento   Acompanhamento `gorm:"foreignKey:AcompanhamentoID;references:ID" json:"-" swaggerignore:"true"`

	DataContato time.Time `gorm:"not null" json:"data_contato"`
	Notas       string    `gorm:"type:text;not null" json:"notas"`
}
