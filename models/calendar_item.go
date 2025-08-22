package models

import "time"

// CalendarItem ครอบคลุม 3 ประเภท: normal / holiday / event
type CalendarItem struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Type string `json:"type" gorm:"type:varchar(20);index"` // normal | holiday | event
	Note string `json:"note" gorm:"type:varchar(200)"`

	// ----- NORMAL -----
	Semester     string `json:"semester" gorm:"type:varchar(40)"`
	AcademicYear string `json:"academic_year" gorm:"type:varchar(10)"`
	OpenDate     string `json:"open_date" gorm:"type:varchar(10)"`  // YYYY-MM-DD (string)
	CloseDate    string `json:"close_date" gorm:"type:varchar(10)"` // YYYY-MM-DD (string)
	TimeIn       string `json:"time_in" gorm:"type:varchar(5)"`
	TimeOut      string `json:"time_out" gorm:"type:varchar(5)"`

	// ----- HOLIDAY -----
	Name      string `json:"name" gorm:"type:varchar(80)"`
	StartDate string `json:"start_date" gorm:"type:varchar(10)"` // YYYY-MM-DD (string)
	EndDate   string `json:"end_date" gorm:"type:varchar(10)"`   // YYYY-MM-DD (string, อาจว่าง)

	// ----- EVENT -----
	Title     string `json:"title" gorm:"type:varchar(80)"`
	Date      string `json:"date" gorm:"type:varchar(10)"` // YYYY-MM-DD (string, อาจว่าง)
	StartTime string `json:"start_time" gorm:"type:varchar(5)"`
	EndTime   string `json:"end_time" gorm:"type:varchar(5)"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
