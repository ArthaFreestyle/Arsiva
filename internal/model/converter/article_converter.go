package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToArticleResponse(article *entity.Article) *model.ArticleResponse {
	var content, excerpt, thumbnail, createdAt string
	if article.Konten != nil {
		content = *article.Konten
	}
	if article.Excerpt != nil {
		excerpt = *article.Excerpt
	}
	if article.Thumbnail != nil {
		thumbnail = *article.Thumbnail
	}
	if article.CreatedAt != nil {
		createdAt = article.CreatedAt.Format("2006-01-02 15:04:05")
	}

	return &model.ArticleResponse{
		ArticleId: article.ArticleId,
		Slug: article.Slug,
		Title: article.Judul,
		Content: content,
		Category: model.ArticleCategoryResponse{
			ArticleCategoryId: article.Kategori.ArticleCategoryId,
			NamaKategori: article.Kategori.NamaKategori,
		},
		Status: article.Status,
		Excerpt: excerpt,
		CreatedBy: model.UserResponse{
			ID: article.CreatedBy.UserId,
			Username: article.CreatedBy.Username,
		},
		CreatedAt: createdAt,
		Thumbnail: thumbnail,
	}
}

func ToArticleResponses(articles []*entity.Article) []*model.ArticleResponse {
	var articleResponses []*model.ArticleResponse
	for _,article := range articles {
		articleResponses = append(articleResponses,ToArticleResponse(article))
	}
	return articleResponses
}