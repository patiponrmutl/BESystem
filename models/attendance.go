package models

import "time"

// บันทึกการเข้า-ออก/สถานะรายวันของนักเรียน
type Attendance struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	StudentID uint   `json:"student_id" gorm:"index;not null"`
	Date      string `json:"date" gorm:"size:10;not null"`   // YYYY-MM-DD
	Time      string `json:"time" gorm:"size:5"`             // HH:MM (ถ้ามี)
	Status    string `json:"status" gorm:"size:20;not null"` // เข้า/ออก/มาสาย/ขาด/ลา
	Note      string `json:"note" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
