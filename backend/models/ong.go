package models

import "time"

type Ong struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Nome          string         `gorm:"column:nome;size:255;not null" json:"nome"`
	Endereco      string         `gorm:"column:endereco;size:255" json:"endereco"`
	Telefone      string         `gorm:"column:telefone;size:50" json:"telefone"`
	NomeResponsavel string       `gorm:"column:nome_responsavel;size:255" json:"nome_responsavel"`
	Email         string         `gorm:"column:email;size:100" json:"email"`
	PedidosAdocao []PedidoAdocao `gorm:"foreignKey:OngID" json:"pedidosAdocao,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
