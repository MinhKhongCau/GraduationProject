package dao

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

type PostListFilter struct {
	CategoryID *int64
	TagSlug    string
	AuthorID   *string
	Status     string
	Search     string
	Page       int
	PageSize   int
}

func (r *PostRepository) List(f PostListFilter) ([]PostDAO, int64, error) {
	q := r.db.Model(&PostDAO{}).Where("posts.deleted_at IS NULL")

	if f.Status != "" {
		q = q.Where("posts.status = ?", f.Status)
	} else {
		q = q.Where("posts.status = ?", "PUBLISHED")
	}
	if f.CategoryID != nil {
		q = q.Where("posts.category_id = ?", *f.CategoryID)
	}
	if f.AuthorID != nil {
		q = q.Where("posts.author_id = ?", *f.AuthorID)
	}
	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		q = q.Where("(LOWER(posts.title) LIKE ? OR LOWER(posts.summary) LIKE ?)", like, like)
	}
	if f.TagSlug != "" {
		q = q.Joins("JOIN post_tags pt ON pt.post_id = posts.id").
			Joins("JOIN tags t ON t.id = pt.tag_id AND t.slug = ?", f.TagSlug)
	}

	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	var rows []PostDAO
	if err := q.Order("posts.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *PostRepository) FindBySlug(slug string) (*PostDAO, error) {
	var row PostDAO
	if err := r.db.Where("slug = ? AND deleted_at IS NULL", slug).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (r *PostRepository) FindByID(id int64) (*PostDAO, error) {
	var row PostDAO
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (r *PostRepository) ExistsBySlug(slug string) (bool, error) {
	var count int64
	if err := r.db.Model(&PostDAO{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostRepository) FindByIDs(ids []int64) ([]PostDAO, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []PostDAO
	if err := r.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostRepository) Create(row *PostDAO) error {
	return r.db.Create(row).Error
}

func (r *PostRepository) Update(row *PostDAO) error {
	return r.db.Model(&PostDAO{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
		"category_id":   row.CategoryID,
		"title":         row.Title,
		"summary":       row.Summary,
		"content":       row.Content,
		"thumbnail_url": row.ThumbnailURL,
	}).Error
}

func (r *PostRepository) UpdateStatus(id int64, status string) error {
	return r.db.Model(&PostDAO{}).Where("id = ?", id).Update("status", status).Error
}

func (r *PostRepository) SoftDelete(id int64) error {
	return r.db.Model(&PostDAO{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (r *PostRepository) IncrementViewCount(id int64) error {
	return r.db.Model(&PostDAO{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *PostRepository) IncrementCommentCount(id int64, delta int) error {
	return r.db.Model(&PostDAO{}).Where("id = ?", id).UpdateColumn("comment_count", gorm.Expr("comment_count + ?", delta)).Error
}

func (r *PostRepository) IncrementLikeCount(id int64, delta int) error {
	return r.db.Model(&PostDAO{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + ?", delta)).Error
}

func (r *PostRepository) IncrementBookmarkCount(id int64, delta int) error {
	return r.db.Model(&PostDAO{}).Where("id = ?", id).UpdateColumn("bookmark_count", gorm.Expr("bookmark_count + ?", delta)).Error
}
