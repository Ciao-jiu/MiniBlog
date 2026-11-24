package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"MiniBlog/database"

	"github.com/redis/go-redis/v9"
)

// 统一规范 Redis Key 的前缀
const listKeyPrefix = "articles:list:"

// ListKey 让缓存 key 带上分页维度，不同页码互不干扰。
func ListKey(page, pageSize int) string {
	return fmt.Sprintf("%spage:%d:size:%d", listKeyPrefix, page, pageSize)
}

// Get 三返回值区分「未命中」与「Redis 故障」：
func Get(ctx context.Context, key string) (value string, hit bool, err error) {
	v, err := database.RDB.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		//缓存未命中
		return "", false, nil
	}
	if err != nil {
		//网络超时、Redis 宕机
		return "", false, err
	}
	//缓存命中
	return v, true, nil
}

// 必须强制要求传入 ttl，存缓存不设过期时间，会导致 Redis 内存被撑爆
func Set(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	return database.RDB.Set(ctx, key, data, ttl).Err()
}

// InvalidateList 用 SCAN 和 UNLINK 替代 KEYS 和 DEL，避免阻塞 Redis 主线程。
func InvalidateList(ctx context.Context) error {
	var cursor uint64
	var keys []string
	for {
		batch, next, err := database.RDB.Scan(ctx, cursor, listKeyPrefix+"*", 100).Result()
		if err != nil {
			return err
		}
		keys = append(keys, batch...)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	if len(keys) == 0 {
		return nil
	}
	// UNLINK 异步回收，比 DEL 更不容易造成主线程抖动
	return database.RDB.Unlink(ctx, keys...).Err()
}

// MustInvalidate 写路径统一入口：缓存失效失败只记日志，不回滚已成功的 DB 操作。
func MustInvalidate(ctx context.Context) {
	if err := InvalidateList(ctx); err != nil {
		log.Printf("[cache] invalidate article list failed: %v", err)
	}
}
