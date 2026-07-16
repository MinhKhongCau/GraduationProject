package service

import (
	"errors"

	"forum-service/internal/app/entity"
	"forum-service/internal/repository/dao"
)

type CommentService struct {
	commentRepo *dao.CommentRepository
	postRepo    *dao.PostRepository
	events      *EventPublisher
}

func NewCommentService(commentRepo *dao.CommentRepository, postRepo *dao.PostRepository, events *EventPublisher) *CommentService {
	return &CommentService{commentRepo: commentRepo, postRepo: postRepo, events: events}
}

// Tree returns a post's comments assembled into a nested structure, ordered
// by ltree path (SPEC.md §3.1). Soft-deleted comments are rendered as
// content-less placeholders rather than removed, so reply chains under them
// stay intact (SPEC.md §3.6).
func (s *CommentService) Tree(postID int64) ([]*entity.Comment, error) {
	rows, err := s.commentRepo.FindByPostID(postID)
	if err != nil {
		return nil, err
	}

	byID := make(map[int64]*entity.Comment, len(rows))
	order := make([]int64, 0, len(rows))
	for _, r := range rows {
		e := r.ToEntity()
		if e.IsDeleted() {
			e.Content = ""
		}
		byID[e.ID] = &e
		order = append(order, e.ID)
	}

	var roots []*entity.Comment
	for _, id := range order {
		c := byID[id]
		if c.ParentID == nil {
			roots = append(roots, c)
			continue
		}
		if parent, ok := byID[*c.ParentID]; ok {
			parent.Replies = append(parent.Replies, c)
		} else {
			// Parent not found under this post (shouldn't happen) —
			// surface as a root rather than silently dropping it.
			roots = append(roots, c)
		}
	}

	return roots, nil
}

func (s *CommentService) Create(postID, userID int64, parentID *int64, content string) (*entity.Comment, error) {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var parent *dao.CommentDAO
	if parentID != nil {
		parent, err = s.commentRepo.FindByID(*parentID)
		if err != nil {
			if errors.Is(err, dao.ErrNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		if parent.PostID != postID {
			return nil, ErrNotFound
		}
	}

	row, err := s.commentRepo.CreateWithPath(postID, userID, parentID, content)
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if err := s.postRepo.IncrementCommentCount(postID, 1); err != nil {
		return nil, err
	}

	e := row.ToEntity()

	var parentAuthorID *int64
	if parent != nil {
		parentAuthorID = &parent.UserID
	}
	s.events.CommentCreated(CommentCreatedEvent{
		CommentID: e.ID, PostID: e.PostID, UserID: e.UserID,
		PostAuthorID: post.AuthorID, ParentID: parentID, ParentAuthorID: parentAuthorID,
	})

	return &e, nil
}

func (s *CommentService) Edit(id, requesterID int64, isAdmin bool, content string) (*entity.Comment, error) {
	row, err := s.commentRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !isAdmin && row.UserID != requesterID {
		return nil, ErrForbidden
	}
	if err := s.commentRepo.UpdateContent(id, content); err != nil {
		return nil, err
	}
	row.Content = content
	e := row.ToEntity()
	return &e, nil
}

func (s *CommentService) SoftDelete(id, requesterID int64, isAdmin bool) error {
	row, err := s.commentRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	if !isAdmin && row.UserID != requesterID {
		return ErrForbidden
	}
	if err := s.commentRepo.SoftDelete(id); err != nil {
		return err
	}
	return s.postRepo.IncrementCommentCount(row.PostID, -1)
}
