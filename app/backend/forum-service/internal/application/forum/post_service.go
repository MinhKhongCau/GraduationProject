package forum

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	postdomain "forum-service/internal/domain/post"
	"forum-service/internal/infrastructure/persistence/models"
	"forum-service/internal/infrastructure/persistence/repository"
)

var slugInvalidChars = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	slug := strings.ToLower(strings.TrimSpace(s))
	slug = slugInvalidChars.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

// validPostTransitions encodes the DRAFT -> PUBLISHED -> ARCHIVED lifecycle
// from SPEC.md §3.4: ARCHIVED is terminal, there is no transition back to
// PUBLISHED.
var validPostTransitions = map[postdomain.PostStatus]map[postdomain.PostStatus]bool{
	postdomain.PostStatusDraft: {
		postdomain.PostStatusPublished: true,
		postdomain.PostStatusArchived:  true,
	},
	postdomain.PostStatusPublished: {
		postdomain.PostStatusArchived: true,
	},
	postdomain.PostStatusArchived: {},
}

type PostService struct {
	db           *gorm.DB
	postRepo     *repository.PostRepository
	categoryRepo *repository.CategoryRepository
	tagRepo      *repository.TagRepository
	events       EventPublisher
}

func NewPostService(db *gorm.DB, postRepo *repository.PostRepository, categoryRepo *repository.CategoryRepository, tagRepo *repository.TagRepository, events EventPublisher) *PostService {
	return &PostService{db: db, postRepo: postRepo, categoryRepo: categoryRepo, tagRepo: tagRepo, events: events}
}

type PostListResult struct {
	Items    []postdomain.Post
	Page     int
	PageSize int
	Total    int64
}

