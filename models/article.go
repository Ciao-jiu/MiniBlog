package models

import "gorm.io/gorm"

type Article struct {
	//内嵌gorm.Model,自动获得ID，CreatedAt, UpdatedAt, DeletedAt 字段
	gorm.Model
	//gorm 标签用来精确控制列的类型和约束，json 标签控制 API 返回的字段名。
	Title   string `gorm:"type:varchar(100); not null" json:"title"`
	Content string `gorm:"type:text;not null" json:"content"`
	UserID  uint   `gorm:"not null;index" json:"user_id"`
	//用引用类型而不是值类型？
	//指针的零值是 nil，可以表示“不存在”；值类型的零值是一个空壳，会污染 JSON。所以关联字段用指针
	//omiempty意思是如果字段是零值，就完全不出现在 JSON 里
	User *User `gorm:"foreignkey:UserID" json:"user,omitempty"`
}
