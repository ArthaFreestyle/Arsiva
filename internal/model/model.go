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

// NewPageMetaData builds paging metadata from the requested page/size and the
// total number of rows matching the query (before LIMIT/OFFSET).
func NewPageMetaData(page int, size int, total int) *PageMetaData {
	totalPage := 0
	if size > 0 {
		totalPage = (total + size - 1) / size
	}
	return &PageMetaData{
		Page:      page,
		Size:      size,
		TotalItem: total,
		TotalPage: totalPage,
	}
}
