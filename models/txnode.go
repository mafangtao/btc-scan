package models

import (
	"github.com/liyue201/btc-scan/storage"
)

func (m *TxNode) Encode() ([]byte, error) {
	return Encode(m)
}

func (m *TxNode) Decode(data []byte) error {
	return Decode(data, m)
}

func (m *TxNode) Save(store storage.Storage, key string) error {
	data, err := m.Encode()
	if err != nil {
		return err
	}
	return store.Set(key, string(data))
}

func (m *TxNode) Load(store storage.Storage, key string) error {
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

func NewTxNode(id int64, txid string, Type uint32) *TxNode {
	return &TxNode{
		Id:   id,
		TxId: txid,
		Type: Type,
	}
}

func ScanTxNodes(store storage.Storage, keyStart, keyEnd string, limit int64) ([]*TxNode, string, string, error) {
	txnodes := []*TxNode{}
	values, err := store.Scan(keyStart, keyEnd, limit)
	if err != nil {
		return txnodes,  keyStart, keyEnd , err
	}
	return HandelTxNodesResult(values)
}

func HandelTxNodesResult(values map[string]string) ([]*TxNode, string, string, error) {
	txnodes := []*TxNode{}
	minKey := ""
	maxKey := ""

	for k, v := range values {
		node := &TxNode{}
		err := node.Decode([]byte(v))
		if err != nil {
			return txnodes, minKey, maxKey, err
		}

		txnodes = append(txnodes, node)
		if minKey == "" {
			minKey = k
			maxKey = k
		}
		if minKey > k {
			minKey = k
		}
		if maxKey < k {
			maxKey = k
		}
	}
	return txnodes, minKey, maxKey, nil
}

func RScanTxNodes(store storage.Storage, keyStart, keyEnd string, limit int64)([]*TxNode, string, string, error) {
	txnodes := []*TxNode{}

	//logger.Debugf("start = %s, end = %s limt= %d", keyStart, keyEnd, limit)

	values, err := store.RScan(keyStart, keyEnd, limit)

	if err != nil {
		return txnodes,  keyStart, keyEnd , err
	}
	return HandelTxNodesResult(values)
}
