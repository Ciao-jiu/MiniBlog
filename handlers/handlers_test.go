package handlers_test

import (
	"MiniBlog/database"
	"MiniBlog/models"
	"MiniBlog/router"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 初始化测试环境
// 中间件/路由测试
func setupRouter() *gin.Engine {

	return router.SetupRouter()
}

// Handler/数据库集成测试
func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	dsn := os.Getenv("MINIBLOG_TEST_DB_DSN")
	if dsn == "" {
		t.Fatal("MINIBLOG_TEST_DB_DSN 环境变量未设置")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("连接测试数据库失败：%v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Article{}); err != nil {
		t.Fatalf("迁移测试表失败:%v", err)
	}
	database.DB = db

	redisAddr := os.Getenv("MINIBLOG_TEST_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	//RDB Redis DataBase
	database.RDB = redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := database.RDB.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("清理测试Redis失败: %v", err)
	}
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Article{})
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.User{})
	gin.SetMode(gin.TestMode)
	return router.SetupRouter()

}

//发HTTP请求

func performJSONRequest(t *testing.T, r http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	return recorder
}

//登录拿token

func loginAs(t *testing.T, r http.Handler, username, password string) string {
	t.Helper()
	body := fmt.Sprintf(`{"username":%q, "password":%q}`, username, password)
	recorder := performJSONRequest(t, r, http.MethodPost, "/api/v1/login", body, "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("登录失败: %s", recorder.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	json.Unmarshal(recorder.Body.Bytes(), &resp)
	return resp.Token
}

func TestCreateArticleRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupRouter()

	recorder := performJSONRequest(t, r, http.MethodPost, "/api/v1/articles", `{"title":"test","content":"body"}`, "")

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("期望状态码 %d, 实际得到 %d, 响应：%s", http.StatusUnauthorized, recorder.Code, recorder.Body.String())
	}
}

func TestUpdateArticleNotFound(t *testing.T) {
	r := setupTestRouter(t)

	registerResponse := performJSONRequest(t, r, http.MethodPost, "/api/v1/register", `{"username":"testuser", "password":"password123"}`, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("注册用户失败：状态码 %d, 响应：%s", registerResponse.Code, registerResponse.Body.String())
	}
	token := loginAs(t, r, "testuser", "password123")
	recorder := performJSONRequest(t, r, http.MethodPut, "/api/v1/articles/999", `{"title":"updated", "content":"updated content"}`, token)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("期望状态码 %d, 实际得到 %d, 响应：%s", http.StatusNotFound, recorder.Code, recorder.Body.String())
	}
}

func TestUpdateArticleForbiddenForOtherUser(t *testing.T) {
	r := setupTestRouter(t)
	register := func(username string) {
		recorder := performJSONRequest(t, r, http.MethodPost, "/api/v1/register", fmt.Sprintf(`{"username":%q,"password":"password123"}`, username), "")
		if recorder.Code != http.StatusCreated {
			t.Fatalf("注册用户 %s 失败：状态码 %d, 响应：%s", username, recorder.Code, recorder.Body.String())
		}
	}

	register("author")
	register("otheruser")

	authorToken := loginAs(t, r, "author", "password123")
	otherToken := loginAs(t, r, "otheruser", "password123")

	createResponse := performJSONRequest(t, r, http.MethodPost, "/api/v1/articles", `{"title":"original title","content":"original content"}`, authorToken)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("创建文章失败：状态码 %d, 响应 %s", createResponse.Code, createResponse.Body.String())
	}

	var createdArticle struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(createResponse.Body.Bytes(), &createdArticle); err != nil {
		t.Fatalf("解析创建文章响应失败: %v", err)
	}
	if createdArticle.ID == 0 {
		t.Fatalf("解析创建文章响应中的ID为空")
	}

	updateResponse := performJSONRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/articles/%d", createdArticle.ID), `{"title":"hacked title","content":"hacked content"}`, otherToken)
	if updateResponse.Code != http.StatusForbidden {
		t.Fatalf("期望状态码 %d,实际得到 %d, 响应：%s", http.StatusForbidden, updateResponse.Code, updateResponse.Body.String())
	}

	getResponse := performJSONRequest(t, r, http.MethodGet, fmt.Sprintf("/api/v1/articles/%d", createdArticle.ID), "", "")
	if getResponse.Code != http.StatusOK {
		t.Fatalf("查询文章失败：状态码 %d,响应：%s", getResponse.Code, getResponse.Body.String())
	}

	var article struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.Unmarshal(getResponse.Body.Bytes(), &article); err != nil {
		t.Fatalf("解析文章响应失败：%v", err)
	}
	if article.Title != "original title" || article.Content != "original content" {
		t.Fatalf("文章被越权修改：title=%q,content=%q", article.Title, article.Content)
	}
}
