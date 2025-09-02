package database

import (
	"log"

	"github.com/patiponrmutl/BESystem/config"
	"github.com/patiponrmutl/BESystem/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(cfg *config.Config) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	DB = db

	// ----- AutoMigrate โครงสร้างทั้งหมดของเรา -----
	if err := DB.AutoMigrate(
		&models.School{},
		&models.Student{},
		&models.Teacher{},
		&models.Homeroom{},
		&models.StudentMove{},  // ✅ การย้ายนักเรียน (ครั้งเดียวพอ)
		&models.CalendarItem{}, // ✅ ปฏิทินการศึกษา
		&models.Attendance{},
		&models.User{},
		&models.Parent{},
		&models.LeaveRequest{},
	); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// ----- ลบคอลัมน์ legacy: users.password (เราใช้เฉพาะ password_hash แล้ว) -----
	// ทำแบบปลอดภัย: เช็คก่อนว่ามีคอลัมน์ค้างอยู่ไหม
	if DB.Migrator().HasColumn(&models.User{}, "password") {
		if err := DB.Migrator().DropColumn(&models.User{}, "password"); err != nil {
			log.Printf("[migrate] warn: drop users.password failed: %v", err)
		} else {
			log.Printf("[migrate] dropped legacy column users.password")
		}
	}
}
