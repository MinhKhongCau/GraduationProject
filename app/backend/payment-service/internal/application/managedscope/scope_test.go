package managedscope

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type stubResolver struct {
	ids []uuid.UUID
	err error
}

func (s stubResolver) ListManagedExpertIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return s.ids, s.err
}

func TestResolve(t *testing.T) {
	admin, managed, other := uuid.New(), uuid.New(), uuid.New()
	resolver := stubResolver{ids: []uuid.UUID{managed, managed}}

	scope, err := Resolve(context.Background(), resolver, admin, nil)
	if err != nil || len(scope.ExpertIDs()) != 1 || !scope.Contains(managed) || scope.Contains(other) {
		t.Fatalf("unexpected scope %+v err=%v", scope.ExpertIDs(), err)
	}
	if narrowed, err := Resolve(context.Background(), resolver, admin, &managed); err != nil || len(narrowed.ExpertIDs()) != 1 {
		t.Fatalf("expected narrowed scope, err=%v", err)
	}
	if _, err := Resolve(context.Background(), resolver, admin, &other); !errors.Is(err, ErrNotManagedExpert) {
		t.Fatalf("expected ErrNotManagedExpert, got %v", err)
	}
	if _, err := Resolve(context.Background(), nil, admin, nil); !errors.Is(err, ErrResolverUnavailable) {
		t.Fatalf("expected ErrResolverUnavailable, got %v", err)
	}
	if empty, err := Resolve(context.Background(), stubResolver{}, admin, nil); err != nil || !empty.IsEmpty() {
		t.Fatalf("expected empty scope, err=%v", err)
	}
}
