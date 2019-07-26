package btc

import (
	"context"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/liyue201/btc-scan/logic"
	"github.com/liyue201/btc-scan/models"
	"github.com/liyue201/btc-scan/uitils"
	"github.com/liyue201/go-logger"
	"time"
)

const (
	confirmationBlockNum = 6 //6区块确认
)

var SatoshiToBitcoin = float64(100000000)

type Worker struct {
	cli     *Client
	txmgr   *logic.TxMgr
	blockCh chan *logic.Block
	cancel  context.CancelFunc
}

func NewWorker(cli *Client, txmgr *logic.TxMgr) *Worker {
	return &Worker{cli: cli,
		txmgr:   txmgr,
		blockCh: make(chan *logic.Block, 10),
	}
}

func (w *Worker) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	wg := utils.WaitGroupWrapper{}

	wg.Wrap(func() {
		w.FetchBlock(ctx)
	})
	wg.Wrap(func() {
		w.ProcessBlock(ctx)
	})
	wg.Wrap(func() {
		w.ProcessMempool(ctx)
	})
	wg.Wait()
}

func (w *Worker) isDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
	return false
}

func (w *Worker) Stop() {
	w.cancel()
}

func (w *Worker) FetchBlock(ctx context.Context) {

	const TimeInterval = time.Second * 5
	const ErrorTimeInterval = time.Second * 10 //出现错误后的暂停时间

	context := w.txmgr.CFBlock.GetContext()

	confirmedHeight := int64(0)
	curHeight := int64(0)
	if context.BlockTotalTx == context.BlockProcessedTx && context.BlockTotalTx > 0 {
		//上个区块已经处理完成
		confirmedHeight = context.BlockHeight
		curHeight = context.BlockHeight + 1
	} else {
		//上个区块只处理部分交易
		confirmedHeight = context.BlockHeight
		curHeight = context.BlockHeight
	}

	logger.Infof("curHeight:%d", curHeight)

	hasErr := false
	for {
		if w.isDone(ctx) {
			logger.Info("FetchBlock stopped")
			return
		}
		//	time.Sleep(time.Second * 10)

		if hasErr {
			time.Sleep(ErrorTimeInterval)
			hasErr = false
		}

		bestHeight, err := w.cli.RpcClient.GetBlockCount()
		if err != nil {
			logger.Errorf("FetchBlock w.cli.RpcClient.GetBlockCount %s", err)
			hasErr = true
			continue
		}

		if confirmedHeight < curHeight && confirmedHeight+1+confirmationBlockNum-1 <= bestHeight {
			curHeight = confirmedHeight + 1
		}

		if curHeight >= bestHeight {
			time.Sleep(TimeInterval)
			continue
		}

		if curHeight+confirmationBlockNum-1 <= bestHeight {
			//已确认的区块
			hash, err := w.cli.RpcClient.GetBlockHash(curHeight)
			if err != nil {
				logger.Errorf("FetchBlock GetBlockHash %s", err.Error())
				hasErr = true
				continue
			}
			block, err := w.txmgr.UnCFBlock.GetBlock(curHeight)
			if err == nil {
				if block.Txs[0].BlockHash != hash.String() {
					for height := curHeight; height <= w.txmgr.UnCFBlock.GetMaxBlockHeight(); height++ {
						w.txmgr.UnCFBlock.RemoveBlock(height)
					}
					block, err = w.GetBlock(curHeight)
					if err != nil {
						logger.Errorf("FetchBlock GetBlock %s", err.Error())
						hasErr = true
						continue
					}
				}
			} else {
				block, err = w.GetBlock(curHeight)
				if err != nil {
					logger.Errorf("FetchBlock GetBlock %s", err.Error())
					hasErr = true
					continue
				}
			}

			logger.Infof("FetchBlock height=%d", curHeight)
			confirmedHeight = curHeight
			w.removeBlockTxsFromMempool(block.Txs)
			w.blockCh <- block
		} else {
			//未确认的区块
			block, err := w.GetBlock(curHeight)
			if err != nil {
				logger.Errorf("FetchBlock GetBlock %s", err.Error())
				hasErr = true
				continue
			}

			err = w.txmgr.UnCFBlock.AddBlock(block)
			if err != nil {
				logger.Errorf("FetchBlock w.txmgr.UnCFBlock.AddBlock %s", err.Error())
				hasErr = true
				continue
			}
			w.removeBlockTxsFromMempool(block.Txs)

			logger.Infof("FetchBlock height=%d", curHeight)
		}
		curHeight++
	}
}

