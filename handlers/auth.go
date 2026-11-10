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
	//bcrypt 算法在处理密码时，只识别前 72 个字节。
	//如果用户输入超过 72 字节，部分版本的 bcrypt 会直接报错崩溃。
	//加上限制，既防崩溃，又防恶意超长字符串攻击。
	Password string `json:"password" binding:"required,min=6,max=72"`
}

func Register(c *gin.Context) {
	var input RegisterInput
	//前端发来 JSON，Gin 自动解析并执行上面的 binding 验证规则
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//加密
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	user := models.User{
		Username: input.Username,
		Password: hashedPassword,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已被占用"})
			return
		}
		log.Printf("注册失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败："})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功", "id": user.ID})
}

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user models.User
	//使用占位符 ? 是防 SQL 注入的标准写法
	//sql := "SELECT * FROM users WHERE username = '" + input.Username + "'"
	//通过拼接字符串来绕过密码登录
	if err := database.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	if !utils.CheckPassword(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	//载荷是什么？
	//在 JWT（JSON Web Token）中，Token 由三部分组成（头部.载荷.签名）。载荷就是中间的 JSON 数据部分。
	//把用户的 user.ID 塞进 JWT 的载荷（Payload），加上过期时间，用服务器的密钥（Secret）签名，生成一长串字符串
	//只放ID，因为 JWT 是 Base64 编码的，前端能解开看，放 ID 最安全
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token生成失败"})
		return
	}
	//把 token 返回给前端。前端拿到后，存进 LocalStorage，以后每次发文都带上
	//LocalStorage 是浏览器自带的一个小仓库。可以在里面存字符串（通常限制在 5MB 左右）。只要你不清除浏览器缓存，把电脑关了，数据还在。
	//因为 HTTP 请求是“短记忆”的，页面一刷新，内存里的变量就没了。存进 LocalStorage，不管你刷新页面还是关掉浏览器再开，Token 都在。前端拦截器每次发请求，都会去这里把 Token 拿出来，塞进请求头。
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})

}
