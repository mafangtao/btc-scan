package logic

import (
	"container/list"
	"github.com/liyue201/btc-scan/models"
	st "github.com/liyue201/btc-scan/storage"
	"github.com/liyue201/go-logger"
	"sort"
	"time"
)

type ConfirmedBlockMgr struct {
	storage st.Storage
	ctx     *models.Context
}

type TxAddr struct {
	addr string
	Type uint32
}

func NewConfirmedBlockMgr(storage st.Storage) (*ConfirmedBlockMgr, error) {
	m := ConfirmedBlockMgr{
		storage: storage,
	}
	ctx := &models.Context{}
	err := ctx.Load(storage, "")
	if err != err && err != models.ErrorNotFound {
		return nil, err
	}
	m.ctx = ctx

	return &m, nil
}

func (m *ConfirmedBlockMgr) GetContext() *models.Context {
	return m.ctx
}

//根据交易的的依赖关系排序, 其实就是一个有向无环图的拓扑排序
func (m *ConfirmedBlockMgr) SortTransactions(txs []*models.Transaction) {
	n := len(txs)
	txMap := make(map[string]int)
	edges := make(map[int]map[int]bool)

	//先按txid排序
	sort.Sort(models.TxByTxidSlice(txs))

	for i := 0; i < n; i++ {
		txMap[txs[i].TxId] = i
	}

	//构建DAG
	inDegree := make([]int, n)
	for i := 0; i < n; i++ {
		for _, in := range txs[i].Inputs {
			j, exist := txMap[in.TxId]
			if exist {
				//从i到j建立一条边
				e, exist := edges[i]
				if !exist {
					e = make(map[int]bool)
					edges[i] = e
				}
				if _, exist := e[j]; !exist {
					e[j] = true
					inDegree[j]++
				}
			}
		}
	}

	//拓扑排序
	queue := list.New()
	for i := 0; i < n; i++ {
		if inDegree[i] == 0 {
			queue.PushBack(i)
		}
	}
	newtxs := []*models.Transaction{}
	for queue.Len() > 0 {
		item := queue.Front()
		idx := item.Value.(int)
		queue.Remove(item)
		newtxs = append(newtxs, txs[idx])
		e, _ := edges[idx]
		for k, _ := range e {
			inDegree[k]--
			if inDegree[k] == 0 {
				queue.PushBack(k)
			}
		}
	}
	//逆序
	for i := 0; i < n; i++ {
		txs[i] = newtxs[n-1-i]
	}
}

func (m *ConfirmedBlockMgr) ProcessBlockTxs(txs []*models.Transaction) error {

	//根据交易顺序排序
	m.SortTransactions(txs)

	idx := 0
	totalTxs := len(txs)
	if totalTxs == 0 {
		return nil
	}

	//跳过已经处理完成的交易
	if txs[0].BlockHeight == m.ctx.BlockHeight && m.ctx.BlockProcessedTx < int32(totalTxs) {
		idx = int(m.ctx.BlockProcessedTx)
	}

	//继续处理剩下的交易
	for ; idx < totalTxs; idx++ {
		tx := txs[idx]
		//需要构建新的ctx，异常时不需要回滚
		tempCtx := m.ctx.Clone()
		err := m.ProcessBlockTx(tempCtx, tx, idx)
		if err != nil {
			logger.Errorf("ProcessBlockTxs ProcessBlockTx %s", err)
			return err
		}
		tempCtx.BlockHeight = tx.BlockHeight
		tempCtx.BlockHash = tx.BlockHash
		tempCtx.BlockProcessedTx = int32(idx + 1)
		tempCtx.BlockTotalTx = int32(len(txs))
		m.ctx = tempCtx
	}

	err := m.ctx.Save(m.storage, models.BlockCtxKey(m.ctx.BlockHeight))
	if err != nil {
		logger.Errorf("ProcessBlockTxs m.ctx.Save %s", err)
		return err
	}
	err = m.ctx.Save(m.storage, "")
	if err != nil {
		logger.Errorf("ProcessBlockTxs m.ctx.Save %s", err)
		return err
	}
	return nil
}

func (m *ConfirmedBlockMgr) GetTx(txid string) (*models.Transaction, error) {
	t := &models.Transaction{}
	err := t.Load(m.storage, txid)
	return t, err
}

func (m *ConfirmedBlockMgr) SaveTx(tx *models.Transaction) error {
	err := tx.Save(m.storage)
	return err
}

