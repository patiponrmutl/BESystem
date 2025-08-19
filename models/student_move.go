package models

import "time"

type StudentMove struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// อ้างถึงนักเรียน
	StudentID uint `gorm:"not null" json:"student_db_id"` // id ของ record ในตาราง students

	// จาก → ไป (เก็บแบบ string ให้ยืดหยุ่นกับรูปแบบชั้น/ปี)
	FromYear  string `gorm:"size:8;not null"  json:"from_year"` // มักเป็น พ.ศ. ที่ FE ส่งมา
	FromGrade string `gorm:"size:20;not null" json:"from_grade"`
	FromRoom  string `gorm:"size:5;not null"  json:"from_room"`

	ToYear  string `gorm:"size:8;not null"  json:"to_year"`
	ToGrade string `gorm:"size:20;not null" json:"to_grade"`
	ToRoom  string `gorm:"size:5;not null"  json:"to_room"`

	MoveDate time.Time `json:"move_date"`            // YYYY-MM-DD
	Note     string    `gorm:"size:255" json:"note"` // optional

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
