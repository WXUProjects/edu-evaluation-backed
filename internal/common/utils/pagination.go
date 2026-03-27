package utils

// PageNumHandle 处理分页参数，确保分页参数在合理范围内
// page: 当前页码，从1开始
// size: 每页条数
// 返回值: 处理后的页码和每页条数
//
// 处理规则:
// - 如果page <= 0，默认设为1
// - 如果size <= 0，默认设为10
// - 如果size >= 100，限制为100（防止一次性查询过多数据）
func PageNumHandle(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size >= 100 {
		size = 100
	}
	return page, size
}

// CalculateOffset 计算数据库查询的偏移量
// page: 当前页码（从1开始）
// size: 每页条数
// 返回值: 偏移量，用于GORM的Offset方法
func CalculateOffset(page, size int) int {
	page, size = PageNumHandle(page, size)
	return (page - 1) * size
}
