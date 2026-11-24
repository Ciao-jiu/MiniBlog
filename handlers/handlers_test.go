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
