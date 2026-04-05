package gorm

import (
	"edu-evaluation-backed/internal/conf"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitGorm 根据 配置初始化 GORM 数据库连接，支持 postgres 和 mysql 驱动
//
// 参数:
//   - conf *conf.Data 数据配置，包含数据库驱动类型和连接源
//
// 返回值:
//   - *gorm.DB 初始化后的 GORM 数据库实例
func InitGorm(conf *conf.Data) *gorm.DB {
	var db *gorm.DB
	var err error
	switch conf.Database.Driver {
	case "postgres":
		db, err = gorm.Open(postgres.Open(conf.Database.Source))
		if err != nil {
			panic("数据库：postgres数据库连接失败" + err.Error())
		}
	case "mysql":
		db, err = gorm.Open(mysql.Open(conf.Database.Source))
		if err != nil {
			panic("数据库：mysql数据库连接失败" + err.Error())
		}
	}
	if db == nil {
		panic("数据库：数据库连接失败")
	}
	return db
}
