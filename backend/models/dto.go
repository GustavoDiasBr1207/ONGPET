package models

import "github.com/google/uuid"

type CampoFormularioDTO struct {
	ID           uuid.UUID `json:"id"`
	Nome         string    `json:"nome"`
	Configuracao string    `json:"configuracao"` // string simplificada
}

type RespostaFormularioDTO struct {
	ID              uuid.UUID           `json:"id"`
	CampoFormulario CampoFormularioDTO  `json:"campoFormulario"`
	Valor           string              `json:"valor"`
}

type PedidoAdocaoDTO struct {
	ID        uuid.UUID               `json:"id"`
	OngID     uuid.UUID               `json:"ong_id"`
	Respostas []RespostaFormularioDTO `json:"respostas"`
}
