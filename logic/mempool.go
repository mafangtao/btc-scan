package logic

import (
	"container/list"
	"sort"
	"sync"
	"github.com/liyue201/btc-scan/models"
)

type MempoolMgr struct {
	txQueue *list.List //内存池中的交易
	txGeter TxGeter
	txs     map[string]*list.Element                  //txid->tx
	txAddrs map[string]map[string]struct{}            //交易涉及到的地址列表， txid->{addr->struct}
	addrTxs map[string]map[string]*models.Transaction //地址索引，地址的所有交易， addr->{txid->tx}
	mutex   sync.RWMutex

	mempoolSize int
}

func NewMempool(txGeter TxGeter, mempoolSize int) *MempoolMgr {
	m := &MempoolMgr{txGeter: txGeter,
		txQueue:     list.New(),
		txs:         make(map[string]*list.Element),
		txAddrs:     make(map[string]map[string]struct{}),
		addrTxs:     make(map[string]map[string]*models.Transaction),
		mempoolSize: mempoolSize,
	}
	return m
}

func (m *MempoolMgr) AddTransaction(tx *models.Transaction) error {
	m.mutex.Lock()
	if _, exist := m.txs[tx.TxId]; exist {
		m.mutex.Unlock()
		return nil
	}
	e := m.txQueue.PushBack(tx)
	m.txs[tx.TxId] = e
	m.mutex.Unlock()

	txAddrs := make(map[string]struct{})
	for _, in := range tx.Inputs {
		preTx, err := m.txGeter.GetTx(in.TxId)
		if err != nil {
			//logger.Errorf("[AddBlock]: can not find transaction: %v", in.TxId)
			continue
		}
		for _, preOutput := range preTx.Outputs {
			if preOutput.Vout == in.Vout {
				addr := preOutput.Addr
				txAddrs[addr] = struct{}{}
				in.Addr = addr
				in.Amount = preOutput.Amount
				break
			}
		}
	}
	for _, out := range tx.Outputs {
		txAddrs[out.Addr] = struct{}{}
	}

	m.mutex.Lock()
	m.txAddrs[tx.TxId] = txAddrs
	for addr, _ := range txAddrs {
		addrTx, exist := m.addrTxs[addr]
		if !exist {
			addrTx = make(map[string]*models.Transaction)
			m.addrTxs[addr] = addrTx
		}
		addrTx[tx.TxId] = tx
	}
	m.mutex.Unlock()

	m.RemoveTxIfOutOfLimit()

	return nil
}

func (m *MempoolMgr) GetTransactions(txid string) (*models.Transaction, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	e, exist := m.txs[txid]
	if !exist {
		return nil, models.ErrorNotFound
	}
	return e.Value.(*models.Transaction), nil
}

func (m *MempoolMgr) RemoveTxIfOutOfLimit() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.txQueue.Len() > m.mempoolSize {
		e := m.txQueue.Front()
		tx := e.Value.(*models.Transaction)
		m.removeTx(tx.TxId)
	}
}

func (m *MempoolMgr) RemoveTx(txid string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.removeTx(txid)
}

func (m *MempoolMgr) removeTx(txid string) error {
	e, exist := m.txs[txid]
	if !exist {
		return nil
	}
	addrs, exist := m.txAddrs[txid]
	if exist {
		for addr, _ := range addrs {
			addrtx, exist := m.addrTxs[addr]
			if exist {
				delete(addrtx, txid)
			}
		}
		delete(m.txAddrs, txid)
	}
	delete(m.txs, txid)
	m.txQueue.Remove(e)

	return nil
}

func (m *MempoolMgr) GetAddrTxs(addr string) ([]*models.Transaction, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	mempoolTxs := []*models.Transaction{}
	txMap, exist := m.addrTxs[addr]
	if !exist {
		return mempoolTxs, nil
	}
	for _, tx := range txMap {
		mempoolTxs = append(mempoolTxs, tx)
	}
	return mempoolTxs, nil
}

func (m *MempoolMgr) GetAddrTxsWithOffset(addr string, offset, limit uint) ([]*models.Transaction, uint, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	retTxs := []*models.Transaction{}
	mempoolTxs := []*models.Transaction{}

	txMap, exist := m.addrTxs[addr]
	if !exist {
		return retTxs, offset, nil
	}

	lenm := uint(len(txMap))
	if offset >= lenm {
		return retTxs, offset - lenm, nil
	}

	for _, tx := range txMap {
		mempoolTxs = append(mempoolTxs, tx)
	}

	sort.Sort(models.TxByTimeDescSlice(mempoolTxs))

	if offset+limit < uint(len(mempoolTxs)) {
		retTxs = append(retTxs, mempoolTxs[offset:limit+offset]...)
		offset = 0
	} else {
		retTxs = append(retTxs, mempoolTxs[offset:]...)
		offset = offset - uint(len(mempoolTxs))
	}

	return retTxs, offset, nil
}
