package models

import "time"

type LeaveRequest struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	StudentID    uint       `json:"student_id" gorm:"index;not null"`
	Type         string     `json:"type" gorm:"size:40;not null"`      // ป่วย/ธุระส่วนตัว/อื่นๆ
	Reason       string     `json:"reason" gorm:"type:text"`           // เหตุผล (ค้นหาได้)
	DateFrom     string     `json:"date_from" gorm:"size:10;not null"` // YYYY-MM-DD
	DateTo       string     `json:"date_to" gorm:"size:10;not null"`   // YYYY-MM-DD
	Attachments  int        `json:"attachments" gorm:"default:0"`      // จำนวนไฟล์แนบ
	Status       string     `json:"status" gorm:"size:20;not null"`    // รออนุมัติ/อนุมัติ/ปฏิเสธ
	SubmittedAt  time.Time  `json:"submitted_at" gorm:"autoCreateTime"`
	DecidedAt    *time.Time `json:"decided_at"`
	DecidedBy    *uint      `json:"decided_by"` // user_id ของครูที่อนุมัติ/ปฏิเสธ
	RejectReason string     `json:"reject_reason" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
