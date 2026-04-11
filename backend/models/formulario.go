package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ─────────────────────────────────────────────────────────────
// CONFIGURACAO DO CAMPO (serializada como JSONB no banco)
// ─────────────────────────────────────────────────────────────

type TipoCampo string

const (
	TipoCampoTexto    TipoCampo = "texto"
	TipoCampoNumero   TipoCampo = "numero"
	TipoCampoSelect   TipoCampo = "select"
	TipoCampoRadio    TipoCampo = "radio"
	TipoCampoCheckbox TipoCampo = "checkbox"
	TipoCampoTextarea TipoCampo = "textarea"
	TipoCampoEmail    TipoCampo = "email"
	TipoCampoTelefone TipoCampo = "telefone"
	TipoCampoData     TipoCampo = "data"
	TipoCampoImagem   TipoCampo = "imagem" // adotante envia imagem como resposta
)

// CampoConfiguracao é a struct que representa o JSON armazenado em CampoFormulario.Configuracao
type CampoConfiguracao struct {
	// Apresentação
	Label       string `json:"label"`                 // ex: "Foto da casa"
	Placeholder string `json:"placeholder,omitempty"`
	Descricao   string `json:"descricao,omitempty"`

	// Tipo
	Tipo TipoCampo `json:"tipo"` // texto | numero | select | radio | checkbox | textarea | email | telefone | data | imagem

	// Opções (apenas para select, radio, checkbox)
	Opcoes []string `json:"opcoes,omitempty"`

	// Validações
	Obrigatorio bool     `json:"obrigatorio"`
	Ativo       bool     `json:"ativo,omitempty"`
	Min         *int     `json:"min,omitempty"`
	Max         *int     `json:"max,omitempty"`
	Regex       string   `json:"regex,omitempty"`
	MinValor    *float64 `json:"min_valor,omitempty"`
	MaxValor    *float64 `json:"max_valor,omitempty"`
}

// ─────────────────────────────────────────────────────────────
// FORMULARIO MODELO
// ─────────────────────────────────────────────────────────────

type FormularioModelo struct {
	BaseModel

	OngID uuid.UUID `gorm:"type:uuid;not null" json:"ong_id"`
	Ong   Ong       `gorm:"foreignKey:OngID;references:ID" json:"-" swaggerignore:"true"`

	Nome      string            `gorm:"type:text;not null" json:"nome"`
	ImagemURL string            `gorm:"type:text" json:"imagem_url,omitempty"`
	Campos    []CampoFormulario `gorm:"foreignKey:FormularioModeloID;constraint:OnDelete:CASCADE" json:"campos,omitempty"`
}

// ─────────────────────────────────────────────────────────────
// CAMPO FORMULARIO
// ─────────────────────────────────────────────────────────────

type CampoFormulario struct {
	BaseModel

	FormularioModeloID *uuid.UUID `gorm:"type:uuid" json:"formulario_modelo_id"`

	Nome         string         `gorm:"type:text;not null" json:"nome"`
	Ordem        int            `gorm:"default:0" json:"ordem"`
	Configuracao datatypes.JSON `gorm:"type:jsonb;not null" json:"configuracao" swaggertype:"string"`

	FormularioModelo FormularioModelo `gorm:"foreignKey:FormularioModeloID" json:"-" swaggerignore:"true"`
}

// ─────────────────────────────────────────────────────────────
// RESPOSTA FORMULARIO
// ─────────────────────────────────────────────────────────────

type RespostaFormulario struct {
	BaseModel

	PedidoAdocaoID    uuid.UUID `gorm:"type:uuid;not null" json:"pedido_adocao_id"`
	CampoFormularioID uuid.UUID `gorm:"type:uuid;not null" json:"campo_formulario_id"`

	CampoFormulario CampoFormulario `gorm:"foreignKey:CampoFormularioID" json:"campo,omitempty" swaggerignore:"true"`

	// Para tipo "imagem": começa vazio ("") e é preenchido via
	// POST /pedidos-adocao/:pedidoId/respostas/:respostaId/imagem
	Valor string `gorm:"type:text;not null;default:''" json:"valor"`
}