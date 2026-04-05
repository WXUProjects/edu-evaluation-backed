package dal

import (
	"edu-evaluation-backed/internal/common/data/cache"
	"edu-evaluation-backed/internal/common/utils"
	"edu-evaluation-backed/internal/data"
	"edu-evaluation-backed/internal/data/model"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BaseDal 通用数据访问基础层
// 提供可复用的 CRUD 操作，减少重复代码
type BaseDal struct {
	db  *gorm.DB
	rdb *redis.Client
	hc  *cache.HealthChecker
}

// NewBaseDal 创建通用数据访问基础层
func NewBaseDal(data *data.Data) *BaseDal {
	return &BaseDal{
		db:  data.DB,
		rdb: data.RDB,
		hc:  data.HC,
	}
}

// InsertBatch 批量插入，支持 UPSERT 策略
// items: 要插入的数据
// uniqueColumn: 唯一约束列名，用于冲突检测
// 返回值: 插入失败返回错误
func (b *BaseDal) InsertBatch(items interface{}, uniqueColumn string) error {
	return b.db.Clauses(
		clause.OnConflict{
			DoNothing: true,
			Columns:   []clause.Column{{Name: uniqueColumn}},
		}).Create(items).Error
}

// QueryWithPage 分页查询通用方法
// modelType: 模型类型
// page: 页码
// size: 每页条数
// conditions: 查询条件函数
// 返回值: 结果列表，总数，错误
func (b *BaseDal) QueryWithPage(modelType interface{}, page, size int, conditions func(*gorm.DB) *gorm.DB) ([]interface{}, int64, error) {
	page, size = utils.PageNumHandle(page, size)
	var total int64
	query := b.db.Model(modelType)
	if conditions != nil {
		query = conditions(query)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := utils.CalculateOffset(page, size)
	var results []interface{}
	if err := query.Order("id desc").Limit(size).Offset(offset).Find(&results).Error; err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

// UpdateFields 通用字段更新
// modelType: 模型类型
// id: 记录ID
// fields: 要更新的字段映射
// uniqueColumn/uniqueValue: 唯一性检查（可选）
// 返回值: 更新后的记录，错误
func (b *BaseDal) UpdateFields(modelType interface{}, id uint, fields map[string]interface{}, uniqueColumn, uniqueValue string) (interface{}, error) {
	// 先查询记录是否存在
	record := b.db.Model(modelType).Where("id = ?", id).First(modelType)
	if record.Error != nil {
		return nil, record.Error
	}

	// 唯一性检查
	if uniqueColumn != "" && uniqueValue != "" {
		var count int64
		err := b.db.Model(modelType).Where(uniqueColumn+" = ? AND id != ?", uniqueValue, id).Count(&count).Error
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New(uniqueColumn + "已存在")
		}
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	for k, v := range fields {
		if v != nil && v != "" {
			updates[k] = v
		}
	}

	if len(updates) == 0 {
		return modelType, nil
	}

	if err := b.db.Model(modelType).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 返回更新后的数据
	b.db.First(modelType, id)
	return modelType, nil
}

// DeleteWithAssociations 删除记录前清除关联
// record: 要删除的记录（需要先加载）
// associations: 要清除的关联名称列表
// 返回值: 删除失败返回错误
func (b *BaseDal) DeleteWithAssociations(record interface{}, associations []string) error {
	for _, assoc := range associations {
		if err := b.db.Model(record).Association(assoc).Clear(); err != nil {
			return err
		}
	}
	return b.db.Delete(record).Error
}

// GetByID 根据ID获取记录
func (b *BaseDal) GetByID(modelType interface{}, id uint) (interface{}, error) {
	err := b.db.First(modelType, id).Error
	return modelType, err
}

// SearchFuzzy 模糊搜索
// modelType: 模型类型
// page, size: 分页参数
// searchFields: 搜索字段列表
// keyword: 搜索关键词
// 返回值: 结果列表，总数，错误
func (b *BaseDal) SearchFuzzy(modelType interface{}, page, size int, searchFields []string, keyword string) ([]interface{}, int64, error) {
	page, size = utils.PageNumHandle(page, size)
	var total int64
	query := b.db.Model(modelType)

	if keyword != "" {
		for i, field := range searchFields {
			if i == 0 {
				if field == "student_no" || field == "work_no" {
					query = query.Where(field+" LIKE ?", keyword+"%")
				} else {
					query = query.Where(field+" LIKE ?", "%"+keyword+"%")
				}
			} else {
				if field == "student_no" || field == "work_no" {
					query = query.Or(field+" LIKE ?", keyword+"%")
				} else {
					query = query.Or(field+" LIKE ?", "%"+keyword+"%")
				}
			}
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var results []interface{}
	offset := utils.CalculateOffset(page, size)
	if err := query.Order("id desc").Limit(size).Offset(offset).Find(&results).Error; err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

// Student 快捷方法
func (b *BaseDal) DB() *gorm.DB {
	return b.db
}

// InsertStudent 保留原有方法以保持兼容性
func (b *BaseDal) InsertStudent(students []*model.Student) error {
	return b.InsertBatch(students, "student_no")
}

// InsertTeacher 保留原有方法以保持兼容性
func (b *BaseDal) InsertTeacher(teachers []*model.Teacher) error {
	return b.InsertBatch(teachers, "work_no")
}

// ResetAll 清空所有数据，只保留学生表和教师表（硬删除）
// 清空: evaluation_details, evaluation_courses, evaluation_tasks, course_teachers, course_students, courses
func (b *BaseDal) ResetAll() error {
	return b.db.Transaction(func(tx *gorm.DB) error {
		// 先删子表，再删关联表，最后删父表
		// 1. 删评教详情（最底层子表）
		if err := tx.Exec("DELETE FROM evaluation_details").Error; err != nil {
			return err
		}
		// 2. 删 evaluation_courses（在评教任务之前，因为有外键引用）
		if err := tx.Exec("DELETE FROM evaluation_courses").Error; err != nil {
			return err
		}
		// 3. 删评教任务
		if err := tx.Exec("DELETE FROM evaluation_tasks").Error; err != nil {
			return err
		}
		// 4. 删课程和教师的关联表
		if err := tx.Exec("DELETE FROM course_teachers").Error; err != nil {
			return err
		}
		// 5. 删课程和学生的关联表
		if err := tx.Exec("DELETE FROM course_students").Error; err != nil {
			return err
		}
		// 6. 删课程表
		if err := tx.Exec("DELETE FROM courses").Error; err != nil {
			return err
		}
		return nil
	})
}
