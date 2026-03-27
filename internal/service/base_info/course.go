package base_info

import (
	context "context"
	"edu-evaluation-backed/api/v1/base_info/course"
	student_i "edu-evaluation-backed/api/v1/base_info/student"
	teacher_i "edu-evaluation-backed/api/v1/base_info/teacher"
	"edu-evaluation-backed/internal/biz/base_info"
	"edu-evaluation-backed/internal/data/dal"
	"encoding/json"
	"errors"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// CourseService 课程信息服务
// 提供课程详情查询、列表查询、编辑、导入、删除等功能
type CourseService struct {
	courseDal *dal.CourseDal
	courseUC  *base_info.CourseUseCase
}

// Detail 获取课程详情
// 根据课程ID获取课程详细信息，包含关联的教师列表和学生列表
// ctx: 上下文
// req: 获取课程详情请求，包含课程ID
// 返回值: 课程详情响应，包含课程基本信息、教师列表、学生列表，错误信息
func (c CourseService) Detail(ctx context.Context, req *course.GetCourseDetailReq) (*course.GetCourseDetailResp, error) {
	cs, err := c.courseDal.Detail(uint(req.CourseId))
	if err != nil {
		return nil, err
	}
	resp := &course.GetCourseDetailResp{
		Message: "success",
		Data: &course.CourseList{
			Id:          uint32(cs.ID),
			CourseName:  cs.CourseName,
			ClassName:   cs.ClassName,
			TeacherList: make([]*teacher_i.TeacherInfo, 0),
			StudentList: make([]*student_i.StudentInfo, 0),
			Status:      int32(cs.Status),
		},
	}
	for _, t := range cs.Teachers {
		resp.Data.TeacherList = append(resp.Data.TeacherList, &teacher_i.TeacherInfo{
			Id:     uint32(t.ID),
			Name:   t.Name,
			WorkNo: t.WorkNo,
			Email:  t.Email,
			Sex:    t.Sex,
		})
	}
	for _, s := range cs.Students {
		resp.Data.StudentList = append(resp.Data.StudentList, &student_i.StudentInfo{
			Id:        uint32(s.ID),
			Name:      s.Name,
			StudentNo: s.StudentNo,
			Sex:       s.Sex,
		})
	}
	return resp, nil
}

// Edit 编辑课程信息
// 更新课程的基本信息（课程名称、班级名称），并重新绑定教师列表
// ctx: 上下文
// req: 编辑课程请求，包含课程ID、更新的课程信息和教师ID列表
// 返回值: 编辑成功后返回更新后的课程信息，错误信息
func (c CourseService) Edit(ctx context.Context, req *course.EditCourseReq) (*course.EditCourseResp, error) {
	// 更新课程基本信息
	if req.CourseId == 0 {
		return nil, errors.New("课程ID不能为空")
	}

	// 更新课程名称和班级名称
	if req.CourseName != "" || req.ClassName != "" {
		err := c.courseDal.UpdateCourse(uint(req.CourseId), req.CourseName, req.ClassName)
		if err != nil {
			return nil, err
		}
	}

	// 添加教师到课程
	if len(req.TeacherIds) > 0 {
		err := c.courseDal.AddTeachers(uint(req.CourseId), req.TeacherIds)
		if err != nil {
			return nil, err
		}
	}

	// 查询更新后的课程详情并返回
	cs, err := c.courseDal.Detail(uint(req.CourseId))
	if err != nil {
		return nil, err
	}

	resp := &course.EditCourseResp{
		Message: "修改成功",
		Data: &course.CourseList{
			Id:          uint32(cs.ID),
			CourseName:  cs.CourseName,
			ClassName:   cs.ClassName,
			TeacherList: make([]*teacher_i.TeacherInfo, 0),
			Status:      int32(cs.Status),
		},
	}
	for _, t := range cs.Teachers {
		resp.Data.TeacherList = append(resp.Data.TeacherList, &teacher_i.TeacherInfo{
			Id:     uint32(t.ID),
			Name:   t.Name,
			WorkNo: t.WorkNo,
			Email:  t.Email,
			Sex:    t.Sex,
		})
	}
	return resp, nil
}

// List 获取课程列表
// 支持分页查询课程列表，每个课程包含关联的教师信息
// ctx: 上下文
// req: 获取课程列表请求，包含页码和每页条数
// 返回值: 课程列表响应，包含数据列表和总数，错误信息
func (c CourseService) List(ctx context.Context, req *course.GetCourseListReq) (*course.GetCourseListResp, error) {
	cs, tot, err := c.courseDal.List(int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}
	rsp := &course.GetCourseListResp{
		Message: "success",
		Data:    make([]*course.CourseList, 0),
		Total:   tot,
	}
	for _, cItem := range *cs {
		rsp.Data = append(rsp.Data, &course.CourseList{
			Id:          uint32(cItem.ID),
			CourseName:  cItem.CourseName,
			ClassName:   cItem.ClassName,
			TeacherList: make([]*teacher_i.TeacherInfo, 0),
			Status:      int32(cItem.Status),
		})
		for _, t := range cItem.Teachers {
			rsp.Data[len(rsp.Data)-1].TeacherList = append(rsp.Data[len(rsp.Data)-1].TeacherList, &teacher_i.TeacherInfo{
				Id:     uint32(t.ID),
				Name:   t.Name,
				WorkNo: t.WorkNo,
				Email:  t.Email,
				Sex:    t.Sex,
			})
		}
	}
	return rsp, nil
}

// Import 导入课程信息Excel文件
// 从Excel的Sheet1中读取课程数据，批量创建课程并关联学生
// Excel格式要求：第1列课程名称，第2列班级名称，第4列学生学号，第一行为表头
// 相同课程+班级的学生学号会合并到同一个课程中
// ctx: HTTP上下文，包含上传的文件
// 返回值: 错误信息，成功返回nil，导入过程中的错误会在返回消息中体现
func (c CourseService) Import(ctx http.Context) error {
	req := ctx.Request()
	file, _, err := req.FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()
	iLog := c.courseUC.Import(file)
	if iLog == "" {
		iLog = "导入成功"
	}
	ctx.Response().WriteHeader(200)
	resp, _ := json.Marshal(map[string]string{
		"message": iLog,
	})
	_, _ = ctx.Response().Write(resp)
	return nil
}

// Delete 删除课程
// 根据课程ID删除课程，删除前会清除课程与教师和学生的关联
// ctx: 上下文
// req: 删除课程请求，包含课程ID
// 返回值: 删除成功响应，错误信息
func (c CourseService) Delete(ctx context.Context, req *course.DeleteCourseReq) (*course.DeleteCourseResp, error) {
	// 调用biz层删除
	err := c.courseUC.DeleteCourse(uint(req.Id))
	if err != nil {
		return nil, err
	}

	return &course.DeleteCourseResp{
		Message: "删除成功",
	}, nil
}

// NewCourseService 创建课程信息服务实例
// courseDal: 课程数据访问层
// courseUC: 课程业务用例
// 返回值: 课程信息服务实例指针
func NewCourseService(courseDal *dal.CourseDal, courseUC *base_info.CourseUseCase) *CourseService {
	return &CourseService{
		courseDal: courseDal,
		courseUC:  courseUC,
	}
}
