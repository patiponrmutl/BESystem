package models

import "time"

type Parent struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"uniqueIndex;size:120"`
	Phone     string    `json:"phone" gorm:"size:20"`
	Password  string    `json:"-" gorm:"not null"` // bcrypt hash
	PdpaOK    bool      `json:"pdpa_ok" gorm:"not null;default:false"`
	Name      string    `json:"name" gorm:"size:120"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
