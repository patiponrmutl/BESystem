package models

import "time"

type Student struct {
	ID         uint       `gorm:"primaryKey"            json:"id"`
	NationalID string     `gorm:"size:13;not null"      json:"national_id"`       // เลขบัตร
	StudentID  string     `gorm:"size:20;uniqueIndex;not null" json:"student_id"` // รหัสนักเรียน (แสดงในตาราง)
	Prefix     string     `gorm:"size:20;not null"      json:"prefix"`            // คำนำหน้า
	FirstName  string     `gorm:"size:50;not null"      json:"first_name"`
	LastName   string     `gorm:"size:50;not null"      json:"last_name"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
	Education  string     `gorm:"size:50;not null"      json:"education_stage"` // ช่วงชั้น/ระดับ
	Grade      string     `gorm:"size:20;not null"      json:"grade"`
	Room       string     `gorm:"size:10;not null"      json:"room"`
	Address    string     `gorm:"type:text;not null"    json:"address"`
	Phone      string     `gorm:"size:15;not null"      json:"phone"`
	Status     string     `gorm:"size:20;not null"      json:"status"` // เช่น active|left|suspended
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
