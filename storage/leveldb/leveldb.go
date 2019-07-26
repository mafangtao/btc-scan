package leveldb

import (
	"github.com/liyue201/btc-scan/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type LevelDBStorage struct {
	db *leveldb.DB
}

func NewLevelDbStorage(path string) (storage.Storage, error) {
	s := LevelDBStorage{}
	var err error
	opts := opt.Options{
		BlockSize:   8 * opt.KiB,
		WriteBuffer: 32 * opt.MiB,
		Filter:      filter.NewBloomFilter(10),
	}
	db, err := leveldb.OpenFile(path, &opts)
	if err != nil {
		return nil, err
	}
	s.db = db
	return &s, nil
}

func (s *LevelDBStorage) Set(key string, value string) error {
	err := s.db.Put([]byte(key), []byte(value), nil)
	return err
}

func (s *LevelDBStorage) Get(key string) (string, error) {
	value, err := s.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return string(value), nil
}

func (s *LevelDBStorage) Delete(key string) error {
	err := s.db.Delete([]byte(key), nil)
	return err
}

//scan key_start key_end limit  列出处于区间 (key_start, key_end] 的 key-value 列表.
func (s *LevelDBStorage) Scan(keyStart, keyEnd string, limit int64) (map[string]string, error) {
	m := map[string]string{}

	srange := &util.Range{Start: []byte(keyStart), Limit: []byte(keyEnd)}
	iter := s.db.NewIterator(srange, nil)

	//leveldb的range 前闭后开，与ssdb相反，跳过第一个

	count := int64(0)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		if string(key) == keyStart {
			continue
		}

		m[string(key)] = string(value)

		count++
		if count >= limit {
			break
		}
	}
	iter.Release()

	err := iter.Error()

	if err == nil && count < limit {
		value, err := s.db.Get([]byte(keyEnd), nil)
		if err == nil {
			m[keyEnd] = string(value)
		}
	}

	return m, err
}

//rscan key_start key_end limit 列出处于区间 (key_start, key_end] 的 key-value 列表, 反向.
func (s *LevelDBStorage) RScan(keyStart, keyEnd string, limit int64) (map[string]string, error) {
	m := map[string]string{}

	srange := &util.Range{Start: []byte(keyEnd), Limit: []byte(keyStart)}
	iter := s.db.NewIterator(srange, nil)

	//leveldb的range 前闭后开，与ssdb相反
	count := int64(0)

	if iter.Last() {
		key := iter.Key()
		value := iter.Value()
		m[string(key)] = string(value)
		count++

		if limit > count {
			for iter.Prev() {
				key := iter.Key()
				value := iter.Value()

				m[string(key)] = string(value)
				count++
				if count >= limit {
					break
				}
			}
		}
	}
	iter.Release()
	err := iter.Error()

	return m, err
}

func (s *LevelDBStorage) Close() error {
	return s.db.Close()
}
