# 项目文档 - 无锡学院评教系统

## 项目概述

基于 Go-Kratos 框架的学生评教系统后端，支持 gRPC + HTTP 双协议，数据库使用 PostgreSQL/MySQL，缓存使用 Redis。

## 架构分层

```
cmd/              - 入口点，wire DI 配置
internal/
├── biz/          - 业务逻辑层（UseCase）
├── service/      - 传输层（HTTP Handler，实现 proto 定义的接口）
├── data/         - 数据访问层
│   ├── dal/      - DAL（Repository）实现
│   └── model/    - GORM 模型定义
├── server/       - HTTP Server 配置和路由注册
├── conf/         - 配置（Protobuf 定义）
common/
├── data/         - 通用数据工具（Redis、GORM 初始化）
├── utils/        - 工具函数（分页、Gob编解码）
└── const/        - 常量（JWT Secret 等）
```

## 数据模型 (Models)

### 1. Admin 管理员表
```go
type Admin struct {
    gorm.Model       // ID, CreatedAt, UpdatedAt, DeletedAt
    Username string  // 用户名，唯一索引
    Password string  // 密码
}
```
**中间表**：无（独立表）

### 2. Teacher 教师表
```go
type Teacher struct {
    gorm.Model       // ID, CreatedAt, UpdatedAt, DeletedAt
    Name    string   // 教师姓名
    Sex     string   // 性别
    WorkNo  string   // 工号，唯一索引
    Email   string   // 邮箱
    Courses []Course // 多对多：教师讲授的课程
}
```
**中间表**：`course_teachers`（自动生成）

### 3. Student 学生表
```go
type Student struct {
    gorm.Model       // ID, CreatedAt, UpdatedAt, DeletedAt
    Name      string // 学生姓名
    StudentNo string // 学号，唯一索引
    Sex       string // 性别
    IdCardNo  string // 身份证号
    Courses   []Course // 多对多：学生选修的课程
}
```
**中间表**：`course_students`（通过 GORM many2many 自动生成）

### 4. Course 教学班表
```go
type Course struct {
    gorm.Model       // ID, CreatedAt, UpdatedAt, DeletedAt
    Status     int    // 课程状态：1-进行中 2-已结课
    CourseName string // 课程名称
    ClassName  string // 班级名称，唯一索引
    Teachers   []Teacher // 多对多：授课教师
    Students   []Student // 多对多：班级学生
    // 实时评分字段
    EvaluationScore int // 评教总分
    EvaluationNum   int // 评教人数
}
```
**中间表**：`course_teachers`、`course_students`、`evaluation_courses`

### 5. EvaluationTask 评教任务表
```go
type EvaluationTask struct {
    gorm.Model       // ID, CreatedAt, UpdatedAt, DeletedAt
    Status  int      // 评教状态：0-未开始 1-进行中 2-已结束
    Title   string   // 评教标题
    Courses []Course // 多对多：参与评教的课程
    Details []EvaluationDetail // 一对多：评教详情
}
```
**中间表**：`evaluation_courses`

### 6. EvaluationDetail 评价详情表
```go
type EvaluationDetail struct {
    gorm.Model       // ID, CreatedAt, UpdatedAt, DeletedAt
    TaskId    uint   // 评教任务ID
    CourseId  uint   // 课程ID
    StudentId uint   // 学生ID
    TeacherId uint   // 教师ID
    Course    Course  // 关联：用于 Preload
    Student   Student // 关联：用于 Preload
    Detail    string  // 学生评价的 JSON 信息
    Summary   string  // 学生总结信息
    Score     int     // 本次评价折算后的总分
}
```

## 表关系图

