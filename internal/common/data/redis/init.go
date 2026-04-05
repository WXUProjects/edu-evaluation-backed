package redis

import (
	"edu-evaluation-backed/internal/conf"
	"time"

	"github.com/redis/go-redis/v9"
)

// InitRedis 根据 配置初始化 Redis 客户端连接
//
// 参数:
//   - conf *conf.Data 数据配置，包含 Redis 地址、密码、超时等设置
//
// 返回值:
//   - *redis.Client 初始化后的 Redis 客户端实例
func InitRedis(conf *conf.Data) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         conf.Redis.Addr,
		Password:     conf.Redis.Password,
		DB:           1,
		ReadTimeout:  time.Duration(conf.Redis.ReadTimeout.Nanos),
		WriteTimeout: time.Duration(conf.Redis.WriteTimeout.Nanos),
		PoolSize:     50,
		MinIdleConns: 10,
		MaxRetries:   2,
	})
	return rdb
}
