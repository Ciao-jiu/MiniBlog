package handlers

import (
	"MiniBlog/database"
	models "MiniBlog/models"
	"MiniBlog/utils"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	//加上限制，防崩溃，防恶意超长字符串攻击。
	Password string `json:"password" binding:"required,min=6,max=72"`
}

func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		//400，参数错误
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名或密码格式错误"})
		return
	}

	//加密
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		//500，服务器内部错误
		log.Printf("密码加密失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}
	user := models.User{
		Username: input.Username,
		Password: hashedPassword,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			//409，资源冲突
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已被占用"})
			return
		}
		log.Printf("创建用户失败：%v", err)
		//500
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}
	//201 Created
	c.JSON(http.StatusCreated, gin.H{"message": "注册成功", "id": user.ID})
}

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		//400
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名或密码格式错误"})
		return
	}
	var user models.User
	//使用占位符 ? 防 SQL 注入攻击
	err := database.DB.Where("username = ?", input.Username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		//401，未认证
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	if err != nil {
		log.Printf("登录查询用户失败：%v", err)
		//500
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}

	if !utils.CheckPassword(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})

}
