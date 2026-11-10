package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
	//一对多关联
	Article []Article `gorm:"foreignkey:UserID" json:"articles,omitempty"`
}
