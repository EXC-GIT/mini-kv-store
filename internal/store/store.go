package store

import (
	"errors"
	"sync"
	"time"
)

var ErrKeyNotFound = errors.New("key not found")

type item struct {
	value     string
	expiresAt time.Time
}

type Store interface {
	Get(key string) (string, error)
	Set(key, value string) error
	SetWithTTL(key, value string, ttl time.Duration) error
	Delete(key string) error
}

type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]*item
}

func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		items: make(map[string]*item),
	}
	go store.startExpiryChecker()
	return store
}

func (s *MemoryStore) startExpiryChecker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for k, v := range s.items {
			if !v.expiresAt.IsZero() && v.expiresAt.Before(now) {
				delete(s.items, k)
			}
		}
		s.mu.Unlock()
	}
}

func (s *MemoryStore) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[key]
	if !ok || (!item.expiresAt.IsZero() && item.expiresAt.Before(time.Now())) {
		return "", ErrKeyNotFound
	}
	return item.value, nil
}

func (s *MemoryStore) Set(key, value string) error {
	return s.SetWithTTL(key, value, 0)
}

func (s *MemoryStore) SetWithTTL(key, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	expires := time.Time{}
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}

	s.items[key] = &item{
		value:     value,
		expiresAt: expires,
	}
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
