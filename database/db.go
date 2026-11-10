package database

import (
	models "MiniBlog/models"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/miniblog?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		panic("连接数据库失败：" + err.Error())
	}
	fmt.Println("连接数据库成功")
	err = DB.AutoMigrate(&models.User{}, &models.Article{})
	if err != nil {
		panic("数据库迁移失败：" + err.Error())
	}
	fmt.Println("数据库表结构迁移成功！")
}
