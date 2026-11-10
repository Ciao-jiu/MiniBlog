package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // Redis 地址
		Password: "",               // 没有密码就留空
		DB:       0,                // 默认库
	})

	// 测试连接
	ctx := context.Background()
	pong, err := RDB.Ping(ctx).Result()
	if err != nil {
		panic("Redis 连接失败: " + err.Error())
	}
	fmt.Println("Redis 连接成功！Ping结果:", pong)
}
