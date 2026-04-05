package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// Get 泛型 Cache-Aside 读取：先查 Redis，miss 则查 DB 并回填
func Get[T any](
	ctx context.Context,
	rdb *redis.Client,
	hc *HealthChecker,
	key string,
	ttl time.Duration,
	dbFunc func() (*T, error),
) (*T, error) {
	if hc.IsAvailable() {
		val, err := rdb.Get(ctx, key).Result()
		if err == nil {
			var data T
			if jsonErr := json.Unmarshal([]byte(val), &data); jsonErr == nil {
				return &data, nil
			}
			// 反序列化失败，删除脏缓存
			_ = rdb.Del(ctx, key)
		} else if err != redis.Nil {
			// Redis 非_miss 错误，记录失败
			hc.RecordFailure()
			log.Warnf("redis GET %s error: %v", key, err)
		}
	}

	// 查 DB
	data, err := dbFunc()
	if err != nil {
		return nil, err
	}

	// 异步写缓存
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if b, jsonErr := json.Marshal(data); jsonErr == nil {
			if setErr := rdb.Set(bgCtx, key, b, ttl).Err(); setErr != nil {
				log.Warnf("redis SET %s error: %v", key, setErr)
			}
		}
	}()

	return data, nil
}

// MGet 批量获取缓存，返回 map[key]*T，缺失的 key 不在 map 中
func MGet[T any](
	ctx context.Context,
	rdb *redis.Client,
	hc *HealthChecker,
	keys []string,
	dbFunc func(missingKeys []string) (map[string]*T, error),
	ttl time.Duration,
) (map[string]*T, error) {
	result := make(map[string]*T)

	if hc.IsAvailable() && len(keys) > 0 {
		vals, err := rdb.MGet(ctx, keys...).Result()
		if err != nil && err != redis.Nil {
			hc.RecordFailure()
			log.Warnf("redis MGET error: %v", err)
		} else {
			var missingKeys []string
			for i, val := range vals {
				if val == nil {
					missingKeys = append(missingKeys, keys[i])
					continue
				}
				var data T
				if jsonErr := json.Unmarshal([]byte(val.(string)), &data); jsonErr == nil {
					result[keys[i]] = &data
				} else {
					missingKeys = append(missingKeys, keys[i])
				}
			}

			if len(missingKeys) == 0 {
				return result, nil
			}

			// 查 DB 获取缺失数据
			dbData, dbErr := dbFunc(missingKeys)
			if dbErr != nil {
				return nil, dbErr
			}

			// 合并 DB 结果
			for k, v := range dbData {
				result[k] = v
			}

			// 异步回填缓存
			go func() {
				bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				pipe := rdb.Pipeline()
				for k, v := range dbData {
					if b, jsonErr := json.Marshal(v); jsonErr == nil {
						pipe.Set(bgCtx, k, b, ttl)
					}
				}
				if _, err := pipe.Exec(bgCtx); err != nil {
					log.Warnf("redis pipeline SET error: %v", err)
				}
			}()

			return result, nil
		}
	}

	// 熔断或 Redis 完全不可用，直接查 DB
	if len(keys) > 0 {
		dbData, dbErr := dbFunc(keys)
		if dbErr != nil {
			return nil, dbErr
		}
		for k, v := range dbData {
			result[k] = v
		}
	}

	return result, nil
}

// Delete 删除指定缓存 key
func Delete(ctx context.Context, rdb *redis.Client, hc *HealthChecker, keys ...string) {
	if !hc.IsAvailable() {
		return
	}
	if err := rdb.Del(ctx, keys...).Err(); err != nil {
		hc.RecordFailure()
		log.Warnf("redis DEL error: %v", err)
	}
}

// DeleteByPattern 按通配符删除缓存 key
func DeleteByPattern(ctx context.Context, rdb *redis.Client, hc *HealthChecker, pattern string) {
	if !hc.IsAvailable() {
		return
	}
	var cursor uint64
	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			hc.RecordFailure()
			log.Warnf("redis SCAN %s error: %v", pattern, err)
			return
		}
		if len(keys) > 0 {
			if err := rdb.Del(ctx, keys...).Err(); err != nil {
				log.Warnf("redis DEL by pattern error: %v", err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
}

// CheckAndSetRateLimit 登录限流：15分钟内最多 maxCount 次，返回当前计数和是否允许
func CheckAndSetRateLimit(ctx context.Context, rdb *redis.Client, hc *HealthChecker, key string, maxCount int, window time.Duration) (int, bool, error) {
	if !hc.IsAvailable() {
		// Redis 不可用时跳过限流
		return 0, true, nil
	}

	count, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		hc.RecordFailure()
		// 限流失败不阻塞登录
		return 0, true, nil
	}
	if count == 1 {
		rdb.Expire(ctx, key, window)
	}
	return int(count), count <= int64(maxCount), nil
}
