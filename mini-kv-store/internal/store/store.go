package store

import (
	"errors"
	"sync"
)

var ErrKeyNotFound = errors.New("key not found")

type Store interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
}

type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		items: make(map[string]string),
	}
}

func (s *MemoryStore) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.items[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return val, nil
}

func (s *MemoryStore) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = value
	return nil
}

func (s *MemoryStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[key]; !ok {
		return ErrKeyNotFound
	}
	delete(s.items, key)
	return nil
}