func (w *Worker) ProcessBlock(ctx context.Context) {

	const ErrorTimeInterval = time.Second * 10 //出现错误后的暂停时间
	defer func() {
		//清空队列
		for {
			select {
			case <-w.blockCh:
			default:
				return
			}
		}
	}()

	for {
		if w.isDone(ctx) {
			logger.Info("ProcessBlock stopped")
			return
		}
		select {
		case block := <-w.blockCh:
			logger.Infof("ProcessBlock height=%d, transactions=%d", block.Height, len(block.Txs))

			hasErr := false
			errCount := 0
			for {
				if hasErr {
					time.Sleep(ErrorTimeInterval)
					hasErr = false
				}
				err := w.txmgr.CFBlock.ProcessBlockTxs(block.Txs)
				if err != nil {
					logger.Errorf("ProcessBlock w.txmgr.ProcessConfirmedBlockTxs %s", err)

					if w.isDone(ctx) {
						return
					}

					hasErr = true
					errCount++
					if errCount > 10 {
						logger.Errorf("ProcessBlock stopped since to many error")
						return
					}
					continue
				}
				break
			}
		default:
			time.Sleep(time.Millisecond * 50)
		}
	}
}

func (w *Worker) ProcessMempool(ctx context.Context) {

	//GetRawTransactionVerbose 需要开启btcd的txindex，导致同步数据很慢，暂时关闭

	//txHash, err := w.cli.RpcClient.GetRawMempool()
	//if err != nil {
	//	logger.Errorf("ProcessMempool w.cli.RpcClient.GetRawMempool %s", err.Error())
	//}
	//
	//for _, h := range txHash {
	//	rawtx, err := w.cli.RpcClient.GetRawTransactionVerbose(h)
	//	if err != nil {
	//		logger.Errorf("ProcessMempool c.RpcClient.GetRawTransactionVerbose %s", err.Error())
	//		continue
	//	}
	//	tx := w.RawTxToDbTx(rawtx)
	//	tx.BlockTime = time.Now().Unix()
	//	w.txmgr.Mempool.AddTransaction(tx)
	//}

	const ErrorTimeInterval = time.Second * 10 //出现错误后的暂停时间
	hasErr := false
	for {
		if w.isDone(ctx) {
			return
		}

		if hasErr {
			time.Sleep(ErrorTimeInterval)
			hasErr = false
		}
		select {
		case rawtx := <-w.cli.MempoolTxChan:
			tx := w.RawTxToDbTx(rawtx)
			tx.BlockTime = time.Now().Unix()
			w.txmgr.Mempool.AddTransaction(tx)
		default:
			time.Sleep(time.Millisecond * 50)
		}
	}
}

func (w *Worker) removeBlockTxsFromMempool(txs []*models.Transaction) {
	go func() {
		for _, tx := range txs {
			w.txmgr.Mempool.RemoveTx(tx.TxId)
		}
	}()
}

func (w *Worker) GetBlock(height int64) (*logic.Block, error) {
	block := &logic.Block{
		Height: height,
		Txs:    make([]*models.Transaction, 0),
	}

	hash, err := w.cli.RpcClient.GetBlockHash(height)
	if err != nil {
		logger.Errorf("GetBlock GetBlockHash %s", err.Error())
		return block, err
	}
	block.Hash = hash.String()

	rawBlock, err := w.cli.RpcClient.GetBlockVerboseTx(hash)
	if err != nil {
		logger.Errorf("GetBlock GetBlockVerboseTx %s", err.Error())
		return block, err
	}

	for _, rawTx := range rawBlock.RawTx {
		tx := w.RawTxToDbTx(&rawTx)
		tx.BlockHeight = height

		block.Txs = append(block.Txs, tx)
	}
	return block, nil
}

func (w *Worker) RawTxToDbTx(rawtx *btcjson.TxRawResult) *models.Transaction {
	tx := models.Transaction{}
	tx.TxId = rawtx.Txid
	tx.BlockTime = rawtx.Blocktime
	tx.BlockHash = rawtx.BlockHash
	for _, in := range rawtx.Vin {
		if in.IsCoinBase() {
			break
		}
		tx.Inputs = append(tx.Inputs, &models.TxInput{TxId: in.Txid, Vout: in.Vout})
	}
	for _, out := range rawtx.Vout {
		for _, addr := range out.ScriptPubKey.Addresses {
			amount := int64(out.Value * SatoshiToBitcoin)
			tx.Outputs = append(tx.Outputs, &models.TxOutput{Addr: addr, Amount: amount, Vout: out.N, ScriptPubKey: out.ScriptPubKey.Hex})
		}
	}
	return &tx
}
