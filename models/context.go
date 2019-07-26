package models

import (
	"github.com/liyue201/btc-scan/storage"
)

func (m *Context) Encode() ([]byte, error) {
	return Encode(m)
}

func (m *Context) Decode(data []byte) error {
	return Decode(data, m)
}

func (m *Context) Save(store storage.Storage, key string) error {
	data, err := m.Encode()
	if err != nil {
		return err
	}
	if key == "" {
		key = ContextKey
	}
	return store.Set(key, string(data))
}

func (m *Context) Load(store storage.Storage, key string) error {
	if key == "" {
		key = ContextKey
	}
	data, err := store.Get(key)
	if err != nil {
		return err
	}
	if data == "" {
		return ErrorNotFound
	}
	err = m.Decode([]byte(data))
	if err != nil {
		return err
	}
	return nil
}

func (m *Context) Clone() *Context {
	newCtx := &Context{
		BlockHeight:      m.BlockHeight,
		BlockHash:        m.BlockHash,
		BlockTotalTx:     m.BlockTotalTx,
		BlockProcessedTx: m.BlockProcessedTx,
		TotalTx:          m.TotalTx,
		TotalTxNode:      m.TotalTxNode,
	}
	return newCtx
}
