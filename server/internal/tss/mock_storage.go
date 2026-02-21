package tss

import (
	"context"
	"sync"
)

type MockShardStorage struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMockShardStorage() *MockShardStorage {
	return &MockShardStorage{data: make(map[string][]byte)}
}

func (m *MockShardStorage) Store(shard2ID string, payload []byte) {
	m.mu.Lock()
	copied := make([]byte, len(payload))
	copy(copied, payload)
	m.data[shard2ID] = copied
	m.mu.Unlock()
}

func (m *MockShardStorage) LoadShard2(ctx context.Context, shard2ID string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, ErrContextCanceled
	}

	m.mu.RLock()
	payload, ok := m.data[shard2ID]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrShardNotFound
	}

	copied := make([]byte, len(payload))
	copy(copied, payload)
	return copied, nil
}
