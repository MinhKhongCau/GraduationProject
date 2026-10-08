// Package managedscope giới hạn phạm vi dữ liệu Admin được xem/xử lý: Admin duyệt chuyên gia
// (profile-service) trở thành người quản lý chuyên gia đó và chỉ quản lý giao dịch của họ.
package managedscope

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrNotManagedExpert    = errors.New("expert is not managed by this administrator")
	ErrResolverUnavailable = errors.New("managed expert resolver is unavailable")
)

// Resolver trả về auth id các chuyên gia do adminID quản lý (nguồn: profile-service).
type Resolver interface {
	ListManagedExpertIDs(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error)
}

// Scope là tập chuyên gia một Admin quản lý.
type Scope struct {
	expertIDs []uuid.UUID
	index     map[uuid.UUID]struct{}
}

// Resolve lấy phạm vi của adminID. requestedExpert (nếu có) phải nằm trong phạm vi, khi đó
// phạm vi được thu hẹp còn đúng chuyên gia đó.
func Resolve(ctx context.Context, resolver Resolver, adminID uuid.UUID, requestedExpert *uuid.UUID) (Scope, error) {
	if resolver == nil {
		return Scope{}, ErrResolverUnavailable
	}
	if adminID == uuid.Nil {
		return Scope{}, ErrNotManagedExpert
	}
	ids, err := resolver.ListManagedExpertIDs(ctx, adminID)
	if err != nil {
		return Scope{}, fmt.Errorf("%w: %v", ErrResolverUnavailable, err)
	}
	scope := New(ids)
	if requestedExpert != nil {
		if !scope.Contains(*requestedExpert) {
			return Scope{}, ErrNotManagedExpert
		}
		return New([]uuid.UUID{*requestedExpert}), nil
	}
	return scope, nil
}

func New(expertIDs []uuid.UUID) Scope {
	scope := Scope{expertIDs: make([]uuid.UUID, 0, len(expertIDs)), index: make(map[uuid.UUID]struct{}, len(expertIDs))}
	for _, id := range expertIDs {
		if _, seen := scope.index[id]; seen || id == uuid.Nil {
			continue
		}
		scope.index[id] = struct{}{}
		scope.expertIDs = append(scope.expertIDs, id)
	}
	return scope
}

func (s Scope) ExpertIDs() []uuid.UUID { return s.expertIDs }
func (s Scope) IsEmpty() bool          { return len(s.expertIDs) == 0 }

func (s Scope) Contains(expertID uuid.UUID) bool {
	_, ok := s.index[expertID]
	return ok
}
