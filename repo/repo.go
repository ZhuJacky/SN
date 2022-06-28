// Package repo TODO
package repo

import (
	"context"
	"time"

	"github.com/ZhuJacky/SN/models"
)

// IArticleRepo represent the article's repository contract
type IArticleRepo interface {
	Fetch(ctx context.Context, createdDate time.Time, num int) (res []models.Article, err error)
}
