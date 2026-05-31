package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base is embedded in every entity. It provides a UUID primary key, automatic
// timestamps and soft-delete support (DeletedAt is set instead of hard delete).
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate assigns a UUID when one was not supplied.
func (b *Base) BeforeCreate(*gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
