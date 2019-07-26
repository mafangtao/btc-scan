package mockdb

import (
	"errors"
	"sync"
)

var ErrorNotFound = errors.New("data not found")

type MockDB struct {
	m     map[string]string
	mMutx sync.RWMutex
}

func NewMockDBStorage() *MockDB {
	s := MockDB{
		m: make(map[string]string),
	}
	return &s
}

func (s *MockDB) Set(key string, value string) error {
	s.mMutx.Lock()
	defer s.mMutx.Unlock()

	s.m[key] = value
	return nil
}

func (s *MockDB) Get(key string) (string, error) {
	s.mMutx.RLock()
	defer s.mMutx.RUnlock()

	val, exist := s.m[key]
	if exist {
		return val, nil
	}

	return "", ErrorNotFound
}

func (s *MockDB) Del(key string) error {
	s.mMutx.Lock()
	defer s.mMutx.Lock()

	_, exist := s.m[key]
	if exist {
		delete(s.m, key)
	}
	return nil
}

func (s *MockDB) Scan(keyStart, keyEnd string, limit int64) (map[string]string, error) {

	return nil, nil
}

func (s *MockDB) RScan(keyStart, keyEnd string, limit int64) (map[string]string, error) {
	return nil, nil
}

func (s *MockDB) Close() error {
	return nil
}

