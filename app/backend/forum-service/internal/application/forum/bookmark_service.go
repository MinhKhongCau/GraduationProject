package forum

import (
	"errors"

	postdomain "forum-service/internal/domain/post"
	"forum-service/internal/infrastructure/persistence/repository"
)

type BookmarkService struct {
	bookmarkRepo *repository.PostBookmarkRepository
	postRepo     *repository.PostRepository
	tagRepo      *repository.TagRepository
}

func NewBookmarkService(bookmarkRepo *repository.PostBookmarkRepository, postRepo *repository.PostRepository, tagRepo *repository.TagRepository) *BookmarkService {
	return &BookmarkService{bookmarkRepo: bookmarkRepo, postRepo: postRepo, tagRepo: tagRepo}
}

func (s *BookmarkService) Bookmark(postID int64, userID string) (int, error) {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, ErrNotFound
		}
		return 0, err
	}

	created, err := s.bookmarkRepo.Create(postID, userID)
	if err != nil {
		return 0, err
	}
	if !created {
		return post.BookmarkCount, nil
	}
	if err := s.postRepo.IncrementBookmarkCount(postID, 1); err != nil {
		return 0, err
	}
	return post.BookmarkCount + 1, nil
}

func (s *BookmarkService) RemoveBookmark(postID int64, userID string) (int, error) {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, ErrNotFound
		}
		return 0, err
	}

	removed, err := s.bookmarkRepo.Delete(postID, userID)
	if err != nil {
		return 0, err
	}
	if !removed {
		return post.BookmarkCount, nil
	}
	if err := s.postRepo.IncrementBookmarkCount(postID, -1); err != nil {
		return 0, err
	}
	return post.BookmarkCount - 1, nil
}

// ListMine returns the caller's bookmarked posts, most recently bookmarked
// first (README.md's GET /users/me/bookmarks), reusing PostListResult's
// shape since the response mirrors GET /posts.
func (s *BookmarkService) ListMine(userID string, page, pageSize int) (*PostListResult, error) {
	ids, total, err := s.bookmarkRepo.FindPostIDsByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	rows, err := s.postRepo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]int, len(rows))
	for i, r := range rows {
		byID[r.ID] = i
	}

	tagsByPost, err := s.tagRepo.FindByPostIDs(ids)
	if err != nil {
		return nil, err
	}

	items := make([]postdomain.Post, 0, len(ids))
	for _, id := range ids {
		idx, ok := byID[id]
		if !ok {
			continue // post was soft-deleted since being bookmarked
		}
		row := rows[idx]
		e := row.ToEntity()
		for _, t := range tagsByPost[row.ID] {
			e.Tags = append(e.Tags, t.Slug)
		}
		items = append(items, e)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	return &PostListResult{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}
