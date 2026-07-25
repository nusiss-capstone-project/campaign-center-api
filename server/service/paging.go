package service

const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// pageToOffsetLimit clamps page/pageSize input into a SQL offset and limit.
func pageToOffsetLimit(page, pageSize int) (offset, limit int) {
	if page < 1 {
		page = defaultPage
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return (page - 1) * pageSize, pageSize
}
