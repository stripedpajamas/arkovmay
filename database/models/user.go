package models

import (
	"github.com/jinzhu/gorm"

	"time"
)

type User struct {
	gorm.Model
	Email string `gorm:"unique_index"` // username is email address
}

type LoginToken struct {
	gorm.Model
	Email   string
	Token   string `gorm:"index"`
	Expires time.Time
}

type SessionToken struct {
	gorm.Model
	UserID uint
	Token  string `gorm:"index"`
}
