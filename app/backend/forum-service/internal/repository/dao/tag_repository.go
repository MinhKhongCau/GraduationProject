package dao

import (
	"errors"

	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) FindAll() ([]TagDAO, error) {
	var rows []TagDAO
	if err := r.db.Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *TagRepository) FindBySlug(slug string) (*TagDAO, error) {
	var row TagDAO
	if err := r.db.Where("slug = ?", slug).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

// FindOrCreateBySlug looks up a tag by slug, creating it if it doesn't
// exist yet — used when a post is created/updated with a tag name that
// hasn't been seen before (README.md "Tags" section).
func (r *TagRepository) FindOrCreateBySlug(tx *gorm.DB, name, slug string) (*TagDAO, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var row TagDAO
	if err := db.Where("slug = ?", slug).First(&row).Error; err == nil {
		return &row, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	created := TagDAO{Name: name, Slug: slug}
	if err := db.Create(&created).Error; err != nil {
		return nil, err
	}
	return &created, nil
}

// FindByPostID returns the tags attached to a post.
func (r *TagRepository) FindByPostID(postID int64) ([]TagDAO, error) {
	var rows []TagDAO
	if err := r.db.
		Joins("JOIN post_tags pt ON pt.tag_id = tags.id").
		Where("pt.post_id = ?", postID).
		Order("tags.name ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// FindByPostIDs returns tags for multiple posts at once, keyed by post id —
// used to avoid N+1 queries when rendering a list of posts.
func (r *TagRepository) FindByPostIDs(postIDs []int64) (map[int64][]TagDAO, error) {
	if len(postIDs) == 0 {
		return map[int64][]TagDAO{}, nil
	}
	type row struct {
		PostID int64
		TagDAO
	}
	var rows []row
	if err := r.db.Table("tags").
		Select("post_tags.post_id, tags.*").
		Joins("JOIN post_tags ON post_tags.tag_id = tags.id").
		Where("post_tags.post_id IN ?", postIDs).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[int64][]TagDAO, len(postIDs))
	for _, r := range rows {
		result[r.PostID] = append(result[r.PostID], r.TagDAO)
	}
	return result, nil
}

// ReplacePostTags atomically replaces a post's tag associations with the
// given tag ids (used on post create/update).
func (r *TagRepository) ReplacePostTags(tx *gorm.DB, postID int64, tagIDs []int64) error {
	if err := tx.Where("post_id = ?", postID).Delete(&PostTagDAO{}).Error; err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	links := make([]PostTagDAO, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		links = append(links, PostTagDAO{PostID: postID, TagID: tagID})
	}
	return tx.Create(&links).Error
}
