package models

import "github.com/liyue201/btc-scan/storage"

func (m *Transaction) Encode() ([]byte, error) {
	return Encode(m)
}

func (m *Transaction) Decode(data []byte) error {
	return Decode(data, m)
}

func (m *Transaction) Save(store storage.Storage) error {
	data, err := m.Encode()
	if err != nil {
		return err
	}
	return store.Set(TxKey(m.TxId), string(data))
}

func (m *Transaction) Load(store storage.Storage, txid string) error {
	data, err := store.Get(TxKey(txid))
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

type TxByTxidSlice []*Transaction

func (c TxByTxidSlice) Len() int {
	return len(c)
}
func (c TxByTxidSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c TxByTxidSlice) Less(i, j int) bool {
	return c[i].TxId < c[j].TxId
}

type TxByTimeSlice []*Transaction

func (c TxByTimeSlice) Len() int {
	return len(c)
}
func (c TxByTimeSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c TxByTimeSlice) Less(i, j int) bool {
	return c[i].BlockTime < c[j].BlockTime
}

type TxByTimeDescSlice []*Transaction

func (c TxByTimeDescSlice) Len() int {
	return len(c)
}
func (c TxByTimeDescSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c TxByTimeDescSlice) Less(i, j int) bool {
	return c[i].BlockTime > c[j].BlockTime
}
