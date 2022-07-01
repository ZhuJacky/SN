package service

import (
	"context"
	"my-clean-rchitecture/models"
	"my-clean-rchitecture/repo"
	"time"
)

type articleService struct {
	articleRepo repo.IArticleRepo
}

// NewArticleService will create new an articleUsecase object representation of domain.ArticleUsecase interface
func NewArticleService(a repo.IArticleRepo) IArticleService {
	return &articleService{
		articleRepo: a,
	}
}

// Fetch
func (a *articleService) Fetch(ctx context.Context, createdDate time.Time, num int) (res []models.Article, err error) {
	if num == 0 {
		num = 10
	}
	res, err = a.articleRepo.Fetch(ctx, createdDate, num)
	if err != nil {
		return nil, err
	}
	return
}
