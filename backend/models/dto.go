package models

import "github.com/google/uuid"

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
	Respostas []RespostaFormularioDTO `json:"respostas"`
}