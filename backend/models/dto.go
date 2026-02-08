package models

type CampoFormularioDTO struct {
	ID           uint   `json:"id"`
	Nome         string `json:"nome"`
	Configuracao string `json:"configuracao"` // string simplificada
}

type RespostaFormularioDTO struct {
	ID              uint              `json:"id"`
	CampoFormulario CampoFormularioDTO `json:"campoFormulario"`
	Valor           string            `json:"valor"`
}

type PedidoAdocaoDTO struct {
	ID        uint                    `json:"id"`
	OngID     uint                    `json:"ong_id"`
	Respostas []RespostaFormularioDTO `json:"respostas"`
}