func (s *PostService) List(f repository.PostListFilter) (*PostListResult, error) {
	rows, total, err := s.postRepo.List(f)
	if err != nil {
		return nil, err
	}

	items, err := s.attachTags(rows)
	if err != nil {
		return nil, err
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	return &PostListResult{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

// GetBySlug returns a post by slug, applying the visibility rule from
// SPEC.md §3.4 (PUBLISHED is visible to everyone; DRAFT/ARCHIVED only to
// the author or an ADMIN) and incrementing view_count on success.
func (s *PostService) GetBySlug(slug string, requesterID string, isAdmin bool) (*postdomain.Post, error) {
	row, err := s.postRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	e := row.ToEntity()
	if !e.IsVisibleTo(requesterID, isAdmin) {
		return nil, ErrNotFound
	}

	if err := s.postRepo.IncrementViewCount(row.ID); err != nil {
		return nil, err
	}
	e.ViewCount++

	tags, err := s.tagRepo.FindByPostID(row.ID)
	if err != nil {
		return nil, err
	}
	for _, t := range tags {
		e.Tags = append(e.Tags, t.Slug)
	}

	return &e, nil
}

type CreatePostInput struct {
	CategoryID   int64
	AuthorID     string
	Title        string
	Summary      string
	Content      string
	ThumbnailURL string
	Status       postdomain.PostStatus
	Tags         []string
}

func (s *PostService) Create(in CreatePostInput) (*postdomain.Post, error) {
	if _, err := s.categoryRepo.FindByID(in.CategoryID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	status := in.Status
	if status == "" {
		status = postdomain.PostStatusPublished
	}

	slug, err := s.uniqueSlug(in.Title)
	if err != nil {
		return nil, err
	}

	row := models.PostDAO{
		CategoryID:   in.CategoryID,
		AuthorID:     in.AuthorID,
		Title:        in.Title,
		Slug:         slug,
		Summary:      in.Summary,
		Content:      in.Content,
		ThumbnailURL: in.ThumbnailURL,
		Status:       string(status),
	}

	tagRows, err := s.createWithTags(&row, in.Tags)
	if err != nil {
		return nil, err
	}

	e := row.ToEntity()
	for _, t := range tagRows {
		e.Tags = append(e.Tags, t.Slug)
	}

	if e.Status == postdomain.PostStatusPublished {
		s.events.PostCreated(PostCreatedEvent{
			PostID: e.ID, AuthorID: e.AuthorID, CategoryID: e.CategoryID,
			Title: e.Title, PublishedAt: e.CreatedAt,
		})
	}

	return &e, nil
}

func (s *PostService) createWithTags(row *models.PostDAO, tags []string) ([]models.TagDAO, error) {
	var tagRows []models.TagDAO
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		tagIDs, resolved, err := s.resolveTags(tx, tags)
		if err != nil {
			return err
		}
		tagRows = resolved
		return s.tagRepo.ReplacePostTags(tx, row.ID, tagIDs)
	})
	return tagRows, err
}

// resolveTags finds-or-creates a tag row for each name (README.md's Posts
// section: tags supplied by name that don't exist yet are auto-created).
func (s *PostService) resolveTags(tx *gorm.DB, names []string) ([]int64, []models.TagDAO, error) {
	ids := make([]int64, 0, len(names))
	rows := make([]models.TagDAO, 0, len(names))
	for _, name := range names {
		tagSlug := slugify(name)
		if tagSlug == "" {
			continue
		}
		tag, err := s.tagRepo.FindOrCreateBySlug(tx, name, tagSlug)
		if err != nil {
			return nil, nil, err
		}
		ids = append(ids, tag.ID)
		rows = append(rows, *tag)
	}
	return ids, rows, nil
}

type UpdatePostInput struct {
	CategoryID   *int64
	Title        *string
	Summary      *string
	Content      *string
	ThumbnailURL *string
	Tags         *[]string
}

func (s *PostService) Update(id int64, requesterID string, isAdmin bool, in UpdatePostInput) (*postdomain.Post, error) {
	row, err := s.postRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !isAdmin && row.AuthorID != requesterID {
		return nil, ErrForbidden
	}

	if in.CategoryID != nil {
		if _, err := s.categoryRepo.FindByID(*in.CategoryID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		row.CategoryID = *in.CategoryID
	}
	if in.Title != nil {
		row.Title = *in.Title
	}
	if in.Summary != nil {
		row.Summary = *in.Summary
	}
	if in.Content != nil {
		row.Content = *in.Content
	}
	if in.ThumbnailURL != nil {
		row.ThumbnailURL = *in.ThumbnailURL
	}

	var tagRows []models.TagDAO
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.PostDAO{}).Where("id = ?", row.ID).Updates(map[string]interface{}{
			"category_id":   row.CategoryID,
			"title":         row.Title,
			"summary":       row.Summary,
			"content":       row.Content,
			"thumbnail_url": row.ThumbnailURL,
		}).Error; err != nil {
			return err
		}

		if in.Tags != nil {
			tagIDs, resolved, err := s.resolveTags(tx, *in.Tags)
			if err != nil {
				return err
			}
			tagRows = resolved
			if err := s.tagRepo.ReplacePostTags(tx, row.ID, tagIDs); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	e := row.ToEntity()
	if in.Tags != nil {
		for _, t := range tagRows {
			e.Tags = append(e.Tags, t.Slug)
		}
	} else {
		tags, err := s.tagRepo.FindByPostID(row.ID)
		if err != nil {
			return nil, err
		}
		for _, t := range tags {
			e.Tags = append(e.Tags, t.Slug)
		}
	}

	return &e, nil
}

// ChangeStatus validates and applies a post status transition (SPEC.md
// §3.4), publishing forum.post.created the first time a post transitions
// into PUBLISHED via this endpoint (mirroring Create's behavior).
func (s *PostService) ChangeStatus(id int64, requesterID string, isAdmin bool, newStatus postdomain.PostStatus) (*postdomain.Post, error) {
	row, err := s.postRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !isAdmin && row.AuthorID != requesterID {
		return nil, ErrForbidden
	}

	current := postdomain.PostStatus(row.Status)
	if current != newStatus && !validPostTransitions[current][newStatus] {
		return nil, ErrInvalidTransition
	}

	wasPublished := current == postdomain.PostStatusPublished
	if err := s.postRepo.UpdateStatus(id, string(newStatus)); err != nil {
		return nil, err
	}
	row.Status = string(newStatus)

	e := row.ToEntity()
	if !wasPublished && newStatus == postdomain.PostStatusPublished {
		s.events.PostCreated(PostCreatedEvent{
			PostID: e.ID, AuthorID: e.AuthorID, CategoryID: e.CategoryID,
			Title: e.Title, PublishedAt: e.UpdatedAt,
		})
	}

	return &e, nil
}

func (s *PostService) SoftDelete(id int64, requesterID string, isAdmin bool) error {
	row, err := s.postRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	if !isAdmin && row.AuthorID != requesterID {
		return ErrForbidden
	}
	return s.postRepo.SoftDelete(id)
}

func (s *PostService) attachTags(rows []models.PostDAO) ([]postdomain.Post, error) {
	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	tagsByPost, err := s.tagRepo.FindByPostIDs(ids)
	if err != nil {
		return nil, err
	}

	items := make([]postdomain.Post, 0, len(rows))
	for _, r := range rows {
		e := r.ToEntity()
		for _, t := range tagsByPost[r.ID] {
			e.Tags = append(e.Tags, t.Slug)
		}
		items = append(items, e)
	}
	return items, nil
}

func (s *PostService) uniqueSlug(title string) (string, error) {
	base := slugify(title)
	if base == "" {
		base = "post"
	}
	slug := base
	for i := 2; ; i++ {
		exists, err := s.postRepo.ExistsBySlug(slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}
