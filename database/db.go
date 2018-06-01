package database

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/stripedpajamas/arkovmay/database/models"
)

var DB *gorm.DB

func InitDB() {
	db, err := gorm.Open("mysql", "root@/arkovmay?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("Failed to connect to database:", err.Error())
	}
	DB = db

	DB.AutoMigrate(
		&models.User{},
		&models.LoginToken{},
		&models.SessionToken{},
		&models.Mark{},
	)
}

func CloseDB() {
	DB.Close()
}
