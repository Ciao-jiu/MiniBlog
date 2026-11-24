package database

import (
	"MiniBlog/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() {

	RDB = redis.NewClient(&redis.Options{
		Addr:     config.Env("MINIBLOG_REDIS_ADDR", "127.0.0.1:6379"),
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	pong, err := RDB.Ping(ctx).Result()
	if err != nil {
		//Fail-fast 快速失败 原则
		panic("Redis 连接失败: " + err.Error())
	}
	fmt.Println("Redis 连接成功！Ping结果:", pong)
}
