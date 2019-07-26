package ssdb

import (
	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
	"github.com/liyue201/btc-scan/storage"
)

func NewSSDBStorage(config *conf.Config) (storage.Storage, error) {
	s := SSDBStorage{}
	p, err := gossdb.NewPool(config)
	if err != nil {
		return nil, err
	}
	s.pool = p
	return &s, nil
}

type SSDBStorage struct {
	pool *gossdb.Connectors
}

func (s *SSDBStorage) Set(key string, value string) error {
	c, err := s.pool.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()
	err = c.Set(key, value)
	return err
}

func (s *SSDBStorage) Get(key string) (string, error) {
	c, err := s.pool.NewClient()
	if err != nil {
		return "", nil
	}
	defer c.Close()

	value, err := c.Get(key)
	if err != nil {
		return "", err
	}
	return value.String(), nil
}

func (s *SSDBStorage) Delete(key string) error {
	c, err := s.pool.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()
	err = c.Del(key)
	return err
}

func (s *SSDBStorage) Scan(keyStart, keyEnd string, limit int64) (map[string]string, error) {
	m := map[string]string{}

	c, err := s.pool.NewClient()
	if err != nil {
		return m, nil
	}
	defer c.Close()

	values, err := c.Scan(keyStart, keyEnd, limit)
	if err != nil {
		return m, err
	}
	for key, v := range values{
		m[key] = v.String()
	}
	return m, nil
}

func (s *SSDBStorage) RScan(keyStart, keyEnd string, limit int64) (map[string]string, error) {
	m := map[string]string{}

	c, err := s.pool.NewClient()
	if err != nil {
		return m, nil
	}
	defer c.Close()

	values, err := c.Rscan(keyStart, keyEnd, limit)
	if err != nil {
		return m, err
	}
	for key, v := range values{
		m[key] = v.String()
	}
	return m, nil
}

func (s *SSDBStorage) Close() error {
	s.pool.Close()
	return nil
}