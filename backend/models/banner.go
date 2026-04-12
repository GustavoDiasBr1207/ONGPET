package models

import (
	"time"

	"github.com/google/uuid"
)

// ─────────────────────────────────────────────────────────────
// BANNER STATUS
// ─────────────────────────────────────────────────────────────

type BannerStatus string

const (
	BannerStatusActive   BannerStatus = "active"
	BannerStatusInactive BannerStatus = "inactive"
	BannerStatusDraft    BannerStatus = "draft"
)

// ─────────────────────────────────────────────────────────────
// BANNER
// ─────────────────────────────────────────────────────────────

type Banner struct {
	BaseModel

	Title        string `gorm:"type:text;not null" json:"title"`
	InstagramURL string `gorm:"type:text"          json:"instagram_url"`
	ImageURL     string `gorm:"type:text"          json:"image_url"`

	StartAt time.Time `gorm:"not null" json:"start_at"`
	EndAt   time.Time `gorm:"not null" json:"end_at"`

	Active   bool `gorm:"default:true" json:"active"`
	Position int  `gorm:"default:0"    json:"position"`

	OngID *uuid.UUID `gorm:"type:uuid" json:"ong_id"`
	Ong   Ong        `gorm:"foreignKey:OngID;references:ID" json:"-" swaggerignore:"true"`
}
