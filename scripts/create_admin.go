// scripts/create_admin.go
package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/patiponrmutl/BESystem/config"
	"github.com/patiponrmutl/BESystem/database"
	"github.com/patiponrmutl/BESystem/models"
)

func main() {
	// โหลด config และเชื่อม DB ตามที่ main.go ใช้จริง
	cfg := config.Load()
	database.Connect(cfg) // ฟังก์ชันนี้ของคุณไม่รีเทิร์น error

	username := "Admin"
	password := "1234"

	// แฮชรหัสผ่าน
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	// ตรวจว่ามีผู้ใช้งานชื่อเดียวกันอยู่หรือไม่
	var existing models.User
	if err := database.DB.Where("username = ?", username).First(&existing).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalf("failed to query users: %v", err)
		}
	} else {
		fmt.Println("⚠️  Admin user already exists with username:", username)
		os.Exit(0)
	}

	// สร้าง user ใหม่ role=admin
	u := models.User{
		Username:     username,
		PasswordHash: string(hashed),
		Role:         "admin",
	}
	if err := database.DB.Create(&u).Error; err != nil {
		log.Fatalf("failed to insert admin: %v", err)
	}

	fmt.Println("✅ Admin user created successfully!")
	fmt.Println("   Username:", username)
	fmt.Println("   Password:", password, "(plain, remember to change later!)")
}
