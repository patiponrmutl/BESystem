package models

import "time"

type School struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	SchoolCode     string `gorm:"uniqueIndex;size:20;not null" json:"school_code"`
	SchoolName     string `gorm:"size:100;not null" json:"school_name"`
	Address        string `gorm:"size:255;not null" json:"address"`
	Phone          string `gorm:"size:20;not null" json:"phone"`
	EducationLevel string `gorm:"size:50;not null" json:"education_level"` // อนุบาลศึกษา/ประถมศึกษา/มัธยมศึกษา/ทุกระดับการสอน

	// ✅ ใหม่: ให้ตรง UI (0 = ไม่บังคับ, 1..20)
	TeacherCodeDigits int `gorm:"default:0;not null" json:"teacher_code_digits"`
	StudentCodeDigits int `gorm:"default:0;not null" json:"student_code_digits"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
