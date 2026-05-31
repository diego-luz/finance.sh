package entities

import (
	"time"

	"github.com/google/uuid"
)

// Budget caps spending for a category in a given month.
type Budget struct {
	Base
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	CategoryID     uuid.UUID `gorm:"type:uuid;not null;index" json:"category_id"`
	Amount         int64     `gorm:"not null" json:"amount"`
	Month          int       `gorm:"not null" json:"month"`
	Year           int       `gorm:"not null" json:"year"`
}

func (Budget) TableName() string { return "budgets" }

// Goal is a savings target the user works toward.
type Goal struct {
	Base
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string     `gorm:"not null" json:"name"`
	TargetAmount   int64      `gorm:"not null" json:"target_amount"`
	CurrentAmount  int64      `gorm:"default:0" json:"current_amount"`
	Deadline       *time.Time `json:"deadline,omitempty"`
	Color          string     `gorm:"type:varchar(9);default:'#10b981'" json:"color"`
}

func (Goal) TableName() string { return "goals" }

// Notification is an in-app alert (also delivered via websocket / email).
type Notification struct {
	Base
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	UserID         *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Type           string     `gorm:"type:varchar(30)" json:"type"`
	Title          string     `gorm:"not null" json:"title"`
	Message        string     `json:"message"`
	Read           bool       `gorm:"default:false" json:"read"`
}

func (Notification) TableName() string { return "notifications" }

// AuditLog records security-relevant actions for compliance.
type AuditLog struct {
	Base
	OrganizationID *uuid.UUID `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	UserID         *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action         string     `gorm:"not null" json:"action"`
	Entity         string     `json:"entity,omitempty"`
	EntityID       string     `json:"entity_id,omitempty"`
	IP             string     `json:"ip,omitempty"`
	Metadata       string     `gorm:"type:jsonb" json:"metadata,omitempty"`
}

func (AuditLog) TableName() string { return "audit_logs" }
