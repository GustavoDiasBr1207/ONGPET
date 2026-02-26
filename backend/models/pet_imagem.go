package models

import (
	"github.com/google/uuid"
)

type PetImage struct {
	BaseModel `swaggerignore:"true"`

	URL      string    `gorm:"not null" json:"url"`
	Position int       `gorm:"not null" json:"position"` // ordem no frontend
	PetID    uuid.UUID `gorm:"type:uuid;not null" json:"pet_id"`
}