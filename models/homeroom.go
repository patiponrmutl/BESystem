package models

import "time"

type Homeroom struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	AcademicYear   string    `gorm:"size:4;not null" json:"academic_year"`    // พ.ศ. (เช่น "2568")
	EducationStage string    `gorm:"size:20;not null" json:"education_stage"` // อนุบาลศึกษา/ประถมศึกษา/มัธยมศึกษา
	Grade          string    `gorm:"size:20;not null" json:"grade"`           // เช่น "ประถม 6"
	Room           string    `gorm:"size:3;not null"  json:"room"`            // ตัวเลข ≤ 3 หลัก (string)
	Position       string    `gorm:"size:30;not null" json:"position"`        // ครูประจำชั้นหลัก/รอง
	TeacherID      uint      `gorm:"not null"         json:"teacher_id"`      // FK -> teachers.id (เชื่อมแบบ logic)
	Status         string    `gorm:"size:40;not null;default:'ปฏิบัติงาน'" json:"status"`
	Note           string    `gorm:"size:255"         json:"note"` // หมายเหตุ (optional)
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// หมายเหตุ: ใช้ unique ที่ชั้นเรียน+ตำแหน่ง เพื่อกันสร้างซ้ำ record เดิมของห้องเดียวกัน
// บังคับกติกา "ครูหลักห้ามซ้ำหลายห้อง" จะเช็คในโค้ด handler อีกชั้น
