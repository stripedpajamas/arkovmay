package models

import (
	"github.com/jinzhu/gorm"
)

type Mark struct {
	gorm.Model
	Name     string // name of mark
	UserID   uint   // id of user who owns this mark
	PublicID string `gorm:"index"` // if their account is active
}
