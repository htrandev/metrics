package pool

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type mockStruct struct {
	id    string
	value int
	f     float64
}

func (m *mockStruct) Reset() {
	m.id = ""
	m.value = 0
	m.f = 0.0
}

func TestPool(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		pool := New(func() *mockStruct {
			return &mockStruct{id: uuid.NewString()}
		})

		p1 := pool.Get()
		require.NotNil(t, p1)

		p2 := pool.Get()
		require.NotNil(t, p2)

		require.NotEqual(t, p1.id, p2.id)
	})

	t.Run("put", func(t *testing.T) {
		pool := New(func() *mockStruct {
			return &mockStruct{id: uuid.NewString()}
		})
		p1 := pool.Get()

		p1.id = "new"
		p1.value = 10
		p1.f = 3.14

		pool.Put(p1)

		require.Zero(t, p1.id)
		require.Zero(t, p1.value)
		require.Zero(t, p1.f)
	})
}
