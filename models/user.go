package models

import "time"

type User struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Username     string `json:"username" gorm:"uniqueIndex;size:60;not null"` // สำหรับ staff (admin/teacher)
	PasswordHash string `json:"-" gorm:"size:255;not null"`                   // << เพิ่มฟิลด์นี้                           // bcrypt hash
	Role         string `json:"role" gorm:"size:20;not null"`                 // "admin" | "teacher" | "parent"
	TeacherID    *uint  `json:"teacher_id" gorm:"index"`                      // << และฟิลด์นี้ (nullable)

	// โปรไฟล์ทั่วไป
	Email    string `json:"email" gorm:"size:120;index"`
	Phone    string `json:"phone" gorm:"size:20;index"`
	Timezone string `json:"timezone" gorm:"size:60"` // เช่น Asia/Bangkok
	Locale   string `json:"locale" gorm:"size:10"`   // เช่น "th"|"en"

	// บันทึกกิจกรรมบัญชี
	LastLogin          *time.Time `json:"last_login"`
	LastPasswordChange *time.Time `json:"last_password_change"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