```
┌─────────────┐      ┌──────────────────┐      ┌─────────────┐
│   Admin     │      │ EvaluationTask   │      │   Teacher    │
└─────────────┘      └────────┬─────────┘      └──────┬──────┘
                               │                         │
                               │ many2many              │ many2many
                               ▼                         ▼
                    ┌───────────────────┐      ┌─────────────────┐
                    │ evaluation_courses │      │ course_teachers │
                    └───────────────────┘      └─────────────────┘
                               │                         │
                               │               ┌─────────┴─────────┐
                               │               │                   │
                               ▼               ▼                   ▼
                        ┌─────────────┐ ┌───────────┐     ┌───────────┐
                        │   Course    │ │  Student  │     │  Teacher  │
                        └──────┬──────┘ └─────┬─────┘     └───────────┘
                               │ many2many  │
                               ▼             │
                    ┌───────────────────┐    │
                    │ course_students   │◄───┘
                    └───────────────────┘
```

## API 路由

### 认证模块 (api/v1/auth/auth.proto)
| 方法 | 路由 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/admin/login` | 管理员登录 |
| POST | `/api/v1/auth/student/login` | 学生登录（校验学号+身份证+taskId） |
| GET | `/api/v1/auth/student/info?stuNo=xxx` | 获取学生个人信息 |

### 学生管理 (api/v1/base-info/student)
| 方法 | 路由 | 说明 |
|------|------|------|
| GET | `/api/v1/base-info/student/list` | 获取学生列表（分页、模糊搜索） |
| POST | `/api/v1/base-info/student/update` | 更新学生信息 |
| POST | `/api/v1/base-info/student/delete` | 删除学生 |
| POST | `/api/v1/base-info/student/import` | 导入学生（Excel） |

### 教师管理 (api/v1/base-info/teacher)
| 方法 | 路由 | 说明 |
|------|------|------|
| GET | `/api/v1/base-info/teacher/list` | 获取教师列表 |
| POST | `/api/v1/base-info/teacher/update` | 更新教师信息 |
| POST | `/api/v1/base-info/teacher/delete` | 删除教师 |
| POST | `/api/v1/base-info/teacher/import` | 导入教师（Excel） |

### 课程管理 (api/v1/base-info/course)
| 方法 | 路由 | 说明 |
|------|------|------|
| GET | `/api/v1/base-info/course/list` | 获取课程列表 |
| GET | `/api/v1/base-info/course/detail` | 获取课程详情（包含教师和学生） |
| POST | `/api/v1/base-info/course/edit` | 编辑课程（名称、班级、教师） |
| POST | `/api/v1/base-info/course/delete` | 删除课程 |
| POST | `/api/v1/base-info/course/import` | 导入课程（Excel） |

### 评教任务 (api/v1/eva_task/task.proto)
| 方法 | 路由 | 说明 |
|------|------|------|
| POST | `/api/v1/task/create` | 创建评教任务 |
| GET | `/api/v1/task/list` | 获取任务列表 |
| GET | `/api/v1/task/detail` | 获取任务详情 |
| GET | `/api/v1/task/student_task_detail` | 学生查看任务详情（只看自己） |
| POST | `/api/v1/task/change_status` | 修改任务状态 |
| POST | `/api/v1/task/submit_evaluation` | 提交评价 |
| GET | `/api/v1/task/export?taskId=xxx` | 导出评教结果（xlsx + PDF） |

## DAL 层结构

### BaseInfoDal (internal/data/dal/base_info.go)
- 学生/教师的 CRUD
- `InsertStudent/InsertTeacher`: 批量插入，UPSERT 策略（学号/工号冲突跳过）
- `QueryStudent/QueryTeacher`: 分页 + 模糊搜索
- `GetStudentByID/GetTeacherByID`: ID 查询
- `UpdateStudent/UpdateTeacher`: 更新（含唯一性校验）
- `DeleteStudent/DeleteTeacher`: 删除（清除课程关联后删除实体）
- `AdminLogin`: 管理员登录验证
- `StudentLogin`: 学生登录验证（学号+身份证+task范围内验证）
- `GetStudentByStudentNo`: 按学号查询学生

### CourseDal (internal/data/dal/course.go)
- `Detail`: 获取课程详情（预加载教师和学生）
- `CreateCourse`: 创建课程
- `AddStudent`: 添加学生到课程
- `AddTeachers`: 绑定教师到课程（先清后加）
- `List`: 课程列表（分页）
- `UpdateCourse`: 更新课程信息
- `DeleteCourse`: 删除课程（清除所有关联）

### TaskDal (internal/data/dal/eva_task.go)
- `CreateTask`: 创建评教任务
- `GetTaskList`: 任务列表（分页 + 状态筛选）
- `GetTaskDetail`: 任务详情（预加载课程及关联）
- `ChangeTaskStatus`: 修改任务状态
- `StudentTaskDetail`: 学生视角的任务详情（只返回关联的课程和教师）
- `SubmitEvaluation`: 提交评价
- `GetTaskEvaluationDetail`: 获取单条评价详情
- `GetTaskEvaluationResults`: 获取任务评教结果（用于xlsx导出）
- `GetTeacherEvaluationDetailsForPDF`: 获取教师评教详情（用于PDF导出，包含排名）

## Excel 导入格式

### 学生 Excel (Sheet1)
| 列索引 | 字段 | 说明 |
|--------|------|------|
| 0 | - | 跳过（表头行） |
| 1 | StudentNo | 学号 |
| 2 | Name | 姓名 |
| 3 | Sex | 性别 |
| 4 | IdCardNo | 身份证号 |

### 教师 Excel (Sheet1)
| 列索引 | 字段 | 说明 |
|--------|------|------|
| 0 | - | 跳过（表头行） |
| 1 | WorkNo | 工号 |
| 2 | Name | 姓名 |
| 3 | Sex | 性别 |
| 4 | Email | 邮箱 |

### 课程 Excel (Sheet1)
| 列索引 | 字段 | 说明 |
|--------|------|------|
| 0 | - | 跳过（表头行） |
| 1 | CourseName | 课程名称 |
| 2 | ClassName | 班级名称 |
| 3 | StudentNo | 学生学号（同一课程+班级会合并） |

## 业务逻辑要点

### 学生登录验证逻辑
```sql
-- 验证学生是否在指定 Task 的评教范围内
SELECT COUNT(*) FROM courses c
INNER JOIN evaluation_courses ec ON c.id = ec.course_id
INNER JOIN course_students cs ON c.id = cs.course_id
WHERE ec.evaluation_task_id = ? AND cs.student_student_no = ?
-- 如果 count > 0，说明学生在评教范围内
```

### 提交评价逻辑
1. 检查是否已评价过（task + course + student + teacher 联合唯一）
2. 创建评价详情记录
3. 更新课程的评教人数和总分

### 课程导入逻辑
- 按 `课程名称 + 班级名称` 分组
- 相同组合的学生归入同一个课程
- 班级名称全局唯一，冲突则跳过

## 依赖注入 (Wire)

```
Data (DB + Redis)
    ├── BaseInfoDal
    │     └── AuthUseCase → AuthService
    │     └── StudentUseCase → StudentService
    │     └── TeacherUseCase → TeacherService
    ├── CourseDal
    │     └── CourseUseCase → CourseService
    └── TaskDal
          └── EvaTaskUseCase → EvaTaskService
                                    └── (依赖 BaseInfoDal, CourseDal)
```

## 工具函数

### 分页 (pagination.go)
- `PageNumHandle`: 处理分页参数（page<=0 → 1, size<=0 → 10, size>=100 → 100）
- `CalculateOffset`: 计算 OFFSET 值

### Gob 编解码 (gob.go)
- `GobEncoder`: 编码对象为字节
- `GobDecoder`: 解码字节为对象
- 用于 Redis 缓存序列化

## 配置

配置通过 `internal/conf/conf.proto` 定义，加载自 `configs/config.yaml`：
- Server: HTTP/gRPC 端口、超时
- Data: 数据库驱动、DSN、Redis 配置

## 默认数据

启动时自动插入默认管理员账号：`admin / admin`
