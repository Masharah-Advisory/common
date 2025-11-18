package model

import (
	"time"
)

// Base provides common fields for all entities
type Base struct {
	ID        uint64     `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	CreatedBy *uint64    `json:"created_by,omitempty"`
	UpdatedBy *uint64    `json:"updated_by,omitempty"`
	DeletedBy *uint64    `json:"deleted_by,omitempty"`
}
