package storage

import (
	"github.com/hashicorp/golang-lru"
)

type StorageWithCache struct {
	store Storage
	cache *lru.Cache
}

func NewStorageWithCache(store Storage, cacheSize int) Storage {
	lruCache, _ := lru.New(cacheSize)

	c := &StorageWithCache{store: store,
		cache: lruCache,
	}
	return c
}

func (s *StorageWithCache) Set(key string, value string) error {
	oldValue, exist := s.cache.Get(key)
	s.cache.Add(key, value)
	err := s.store.Set(key, value)
	if err != nil && exist {
		s.cache.Add(key, oldValue)
	}
	return err
}

func (s *StorageWithCache) Get(key string) (string, error) {
	cacheValue, exist := s.cache.Get(key)
	if exist {
		return cacheValue.(string), nil
	}
	val, err := s.store.Get(key)
	if err == nil{
		s.cache.Add(key, val)
	}
	return val, err
}

func (s *StorageWithCache) Delete(key string) error {
	oldValue, exist := s.cache.Get(key)

	s.cache.Remove(key)
	err := s.store.Delete(key)

	if err != nil && exist {
		s.cache.Add(key, oldValue)
	}
	return err
}

func (s *StorageWithCache) Scan(keyStart, keyEnd string, limit int64) (map[string]string, error) {
	return s.store.Scan(keyStart, keyEnd, limit)
}

func (s *StorageWithCache) RScan(keyStart, keyEnd string, limit int64) (map[string]string, error) {
	return s.store.RScan(keyStart, keyEnd, limit)
}

func (s *StorageWithCache) Close() error  {
	return s.store.Close()
}