package models

import (
	"github.com/liyue201/btc-scan/storage"
)

///key:  utxo-{addr}-{Transaction.id}-{Transaction.vout}

func (m *Utxo) Encode() ([]byte, error) {
	return Encode(m)
}

func (m *Utxo) Decode(data []byte) error {
	return Decode(data, m)
}

func (m *Utxo) Save(store storage.Storage, key string) error {
	data, err := m.Encode()
	if err != nil {
		return err
	}
	return store.Set(key, string(data))
}

func (m *Utxo) Load(store storage.Storage, key string) error {
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

func NewUtxo(txid string, vout uint32, amount int64, scriptPubKey string) *Utxo {
	return &Utxo{
		TxId:         txid,
		Vout:         vout,
		Amount:       amount,
		ScriptPubKey: scriptPubKey,
	}
}

func DeleteUtxo(store storage.Storage, addr string, txid int64, vout uint32) error {
	return store.Delete(UtxoKey(addr, txid, vout))
}

func ScanAddrUtxos(store storage.Storage, keyStart, keyEnd string) ([]*Utxo, error) {
	utxos := []*Utxo{}

	count := 0
	for {
		values, err := store.Scan(keyStart, keyEnd, 1000)
		if err != nil {
			return utxos, err
		}

		if len(values) == 0 {
			break
		}

		for k, v := range values {
			node := &Utxo{}
			err = node.Decode([]byte(v))

			if err != nil {
				return utxos, err
			}
			utxos = append(utxos, node)

			if k > keyStart {
				keyStart = k
			}
		}

		//超出100W个utxo结束，防止死循环
		count += len(values)
		if count > 1000000 {
			break
		}
	}

	return utxos, nil
}
