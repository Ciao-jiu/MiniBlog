package handlers

import (
	"MiniBlog/database"
	"MiniBlog/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// func GetArticles(c *gin.Context) {
// 	// 从URL查询参数取page，默认第 1 页，每页 10 条
// 	page := clamp(atoi(c.DefaultQuery("page", "1")), 1, 100000)
// 	pageSize := clamp(atoi(c.DefaultQuery("page_size", "10")), 1, 100)
// 	offset := (page - 1) * pageSize
// 	var total int64
// 	//统计总记录数，写入 total，Count 返回的是 *gorm.DB
// 	//*gorm.DB 是 GORM 的数据库操作上下文。它内部记录着 SQL 构建状态、错误和连接池。
// 	// 链式调用的每一步都在修改这个上下文并返回它，直到调用 Find、Create 等终结方法才真正执行 SQL。
// 	// 错误挂在它的 .Error 字段上，所以必须用 .Error 来取。
// 	if err := database.DB.Model(&models.Article{}).Count(&total).Error; err != nil {
// 		log.Println("count 报错：", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
// 		return
// 	}

// 	var articles []models.Article
// 	// 2. 查询：Preload 预加载 User，Order 倒序，Limit 和 Offset 分页
// 	err := database.DB.
// 		Preload("User").
// 		Order("created_at desc").
// 		Limit(pageSize).
// 		Offset(offset).
// 		//执行查询，结果填入 articles 切片
// 		Find(&articles).Error

// 	if err != nil {
// 		log.Println("查询文章报错:", err)
// 		//c 是*gin.context,代表当前请求的上下文，c.JSON 是 Gin 提供的方法，作用是：
// 		//把一个 Go 对象序列化成 JSON，写进 HTTP 响应体，并设置状态码。
// 		//三件事：1.设置响应头 Content-Type:application/json
// 		//2.调用WriteHeader(code)写状态码
// 		//3.用json.Marshal把obj序列化写入响应体
// 		//签名大概是：
// 		//func (c *Context) JSON(code int, obj interface{})
// 		//http.StatusInternalServerError是Go 标准库net/http里定义的常量，值为500
// 		//http.StatusOK	200	成功
// 		// http.StatusCreated	201	已创建
// 		// http.StatusBadRequest	400	请求参数错误
// 		// http.StatusUnauthorized	401	未认证
// 		// http.StatusForbidden	403	无权限
// 		// http.StatusNotFound	404	资源不存在
// 		// http.StatusInternalServerError	500	服务器内部错误
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"data":      articles,
// 		"page":      page,
// 		"page_size": pageSize,
// 		"total":     total,
// 	})

// }
func GetArticles(c *gin.Context) {
	fmt.Println("🔥 请求成功进入 GetArticles 接口！")
	page := clamp(atoi(c.DefaultQuery("page", "1")), 1, 100000)
	pageSize := clamp(atoi(c.DefaultQuery("page_size", "10")), 1, 100)

	// 1. 生成 Redis 的 Key（区分不同页码，比如 articles:page:1:size:10）
	cacheKey := fmt.Sprintf("articles:page:%d:size:%d", page, pageSize)
	ctx := context.Background()

	// 2. 尝试从 Redis 获取数据
	cachedData, err := database.RDB.Get(ctx, cacheKey).Result()
	if err == nil {
		// 缓存命中！直接把 JSON 字符串返回给前端
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, cachedData)
		return // 直接结束，不查数据库了
	}

	// 3. 缓存未命中，老老实实查 MySQL
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

	// 4. 组装返回的数据
	response := gin.H{
		"data":      articles,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}

	// 5. 将查询结果存入 Redis（设置 5 分钟过期）
	// 把 response 变成 JSON 字符串
	jsonData, _ := json.Marshal(response)
	database.RDB.Set(ctx, cacheKey, jsonData, 5*time.Minute)

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
	//类型断言，把 interface{} 转成 uint
	//uint是无符号整数，Unsigned Integer, 不能存负数，只能存正数。
	//GORM默认的主键ID类型就是uint
	//断言userID的类型是uint，如果是就取出，不是，ok就是false，uid为0
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
	keys, _ := database.RDB.Keys(context.Background(), "articles:*").Result()
	if len(keys) > 0 {
		database.RDB.Del(context.Background(), keys...)
	}
	c.JSON(http.StatusCreated, article)
}
func GetArticleByID(c *gin.Context) {
	//获取路径参数中的id
	idStr := c.Param("id")
	//按十进制解析成 uint64
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID必须是正整数"})
		return
	}
	var article models.Article
	//Preload是预加载，即关联查询，除了查询article表，再使用userid查询user表

	//First是GORM 用来发sql指令的方法，查询条件是主键等于id，结果放到article里
	if err := database.DB.Preload("User").First(&article, id).Error; err != nil {
		// 判断是否是“记录不存在”错误
		//GORM 查询不到数据时，不会返回 nil，而是返回一个特定的错误：gorm.ErrRecordNotFound
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}
		log.Printf("查询文章出错:id = %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}

	c.JSON(200, article)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "err.Error()"})
		return
	}
	var article models.Article

	if err := database.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文章不存在"})
		return
	}

	//拿到当前登录用户的 ID（中间件塞进去的）
	userID, _ := c.Get("userID")
	uid, _ := userID.(uint)

	if article.UserID != uid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无权修改他人文章"})
		return
	}

	article.Title = input.Title
	article.Content = input.Content
	if err := database.DB.Save(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	keys, _ := database.RDB.Keys(context.Background(), "articles:*").Result()
	if len(keys) > 0 {
		database.RDB.Del(context.Background(), keys...)
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": article})
}

func DeletArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID必须是正整数"})
		return

	}

	var article models.Article
	if err := database.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	userID, _ := c.Get("userID")
	uid, _ := userID.(uint)
	if article.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除他人文章"})
		return
	}
	if err := database.DB.Delete(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
