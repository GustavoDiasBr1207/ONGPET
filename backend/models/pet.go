package models

import "time"

type Pet struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Nome      string    `gorm:"size:100;not null" json:"nome"`
	Especie   string    `gorm:"size:50" json:"especie"`
	Raca      string    `gorm:"size:50" json:"raca"`
	Idade     int       `json:"idade"`
	OngID     uint      `json:"ong_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
