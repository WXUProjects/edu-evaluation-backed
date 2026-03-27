package data

import (
	gorm2 "edu-evaluation-backed/internal/common/data/gorm"
	redis2 "edu-evaluation-backed/internal/common/data/redis"
	"edu-evaluation-backed/internal/conf"
	"edu-evaluation-backed/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDataDB,
	NewDataRDB,
)

// Data .
type Data struct {
	DB  *gorm.DB
	RDB *redis.Client
}

// NewDataDB 从 Data 中提取 DB
func NewDataDB(data *Data) *gorm.DB {
	return data.DB
}

// NewDataRDB 从 Data 中提取 RDB
func NewDataRDB(data *Data) *redis.Client {
	return data.RDB
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	data := &Data{DB: gorm2.InitGorm(c), RDB: redis2.InitRedis(c)}
	migrateModels(data.DB)
	cleanup := func() {
		log.Info("closing the data resources")
		sql, _ := data.DB.DB()
		sql.Close()
	}
	return data, cleanup, nil
}

// migrateModels 合并
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

// seedAdmin 插入默认管理员账号
func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&model.Admin{}).Count(&count)
	if count == 0 {
		admin := &model.Admin{Username: "admin", Password: "admin"}
		db.Create(admin)
	}
}
