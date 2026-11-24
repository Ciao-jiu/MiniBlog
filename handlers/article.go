package handlers

import (
	"MiniBlog/cache"
	"MiniBlog/database"
	"MiniBlog/models"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreatedArticleInput struct {
	Title   string `json:"title" binding:"required,max=100"`
	Content string `json:"content" binding:"required"`
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func GetArticles(c *gin.Context) {
	page := clamp(atoi(c.DefaultQuery("page", "1")), 1, 100000)
	pageSize := clamp(atoi(c.DefaultQuery("page_size", "10")), 1, 100)

	cacheKey := cache.ListKey(page, pageSize)
	ctx := c.Request.Context()

	cachedData, hit, err := cache.Get(ctx, cacheKey)
	if err != nil {
		log.Printf("读取文章列表缓存失败: %v", err)
	}
	if hit {
		// 缓存命中(Hit)！直接把 JSON 字符串返回给前端
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, cachedData)
		return
	}

	// 缓存未命中(Miss)，查 MySQL
	offset := (page - 1) * pageSize
	var total int64
	if err := database.DB.Model(&models.Article{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	var articles []models.Article
	err = database.DB.Preload("User").Order("created_at desc").Limit(pageSize).Offset(offset).Find(&articles).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	response := gin.H{
		"data":      articles,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}

	// 5. 将查询结果存入 Redis（设置 5 分钟过期）
	//设置5分钟的原因是1防止Redis内存被撑爆，2是作为一种兜底，数据库更新了但是缓存没删干净，5分钟后也会自动刷新
	// 把 response 变成 JSON 字符串
	jsonData, err := json.Marshal(response)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "响应序列化失败"})
		return
	}
	if err := cache.Set(ctx, cacheKey, jsonData, 5*time.Minute); err != nil {
		log.Printf("写入文章列表缓存失败: %v", err)
	}

	// 6. 返回数据
	c.JSON(http.StatusOK, response)
}

func CreatedArticle(c *gin.Context) {
	//验证前端发来的数据
	var articleInput CreatedArticleInput

	if err := c.ShouldBindJSON(&articleInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//从Gin上下文取中间件存入的 userID
	//c.Get()方法不知道是什么类型，所以使用c.Set()存的都是interface{}，取出来的也都是interface{}
	//interface{}是万能容器，可以装任何类型
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	// 类型断言确保安全转换
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	//从上下文注入，前端无法伪造
	article := models.Article{
		Title:   articleInput.Title,
		Content: articleInput.Content,
		UserID:  uid,
	}

	if err := database.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	//写操作成功后清理列表缓存
	cache.MustInvalidate(c.Request.Context())
	c.JSON(http.StatusCreated, article)
}
func GetArticleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID必须是正整数"})
		return
	}
	var article models.Article

	if err := database.DB.Preload("User").First(&article, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}
		log.Printf("查询文章出错:id = %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, article)
}

type UpdateArticleInput struct {
	Title   string `json:"title" binding:"required,max=100"`
	Content string `json:"content" binding:"required"`
}

func UpdateArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID必须是正整数"})
		return
	}

	var input UpdateArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var article models.Article

	if err := database.DB.First(&article, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	userID, exist := c.Get("userID")
	uid, ok := userID.(uint)

	if !exist || !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	if article.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权修改他人文章"})
		return
	}

	article.Title = input.Title
	article.Content = input.Content
	if err := database.DB.Save(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	cache.MustInvalidate(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": article})
}

func DeleteArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID必须是正整数"})
		return

	}

	var article models.Article
	if err := database.DB.First(&article, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//404 资源不存在
			c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}
		log.Printf("删除前查询文章失败，id=%d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	userID, exist := c.Get("userID")
	uid, ok := userID.(uint)
	if !exist || !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	if article.UserID != uid {
		//403，无权限
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除他人文章"})
		return
	}
	if err := database.DB.Delete(&article).Error; err != nil {
		log.Printf("删除文章失败，id=%d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	cache.MustInvalidate(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