func (m *ConfirmedBlockMgr) ProcessBlockTx(ctx *models.Context, tx *models.Transaction, txIndex int) error {

	t0 := time.Now()

	ctx.TotalTx++
	tx.Id = ctx.TotalTx

	curTxAddr := map[string]*TxAddr{}

	for _, in := range tx.Inputs {
		preTx := &models.Transaction{}
		err := preTx.Load(m.storage, in.TxId)
		if err != nil {
			logger.Errorf("ProcessBlockTx preTx.Load %s:txid=%s", err, in.TxId)
			return err
		}

		for _, preOutput := range preTx.Outputs {
			if preOutput.Vout == in.Vout {
				in.Amount = preOutput.Amount
				in.Addr = preOutput.Addr

				strAddr := preOutput.Addr
				txAddr, exist := curTxAddr[strAddr]
				if !exist {
					txAddr = &TxAddr{
						addr: strAddr,
						Type: models.TxTypeInput,
					}
					curTxAddr[strAddr] = txAddr
				}

				//删除utxo
				err := models.DeleteUtxo(m.storage, txAddr.addr, preTx.Id, in.Vout)
				if err != nil {
					logger.Errorf("ProcessBlockTx models.DeleteUtxo %s", err)
				}
				break
			}
		}
	}

	t1 := time.Now()

	for _, output := range tx.Outputs {
		strAddr := output.Addr
		txAddr, exist := curTxAddr[strAddr]
		if !exist {
			txAddr = &TxAddr{
				addr: strAddr,
				Type: models.TxTypeOutput,
			}
			curTxAddr[strAddr] = txAddr
		} else {
			txAddr.Type |= models.TxTypeOutput
		}

		//增加utxo
		utxo := models.NewUtxo(tx.TxId, output.Vout, output.Amount, output.ScriptPubKey)
		err := utxo.Save(m.storage, models.UtxoKey(txAddr.addr, tx.Id, output.Vout))
		if err != nil {
			logger.Errorf("ProcessBlockTx utxo.Save %s", err)
			return err
		}
	}

	t2 := time.Now()

	err := tx.Save(m.storage)
	if err != nil {
		logger.Errorf("ProcessBlockTx tx.Save %s", err)
		return err
	}

	for _, txAddr := range curTxAddr {
		ctx.TotalTxNode++
		txNode := models.NewTxNode(ctx.TotalTxNode, tx.TxId, txAddr.Type)
		err := txNode.Save(m.storage, models.TxNodeKey(txAddr.addr, txNode.Id))
		if err != nil {
			logger.Errorf("ProcessBlockTx txNode.Save %s", err)
			return err
		}
	}

	t3 := time.Now()

	d1 := (t1.UnixNano() - t0.UnixNano()) / 1000000
	d2 := (t2.UnixNano() - t1.UnixNano()) / 1000000
	d3 := (t3.UnixNano() - t2.UnixNano()) / 1000000
	d := (t3.UnixNano() - t0.UnixNano()) / 1000000
	if d > 1000 {
		logger.Infof("ProcessBlockTx txid=%s, d=%d, d1=%d, d2=%d, d3=%d ", tx.TxId, d, d1, d2, d3)
	}
	return nil
}

func (m *ConfirmedBlockMgr) GetAddressTxs(addr string, limit uint, order int, prevMinKey, prevMaxKey string) ([]*models.Transaction, string, string, error) {

	retTxs := []*models.Transaction{}

	txnodes := []*models.TxNode{}
	minKey := ""
	MaxKey := ""
	var err error

	if order == -1 {
		//逆序查找
		start := ""
		if prevMinKey != "" {
			start = prevMinKey
		} else {
			start = models.TxNodeKeyEnd(addr)
		}
		end := models.TxNodeKeyStart(addr)
		txnodes, minKey, MaxKey, err = models.RScanTxNodes(m.storage, start, end, int64(limit))

		if err != nil {
			return retTxs, minKey, MaxKey, err
		}
	} else {
		//正序查找
		start := ""
		if prevMaxKey != "" {
			start = prevMaxKey
		} else {
			start = models.TxNodeKeyStart(addr)
		}
		end := models.TxNodeKeyEnd(addr)

		txnodes, minKey, MaxKey, err = models.ScanTxNodes(m.storage, start, end, int64(limit))

		if err != nil {
			return retTxs, minKey, MaxKey, err
		}
	}

	for _, txNode := range txnodes {
		tx, err := m.GetTx(txNode.TxId)
		if err != nil {
			logger.Errorf("GetAddressTxsWithOffset m.CFBlock.GetTx %s", err)
			break
		}
		retTxs = append(retTxs, tx)
	}

	return retTxs, minKey, MaxKey, nil
}

func (m *ConfirmedBlockMgr) GetAddrUtxo(addr string) ([]*models.Utxo, error) {

	retutxo, _ := models.ScanAddrUtxos(m.storage, models.UtxoKeyStart(addr), models.UtxoKeyEnd(addr))

	return retutxo, nil
}
