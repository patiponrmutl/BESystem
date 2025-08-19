package models

import "time"

// CalendarItem ใช้ตัวเดียว ครอบคลุม 3 ประเภท: normal / holiday / event
type CalendarItem struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Type string `json:"type" gorm:"type:varchar(20);index"` // normal | holiday | event
	Note string `json:"note" gorm:"type:varchar(200)"`

	// ----- NORMAL (เวลาเรียน/ภาคเรียน) -----
	Semester     string `json:"semester" gorm:"type:varchar(40)"`
	AcademicYear string `json:"academic_year" gorm:"type:varchar(10)"`
	OpenDate     string `json:"open_date" gorm:"type:date"`
	CloseDate    string `json:"close_date" gorm:"type:date"`
	TimeIn       string `json:"time_in" gorm:"type:varchar(5)"`
	TimeOut      string `json:"time_out" gorm:"type:varchar(5)"`

	// ----- HOLIDAY (วันหยุด) -----
	Name      string `json:"name" gorm:"type:varchar(80)"` // ถ้า “อื่นๆ” ให้เก็บข้อความจริงในช่องนี้เลย
	StartDate string `json:"start_date" gorm:"type:date"`
	EndDate   string `json:"end_date" gorm:"type:date"`

	// ----- EVENT (กิจกรรมพิเศษ) -----
	Title     string `json:"title" gorm:"type:varchar(80)"`
	Date      string `json:"date" gorm:"type:date"`
	StartTime string `json:"start_time" gorm:"type:varchar(5)"`
	EndTime   string `json:"end_time" gorm:"type:varchar(5)"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
