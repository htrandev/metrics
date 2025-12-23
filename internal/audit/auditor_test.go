package audit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var _ Observer = (*mockObserver)(nil)

type mockObserver struct {
	id uuid.UUID
}

func (m *mockObserver) Update(ctx context.Context, info AuditInfo) {}

func (m *mockObserver) GetID() string {
	return m.id.String()
}

func TestRegister(t *testing.T) {
	id := uuid.New()
	o := &mockObserver{
		id: id,
	}

	a := NewAuditor()
	a.Register(o)

	_, ok := a.observers[id.String()]
	require.True(t, ok)

	_, ok = a.observers[uuid.New().String()]
	require.False(t, ok)
}

func TestDeregister(t *testing.T) {
	id := uuid.New()
	o := &mockObserver{
		id: id,
	}

	a := NewAuditor()
	a.observers[id.String()] = o
	a.Deregister(o)

	_, ok := a.observers[uuid.New().String()]
	require.False(t, ok)
}
