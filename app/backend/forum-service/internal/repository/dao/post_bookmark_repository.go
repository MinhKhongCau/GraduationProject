package dao

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostBookmarkRepository struct {
	db *gorm.DB
}

func NewPostBookmarkRepository(db *gorm.DB) *PostBookmarkRepository {
	return &PostBookmarkRepository{db: db}
}

func (r *PostBookmarkRepository) Create(postID, userID int64) (bool, error) {
	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&PostBookmarkDAO{PostID: postID, UserID: userID})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *PostBookmarkRepository) Delete(postID, userID int64) (bool, error) {
	result := r.db.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&PostBookmarkDAO{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// FindPostIDsByUser returns the ids of posts the user has bookmarked, most
// recently bookmarked first — used by GET /users/me/bookmarks.
func (r *PostBookmarkRepository) FindPostIDsByUser(userID int64, page, pageSize int) ([]int64, int64, error) {
	var total int64
	if err := r.db.Model(&PostBookmarkDAO{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var rows []PostBookmarkDAO
	if err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.PostID)
	}
	return ids, total, nil
}
