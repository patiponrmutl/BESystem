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

	// ✅ สร้างตาราง schools ให้อัตโนมัติ
	if err := DB.AutoMigrate(&models.School{}); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}
}
