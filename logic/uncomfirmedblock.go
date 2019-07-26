package logic

import (
	"sort"
	"sync"
	"github.com/liyue201/btc-scan/models"
	"github.com/liyue201/go-logger"
)

type Block struct {
	Height int64
	Hash   string
	Txs    []*models.Transaction
}

type txPosition struct {
	BlockHeigh int64 //区块高度
	Index      int   //数组中的第几笔交易
}

//未确认的区块
type UnconfirmedBlockMgr struct {
	txGeter        TxGeter
	blocks         map[int64]*Block
	txIndex        map[string]*txPosition                    //交易索引，交易的位置，txid->txpoint
	txAddrs        map[string]map[string]struct{}            //交易涉及到的地址列表， txid->addr->struct
	addrIndex      map[string]map[string]*models.Transaction //地址索引，地址的所有交易， addr->txid->tx
	maxBlockHeight int64
	mutex          sync.RWMutex
}

func NewUnConfirmedBlockMgr(txGeter TxGeter) *UnconfirmedBlockMgr {
	return &UnconfirmedBlockMgr{
		txGeter:   txGeter,
		blocks:    make(map[int64]*Block),
		txIndex:   make(map[string]*txPosition),
		txAddrs:   make(map[string]map[string]struct{}),
		addrIndex: make(map[string]map[string]*models.Transaction),
	}
}

func (m *UnconfirmedBlockMgr) RemoveBlock(heigth int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.removeBlock(heigth)

	return nil
}

func (m *UnconfirmedBlockMgr) AddBlock(block *Block) error {
	m.mutex.Lock()
	tempBlock, exist := m.blocks[block.Height]
	if exist {
		if tempBlock.Hash == block.Hash {
			m.mutex.Unlock()
			return nil
		}
	}
	m.blocks[block.Height] = block
	m.mutex.Unlock()

	for i, tx := range block.Txs {

		m.mutex.Lock()
		m.txIndex[tx.TxId] = &txPosition{BlockHeigh: block.Height, Index: i}
		m.mutex.Unlock()

		txAddrs := make(map[string]struct{})
		for _, in := range tx.Inputs {
			preTx, err := m.txGeter.GetTx(in.TxId)
			if err != nil {
				logger.Errorf("[AddBlock]: can not find transaction: %v", in.TxId)
				continue
			}
			for _, preOutput := range preTx.Outputs {
				if preOutput.Vout == in.Vout {
					addr := preOutput.Addr
					txAddrs[addr] = struct{}{}
					in.Amount = preOutput.Amount
					in.Addr = preOutput.Addr
					break
				}
			}

		}
		for _, out := range tx.Outputs {
			txAddrs[out.Addr] = struct{}{}
		}

		m.mutex.Lock()

		if m.maxBlockHeight < block.Height {
			m.maxBlockHeight = block.Height
		}

		m.txAddrs[tx.TxId] = txAddrs

		for addr, _ := range txAddrs {
			addrTx, exist := m.addrIndex[addr]
			if !exist {
				addrTx = make(map[string]*models.Transaction)
				m.addrIndex[addr] = addrTx
			}
			addrTx[tx.TxId] = tx
		}
		m.mutex.Unlock()
	}

	return nil
}

func (m *UnconfirmedBlockMgr) GetBlock(heigth int64) (*Block, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	block, exist := m.blocks[heigth]
	if exist {
		return block, nil
	}
	return nil, models.ErrorNotFound
}

func (m *UnconfirmedBlockMgr) GetTx(txid string) (*models.Transaction, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	txPoint, exist := m.txIndex[txid]
	if !exist {
		return nil, models.ErrorNotFound
	}
	block, exist := m.blocks[txPoint.BlockHeigh]
	if !exist || len(block.Txs)-1 < txPoint.Index {
		return nil, models.ErrorNotFound
	}
	tx := block.Txs[txPoint.Index]

	return tx, nil
}

func (m *UnconfirmedBlockMgr) removeBlock(heigth int64) error {
	block, exist := m.blocks[heigth]
	if !exist {
		return nil
	}

	for _, tx := range block.Txs {
		addrs, exist := m.txAddrs[tx.TxId]
		if exist {
			for addr, _ := range addrs {
				addrtx, exist := m.addrIndex[addr]
				if exist {
					delete(addrtx, tx.TxId)
				}
			}
		}
	}

	for _, tx := range block.Txs {
		delete(m.txIndex, tx.TxId)
		delete(m.txAddrs, tx.TxId)
	}

	delete(m.blocks, heigth)

	return nil
}

func (m *UnconfirmedBlockMgr) GetAddrTxs(addr string) ([]*models.Transaction, error) {
	txs := []*models.Transaction{}

	m.mutex.RLock()
	addrTx, exist := m.addrIndex[addr]
	if exist {
		m.mutex.RUnlock()
		for _, tx := range addrTx {
			txs = append(txs, tx)
		}
	} else {
		m.mutex.RUnlock()
	}

	return txs, nil
}

func (m *UnconfirmedBlockMgr) GetAddrTxsWithOffset(addr string, offset, limit uint) ([]*models.Transaction, uint, error) {

	retTxs := []*models.Transaction{}
	txs := []*models.Transaction{}

	m.mutex.RLock()
	addrTx, exist := m.addrIndex[addr]
	if exist {
		m.mutex.RUnlock()
		for _, tx := range addrTx {
			txs = append(txs, tx)
		}
	} else {
		m.mutex.RUnlock()
	}

	if offset >= uint(len(txs)) {
		return retTxs, offset - uint(len(txs)), nil
	}

	sort.Sort(models.TxByTimeDescSlice(txs))

	if offset+limit < uint(len(txs)) {
		retTxs = append(retTxs, txs[offset:limit+offset]...)
		offset = 0
	} else {
		retTxs = append(retTxs, txs[offset:]...)
		offset = offset - uint(len(txs))
	}

	return retTxs, offset, nil
}

func (m *UnconfirmedBlockMgr) GetMaxBlockHeight() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.maxBlockHeight
}
