package models

import (
	"time"

	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────
// DTOs
// ─────────────────────────────────────────────────────────────

type CampoFormularioDTO struct {
	ID           uuid.UUID         `json:"id"`
	Nome         string            `json:"nome"`
	Ordem        int               `json:"ordem"`
	Configuracao CampoConfiguracao `json:"configuracao"`
}

type RespostaFormularioDTO struct {
	ID    uuid.UUID          `json:"id"`
	Campo CampoFormularioDTO `json:"campo"`
	Valor string             `json:"valor"`
}

type PedidoAdocaoDTO struct {
	ID        uuid.UUID               `json:"id"`
	OngID     uuid.UUID               `json:"ong_id"`
	PetID     uuid.UUID               `json:"pet_id"`
	Status    PedidoAdocaoStatus      `json:"status"`
	CreatedAt time.Time               `json:"created_at"`
	Respostas []RespostaFormularioDTO `json:"respostas"`
}

type AcompanhamentoDTO struct {
	ID               uuid.UUID            `json:"id"`
	PetID            uuid.UUID            `json:"pet_id"`
	PetNome          string               `json:"pet_nome,omitempty"`
	AdotanteNome     string               `json:"adotante_nome"`
	AdotanteTelefone string               `json:"adotante_telefone"`
	AdotanteEmail    string               `json:"adotante_email"`
	Frequencia       TrackingFrequency    `json:"frequencia"`
	ProximaData      *time.Time           `json:"proxima_data"`
	Status           AcompanhamentoStatus `json:"status"`
	CreatedAt        time.Time            `json:"created_at"`
	Logs             []LogAcompanhamentoDTO `json:"logs,omitempty"`
}

type LogAcompanhamentoDTO struct {
	ID               uuid.UUID `json:"id"`
	AcompanhamentoID uuid.UUID `json:"acompanhamento_id"`
	DataContato      time.Time `json:"data_contato"`
	Notas            string    `json:"notas"`
	CreatedAt        time.Time `json:"created_at"`
}