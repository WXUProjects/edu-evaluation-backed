package model

import (
	"gorm.io/gorm"
)

// Admin 管理员表
type Admin struct {
	gorm.Model
	Username string `json:"username" gorm:"uniqueIndex;size:32;comment:用户名"`
	Password string `json:"password" gorm:"size:128;comment:密码"`
}

// Teacher 教师表
type Teacher struct {
	gorm.Model
	Name    string   `json:"name" gorm:"size:64;comment:教师姓名"`
	Sex     string   `json:"sex" gorm:"size:1;comment:性别"`
	WorkNo  string   `json:"work_no" gorm:"uniqueIndex;size:32;comment:教师工号"`
	Email   string   `json:"email" gorm:"size:64;comment:教师邮箱"`
	Courses []Course `gorm:"many2many:course_teachers;"`
}

// EvaluationTask 评价
type EvaluationTask struct {
	gorm.Model
	Status  int                `json:"status" gorm:"comment:评教状态 0-未开始 1-进行中 2-已结束"`
	Title   string             `json:"title" gorm:"size:128;comment:评教标题"`
	Courses []Course           `json:"courses" gorm:"many2many:evaluation_courses;comment:参与评教的课程"`
	Details []EvaluationDetail `json:"details" gorm:"comment:评教详情"`
}

// Student 学生表
type Student struct {
	gorm.Model
	Name      string `json:"name" gorm:"size:64;comment:学生姓名"`
	StudentNo string `json:"student_no" gorm:"uniqueIndex;size:32;comment:学生学号"`
	Sex       string `json:"sex" gorm:"size:1;comment:性别"`
	IdCardNo  string `json:"id_card_no" gorm:"size:32;comment:学生身份证号"`
	// 反向引用（可选）：方便通过学生查选了哪些课
	Courses []Course `json:"courses" gorm:"many2many:course_students;"`
}

// Course 教学班
type Course struct {
	gorm.Model
	Status     int    `json:"status" gorm:"comment:课程状态 1-进行中 2-已结课"`
	CourseName string `json:"name" gorm:"size:64;comment:课程名称"`
	ClassName  string `json:"class_name" gorm:"unique;size:128;comment:班级名称"`
	// 老师和课程：多对多
	Teachers []Teacher `json:"teachers" gorm:"many2many:course_teachers;comment:授课教师"`

	// 【关键修改点】：学生和课程：多对多。必须加 many2many 标签，指定中间表名为 course_students
	Students []Student `json:"students" gorm:"many2many:course_students;references:student_no;comment:班级学生"`

	// 实时评分字段
	EvaluationScore int `json:"evaluation_score" gorm:"default:0;comment:评教总分"`
	EvaluationNum   int `json:"evaluation_num" gorm:"default:0;comment:评教人数"`
}

// EvaluationDetail 评价详情
type EvaluationDetail struct {
	gorm.Model
	TaskId    uint `json:"task_id" gorm:"index;comment:评价任务id"`
	CourseId  uint `json:"course_id" gorm:"index;comment:课程id"`
	StudentId uint `json:"student_id" gorm:"index;comment:学生id"`
	// 这里的实体引用主要用于 Preload 查询，不影响表结构生成
	Course  Course  `json:"-" gorm:"foreignKey:CourseId"`
	Student Student `json:"student" gorm:"foreignKey:StudentId"`
	Detail  string  `json:"detail" gorm:"type:text;comment:学生评价的json信息"`
	Score   int     `json:"score" gorm:"comment:本次评价折算后的总分"`
}
