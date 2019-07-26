package logic

import (
	"fmt"
	"github.com/liyue201/go-logger"
	"sort"
	"github.com/liyue201/btc-scan/models"
	"github.com/liyue201/btc-scan/storage"
)

type TxGeter interface {
	GetTx(txid string) (*models.Transaction, error)
}

type TxMgr struct {
	UnCFBlock *UnconfirmedBlockMgr
	CFBlock   *ConfirmedBlockMgr
	Mempool   *MempoolMgr
}

func NewTxMgr(st storage.Storage, mempoolSize int) (*TxMgr, error) {
	m := &TxMgr{}
	cfBlock, err := NewConfirmedBlockMgr(st)
	if err != nil {
		return nil, err
	}
	m.CFBlock = cfBlock
	m.UnCFBlock = NewUnConfirmedBlockMgr(m)
	m.Mempool = NewMempool(m, mempoolSize)
	return m, nil
}

//TxGeter function
func (m *TxMgr) GetTx(txid string) (*models.Transaction, error) {
	tx, err := m.UnCFBlock.GetTx(txid)
	if err == nil {
		return tx, nil
	}

	tx, err = m.CFBlock.GetTx(txid)
	if err == nil {
		return tx, nil
	}

	tx, err = m.Mempool.GetTransactions(txid)
	if err == nil {
		return tx, nil
	}
	return nil, models.ErrorNotFound
}

func (m *TxMgr) GetAddressConfirmedTxs(address string, limit uint, order int, prevMinKey, prevMaxKey string) ([]*models.Transaction, string, string, error) {
	retTxs := []*models.Transaction{}

	txs, minKey, maxKey, _ := m.CFBlock.GetAddressTxs(address, limit, order, prevMinKey, prevMaxKey)
	if len(txs) > 0 {
		retTxs = append(retTxs, txs...)
	}
	if len(retTxs) > 0 {
		sort.Sort(models.TxByTimeDescSlice(retTxs))
	}
	return retTxs, minKey, maxKey, nil
}

func (m *TxMgr) GetAddressUnConfirmedTxs(address string) ([]*models.Transaction, error) {

	retTxs := []*models.Transaction{}

	//从内存池取
	mempoolTxs, _ := m.Mempool.GetAddrTxs(address)
	if len(mempoolTxs) > 0 {
		retTxs = append(retTxs, mempoolTxs...)
	}

	//从未确认区块取
	unconfirmedTxs, _ := m.UnCFBlock.GetAddrTxs(address)
	if len(unconfirmedTxs) > 0 {
		retTxs = append(retTxs, unconfirmedTxs...)
	}

	if len(retTxs) > 0 {
		sort.Sort(models.TxByTimeDescSlice(retTxs))
	}

	return retTxs, nil
}

func (m *TxMgr) GetAddressUtxos(address string) ([]*models.Utxo, error) {

	utxoMap := make(map[string]*models.Utxo)

	utxo, err := m.CFBlock.GetAddrUtxo(address)
	if err != nil {
		logger.Errorf("GetAddressUtxos %s", err)
	}
	for _, utxo := range utxo {
		key := fmt.Sprintf("%s-%d", utxo.TxId, utxo.Vout)
		utxoMap[key] = utxo
	}

	txs, _ := m.UnCFBlock.GetAddrTxs(address)
	mempoolTxs, _ := m.Mempool.GetAddrTxs(address)

	txs = append(txs, mempoolTxs...)

	for _, tx := range txs {
		for _, output := range tx.Outputs {
			if output.Addr == address {
				utxo := &models.Utxo{TxId: tx.TxId, Vout: output.Vout, Amount: output.Amount, ScriptPubKey: output.ScriptPubKey}
				key := fmt.Sprintf("%s-%d", utxo.TxId, utxo.Vout)
				utxoMap[key] = utxo
			}
		}
	}

	for _, tx := range txs {
		for _, input := range tx.Inputs {
			key := fmt.Sprintf("%s-%d", input.TxId, input.Vout)
			delete(utxoMap, key)
		}
	}

	retUtxos := []*models.Utxo{}
	for _, utxo := range utxoMap {
		retUtxos = append(retUtxos, utxo)
	}

	return retUtxos, nil
}

func  (m *TxMgr) GetCtx() *models.Context {
	return  m.CFBlock.ctx
}