package models

import "time"

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:60;not null"`
	Password  string    `json:"-" gorm:"not null"`            // เก็บ bcrypt hash
	Role      string    `json:"role" gorm:"size:20;not null"` // "admin" | "teacher"
	Name      string    `json:"name" gorm:"size:120"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
