package storage


type Storage interface {
	Set(key string, value string) error

	Get(key string) (string, error)

	Delete(key string) error

	//scan key_start key_end limit  列出处于区间 (key_start, key_end] 的 key-value 列表.
	Scan(keyStart, keyEnd string, limit int64) (map[string]string, error)

	//rscan key_start key_end limit 列出处于区间 (key_start, key_end] 的 key-value 列表, 反向.
	RScan(keyStart, keyEnd string, limit int64) (map[string]string, error)

	Close() error
}

