package data

import (
	"context"
	"time"

	gorm2 "edu-evaluation-backed/internal/common/data/gorm"
	"edu-evaluation-backed/internal/common/data/cache"
	redis2 "edu-evaluation-backed/internal/common/data/redis"
	"edu-evaluation-backed/internal/conf"
	"edu-evaluation-backed/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ProviderSet 数据层 Wire 依赖注入提供者集合
var ProviderSet = wire.NewSet(
	NewData,
	NewDataDB,
	NewDataRDB,
	NewHealthChecker,
)

// Data 数据层上下文，持有数据库和 Redis 连接
type Data struct {
	DB  *gorm.DB
	RDB *redis.Client
	HC  *cache.HealthChecker
}

// NewDataDB 从 Data 中提取 GORM 数据库连接
//
// 参数:
//   - data *Data 数据层上下文
//
// 返回值:
//   - *gorm.DB
func NewDataDB(data *Data) *gorm.DB {
	return data.DB
}

// NewDataRDB 从 Data 中提取 Redis 客户端
//
// 参数:
//   - data *Data 数据层上下文
//
// 返回值:
//   - *redis.Client
func NewDataRDB(data *Data) *redis.Client {
	return data.RDB
}

// NewData 创建数据层上下文，初始化数据库和 Redis 连接，并执行自动迁移
//
// 参数:
//   - c *conf.Data 数据配置
//
// 返回值:
//   - *Data 数据层上下文
//   - func() 清理函数，用于关闭数据库连接
//   - error 初始化失败时返回错误
func NewData(c *conf.Data) (*Data, func(), error) {
	db := gorm2.InitGorm(c)
	rdb := redis2.InitRedis(c)

	// Ping Redis 做健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Warnf("redis ping failed: %v, service will run without cache", err)
	} else {
		log.Info("redis connected successfully")
	}

	data := &Data{DB: db, RDB: rdb, HC: cache.NewHealthChecker()}
	migrateModels(data.DB)
	cleanup := func() {
		log.Info("closing the data resources")
		sql, _ := data.DB.DB()
		sql.Close()
		_ = data.RDB.Close()
	}
	return data, cleanup, nil
}

// migrateModels 自动迁移所有数据模型到数据库
//
// 参数:
//   - db *gorm.DB 数据库连接
func migrateModels(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.Admin{},
		&model.Student{},
		&model.Teacher{},
		&model.Course{},
		&model.EvaluationTask{},
		&model.EvaluationDetail{},
	)
	if err != nil {
		panic("数据库：数据库自动合并失败" + err.Error())
	}
	// 插入默认管理员账号
	seedAdmin(db)
}

// seedAdmin 当管理员表为空时插入默认管理员账号（admin/admin）
//
// 参数:
//   - db *gorm.DB 数据库连接
func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&model.Admin{}).Count(&count)
	if count == 0 {
		admin := &model.Admin{Username: "admin", Password: "admin"}
		db.Create(admin)
	}
}

// NewHealthChecker 从 Data 中提取健康检查器
func NewHealthChecker(data *Data) *cache.HealthChecker {
	return data.HC
}
