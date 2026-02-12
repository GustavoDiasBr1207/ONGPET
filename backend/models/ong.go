package models

import "time"

type Ong struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Nome          string         `gorm:"size:255;not null" json:"nome"`
	Endereco      string         `gorm:"size:255" json:"endereco"`
	Telefone      string         `gorm:"size:50" json:"telefone"`
	NomeResponsavel string       `gorm:"size:255" json:"nome_responsavel"`
	Email         string         `gorm:"size:100" json:"email"`
	PedidosAdocao []PedidoAdocao `gorm:"foreignKey:OngID" json:"pedidosAdocao,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
