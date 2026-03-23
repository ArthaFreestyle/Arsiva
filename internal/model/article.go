package model

type ArticleCreateRequest struct {
	Title string `json:"title" validate:"required"`
	CategoryId string `json:"category_id" validate:"required"`
	CreatedBy string `json:"created_by" validate:"required"`
}

type ArticleUpdateRequest struct {
	Title string `json:"title" validate:"required"`
	CategoryId string `json:"category_id" validate:"required"`
	Status string `json:"status" validate:"required"`
	Content string `json:"content" validate:"required"`
	Thumbnail string `json:"thumbnail"`
}

type ArticleResponse struct {
	ArticleId string `json:"article_id"`
	Slug string `json:"slug,omitempty"`
	Title string `json:"title"`
	Content string `json:"content,omitempty"`
	Category ArticleCategoryResponse `json:"category"`
	Status string `json:"status"`
	Excerpt string `json:"excerpt,omitempty"`
	CreatedBy UserResponse `json:"created_by"`
	CreatedAt string `json:"created_at"`
	Thumbnail string `json:"thumbnail,omitempty"`
}