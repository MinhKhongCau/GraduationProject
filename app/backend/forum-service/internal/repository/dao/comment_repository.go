package dao

import (
	"errors"
	"strconv"

	"gorm.io/gorm"

	"forum-service/internal/app/entity"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) FindByPostID(postID int64) ([]CommentDAO, error) {
	var rows []CommentDAO
	if err := r.db.Where("post_id = ?", postID).Order("path ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CommentRepository) FindByID(id int64) (*CommentDAO, error) {
	var row CommentDAO
	if err := r.db.Where("id = ?", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

// CreateWithPath implements the two-step insert-then-update path
// computation from SPEC.md §3.1: insert with a placeholder path to obtain
// the generated id, then compute the real ltree path from that id (root
// comment: its own id; reply: parent's path + its own id) and update the
// row within the same transaction.
func (r *CommentRepository) CreateWithPath(postID, userID int64, parentID *int64, content string) (*CommentDAO, error) {
	var created CommentDAO
	err := r.db.Transaction(func(tx *gorm.DB) error {
		row := CommentDAO{
			PostID:   postID,
			UserID:   userID,
			ParentID: parentID,
			Path:     entity.LTree("0"),
			Content:  content,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}

		newPath := entity.LTree(strconv.FormatInt(row.ID, 10))
		if parentID != nil {
			var parent CommentDAO
			if err := tx.Where("id = ?", *parentID).First(&parent).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrNotFound
				}
				return err
			}
			newPath = entity.LTree(parent.Path.String() + "." + strconv.FormatInt(row.ID, 10))
		}

		if err := tx.Model(&CommentDAO{}).Where("id = ?", row.ID).Update("path", newPath).Error; err != nil {
			return err
		}
		row.Path = newPath
		created = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *CommentRepository) UpdateContent(id int64, content string) error {
	return r.db.Model(&CommentDAO{}).Where("id = ?", id).Update("content", content).Error
}

func (r *CommentRepository) SoftDelete(id int64) error {
	return r.db.Model(&CommentDAO{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}
