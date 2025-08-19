package models

import "time"

type Teacher struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TeacherCode string    `gorm:"size:20;not null;uniqueIndex" json:"teacher_code"`
	Prefix      string    `gorm:"size:20;not null" json:"prefix"`
	FirstName   string    `gorm:"size:50;not null" json:"first_name"`
	LastName    string    `gorm:"size:50;not null" json:"last_name"`
	Phone       string    `gorm:"size:15;not null" json:"phone"`
	Email       string    `gorm:"size:50;not null;uniqueIndex" json:"email"`
	Position    string    `gorm:"size:50;not null" json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
