package dao

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostLikeRepository struct {
	db *gorm.DB
}

func NewPostLikeRepository(db *gorm.DB) *PostLikeRepository {
	return &PostLikeRepository{db: db}
}

// Create inserts a like, silently doing nothing if the (post_id, user_id)
// pair already exists (composite PK — SPEC.md §3.3 idempotency). Returns
// whether a new row was actually inserted.
func (r *PostLikeRepository) Create(postID int64, userID string) (bool, error) {
	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&PostLikeDAO{PostID: postID, UserID: userID})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *PostLikeRepository) Delete(postID int64, userID string) (bool, error) {
	result := r.db.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&PostLikeDAO{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
