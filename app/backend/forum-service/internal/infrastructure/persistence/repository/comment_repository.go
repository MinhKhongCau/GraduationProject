package repository

import (
	"errors"
	"strconv"

	"gorm.io/gorm"

	commentdomain "forum-service/internal/domain/comment"
	"forum-service/internal/infrastructure/persistence/models"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) FindByPostID(postID int64) ([]models.CommentDAO, error) {
	var rows []models.CommentDAO
	if err := r.db.Where("post_id = ?", postID).Order("path ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CommentRepository) FindByID(id int64) (*models.CommentDAO, error) {
	var row models.CommentDAO
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
func (r *CommentRepository) CreateWithPath(postID int64, userID string, parentID *int64, content string) (*models.CommentDAO, error) {
	var created models.CommentDAO
	err := r.db.Transaction(func(tx *gorm.DB) error {
		row := models.CommentDAO{
			PostID:   postID,
			UserID:   userID,
			ParentID: parentID,
			Path:     commentdomain.LTree("0"),
			Content:  content,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}

		newPath := commentdomain.LTree(strconv.FormatInt(row.ID, 10))
		if parentID != nil {
			var parent models.CommentDAO
			if err := tx.Where("id = ?", *parentID).First(&parent).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrNotFound
				}
				return err
			}
			newPath = commentdomain.LTree(parent.Path.String() + "." + strconv.FormatInt(row.ID, 10))
		}

		if err := tx.Model(&models.CommentDAO{}).Where("id = ?", row.ID).Update("path", newPath).Error; err != nil {
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
	return r.db.Model(&models.CommentDAO{}).Where("id = ?", id).Update("content", content).Error
}

func (r *CommentRepository) SoftDelete(id int64) error {
	return r.db.Model(&models.CommentDAO{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}
