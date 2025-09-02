// models/user.go
package models

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"size:50;uniqueIndex;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	Role         string `gorm:"size:20;not null;index"` // admin | teacher | parent (ถ้ามี)
	TeacherID    *uint  `gorm:"index"`                  // null ได้ ถ้าเป็น admin
	Email        string `gorm:"size:120"`
	Phone        string `gorm:"size:30"`
	Timezone     string `gorm:"size:64"`
	Locale       string `gorm:"size:16"`

	// ===== ฟิลด์ที่ใช้ใน TeacherAccountHandler =====
	Enabled             bool `gorm:"not null;default:true"`  // เปิด/ปิดการใช้งาน
	ForcePasswordChange bool `gorm:"not null;default:false"` // ให้เปลี่ยนรหัสครั้งถัดไป

	LastLogin          *time.Time
	LastPasswordChange *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
