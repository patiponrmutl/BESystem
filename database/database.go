package database

import (
	"log"

	"github.com/patiponrmutl/BESystem/config"
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

	// ถ้ายังไม่มีโมเดล ไม่ต้อง AutoMigrate อะไรในขั้นนี้
	// ตัวอย่างต่อไป: DB.AutoMigrate(&models.User{})
}
