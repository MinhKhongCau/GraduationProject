package service

import (
	"errors"

	"forum-service/internal/repository/dao"
)

type LikeService struct {
	likeRepo *dao.PostLikeRepository
	postRepo *dao.PostRepository
	events   *EventPublisher
}

func NewLikeService(likeRepo *dao.PostLikeRepository, postRepo *dao.PostRepository, events *EventPublisher) *LikeService {
	return &LikeService{likeRepo: likeRepo, postRepo: postRepo, events: events}
}

// Like inserts a like for (postID, userID), no-op if it already exists
// (SPEC.md §3.3 idempotency via the composite PK). Returns the post's
// resulting like_count either way, and publishes forum.post.liked only on
// an actual new insert (SPEC.md §5).
func (s *LikeService) Like(postID, userID int64) (int, error) {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return 0, ErrNotFound
		}
		return 0, err
	}

	created, err := s.likeRepo.Create(postID, userID)
	if err != nil {
		return 0, err
	}
	if !created {
		return post.LikeCount, nil
	}

	if err := s.postRepo.IncrementLikeCount(postID, 1); err != nil {
		return 0, err
	}

	s.events.PostLiked(PostLikedEvent{PostID: postID, UserID: userID, PostAuthorID: post.AuthorID})

	return post.LikeCount + 1, nil
}

func (s *LikeService) Unlike(postID, userID int64) (int, error) {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return 0, ErrNotFound
		}
		return 0, err
	}

	removed, err := s.likeRepo.Delete(postID, userID)
	if err != nil {
		return 0, err
	}
	if !removed {
		return post.LikeCount, nil
	}

	if err := s.postRepo.IncrementLikeCount(postID, -1); err != nil {
		return 0, err
	}

	return post.LikeCount - 1, nil
}
