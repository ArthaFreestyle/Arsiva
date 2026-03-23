package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToArticleResponse(article *entity.Article) *model.ArticleResponse {
	return &model.ArticleResponse{
		ArticleId: article.ArticleId,
		Slug: article.Slug,
		Title: article.Judul,
		Content: article.Konten,
		Category: model.ArticleCategoryResponse{
			ArticleCategoryId: article.KategoriId.ArticleCategoryId,
			NamaKategori: article.KategoriId.NamaKategori,
		},
		Status: article.Status,
		Excerpt: article.Excerpt,
		CreatedBy: model.UserResponse{
			ID: article.CreatedBy.UserId,
			Username: article.CreatedBy.Username,
		},
		CreatedAt: article.CreatedAt.Format("2006-01-02 15:04:05"),
		Thumbnail: article.Thumbnail,
	}
}

func ToArticleResponses(articles []*entity.Article) []*model.ArticleResponse {
	var articleResponses []*model.ArticleResponse
	for _,article := range articles {
		articleResponses = append(articleResponses,ToArticleResponse(article))
	}
	return articleResponses
}