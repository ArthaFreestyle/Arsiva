package model

type WebResponse[T any] struct {
	Data   T             `json:"data"`
	Paging *PageMetaData `json:"paging,omitempty"`
	Errors string        `json:"errors,omitempty"`
}

type PageMetaData struct {
	Page      int `json:"page"`
	Size      int `json:"size"`
	TotalItem int `json:"total_item"`
	TotalPage int `json:"total_page"`
}

const (
	DefaultPageSize = 10
	MaxPageSize     = 50
)

// NormalizePaging applies the paging rules shared by every list endpoint:
// page starts at 1, and a size outside 1..MaxPageSize falls back to
// DefaultPageSize.
func NormalizePaging(page int, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > MaxPageSize {
		size = DefaultPageSize
	}
	return page, size
}

// NewPageMetaData builds paging metadata from the requested page/size and the
// total number of rows matching the query (before LIMIT/OFFSET). page/size are
// normalized the same way the usecases do, so the echoed values match the
// data actually returned.
func NewPageMetaData(page int, size int, total int) *PageMetaData {
	page, size = NormalizePaging(page, size)
	return &PageMetaData{
		Page:      page,
		Size:      size,
		TotalItem: total,
		TotalPage: (total + size - 1) / size,
	}
}
