package cache

import "fmt"

// 缓存 Key 前缀
const (
	PrefixTask     = "edu_eval:task"
	PrefixEval     = "edu_eval:eval"
	PrefixCourse   = "edu_eval:course"
	PrefixLogin    = "edu_eval:login"
)

// 任务列表 Key: edu_eval:task:list:{status}:{page}:{size}
func TaskListKey(status, page, size int) string {
	return fmt.Sprintf("%s:list:%d:%d:%d", PrefixTask, status, page, size)
}

// 学生在某任务下的课程+教师列表: edu_eval:task:courses:{taskID}:{studentNo}
func TaskCoursesKey(taskID uint, studentNo string) string {
	return fmt.Sprintf("%s:courses:%d:%s", PrefixTask, taskID, studentNo)
}

// 评教状态检查: edu_eval:eval:check:{taskID}:{courseID}:{studentNo}:{teacherID}
func EvalCheckKey(taskID, courseID uint, studentNo string, teacherID uint) string {
	return fmt.Sprintf("%s:check:%d:%d:%s:%d", PrefixEval, taskID, courseID, studentNo, teacherID)
}

// 课程详情: edu_eval:course:detail:{courseID}
func CourseDetailKey(courseID uint) string {
	return fmt.Sprintf("%s:detail:%d", PrefixCourse, courseID)
}

// 登录限流: edu_eval:login:rate:{stuNo}
func LoginRateKey(stuNo string) string {
	return fmt.Sprintf("%s:rate:%s", PrefixLogin, stuNo)
}

// 任务列表通配符: edu_eval:task:list:*
func TaskListPattern() string {
	return fmt.Sprintf("%s:list:*", PrefixTask)
}

// 评教状态通配符（某任务+学生）: edu_eval:eval:check:{taskID}:*:{studentNo}:*
func EvalCheckPattern(taskID uint, studentNo string) string {
	return fmt.Sprintf("%s:check:%d:*:%s:*", PrefixEval, taskID, studentNo)
}

// 课程详情通配符: edu_eval:course:detail:*
func CourseDetailPattern() string {
	return fmt.Sprintf("%s:detail:*", PrefixCourse)
}
