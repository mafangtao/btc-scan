package rest

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/liyue201/go-logger"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/liyue201/btc-scan/btc"
	"github.com/liyue201/btc-scan/logic"
	"github.com/liyue201/btc-scan/models"
	"io/ioutil"
)

type RestSever struct {
	httpServer *http.Server
	txmgr      *logic.TxMgr
	btcCli     *btc.Client
}

func NewHttpServer(port int, txmgr *logic.TxMgr, btcClient *btc.Client) *RestSever {
	gin.SetMode(gin.ReleaseMode)
	engin := gin.Default()

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	httpServer := &http.Server{Addr: addr, Handler: engin}
	server := &RestSever{
		httpServer: httpServer,
		txmgr:      txmgr,
		btcCli:     btcClient,
	}
	server.initRoute(engin)

	return server
}

func (s *RestSever) Run() {
	err := s.httpServer.ListenAndServe()
	if err != nil {
		logger.Errorf("RestSever.Run %s", err)
	}
}

func (s *RestSever) Stop() {
	s.httpServer.Shutdown(context.Background())
}

func (s *RestSever) initRoute(r gin.IRouter) {
	r.GET("/api/v1/txs", s.GetAddressTransations)
	r.GET("/api/v1/uctxs", s.GetAddressUnconfirmedTransations)
	r.GET("/api/v1/addr/:addr/utxo", s.GetAddressUtxo)
	r.GET("/api/v1/best_height", s.GetBestHeight)
	r.POST("/api/v1/tx/send", s.SendTx)

	//查看同步情况的接口
	r.GET("/test/ctx", s.GetCtx)
}

func (s *RestSever) GetConfirmedAddressTransations(c *gin.Context) {
	addr := c.Query("address")

	strLimit := c.Query("limit")
	if strLimit == "" {
		strLimit = "10"
	}
	limit, _ := strconv.Atoi(strLimit)

	strOrder := c.Query("order")
	if strOrder == "" {
		strOrder = "-1"
	}
	order, _ := strconv.Atoi(strOrder)

	prevMinKey := c.Query("prevminkey")
	prevMaxKey := c.Query("prevmaxkey")

	logger.Debugf("addr:%s, limit:%d, order:%d, prevminkey:%s, prevmaxkey:%s", addr, limit, order, prevMinKey, prevMaxKey)

	if addr == "" {
		RespJson(c, BadRequest, nil)
		return
	}

	txs, minKey, maxKey, err := s.txmgr.GetAddressConfirmedTxs(addr, uint(limit), order, prevMinKey, prevMaxKey)
	if err != nil {
		logger.Errorf("GetAddressTransations %s", err)
	}
	ret := struct {
		Txs    interface{} `json:"txs"`
		MinKey string      `json:"minkey"`
		MaxKey string      `json:"maxkey"`
	}{
		Txs:    txs,
		MinKey: minKey,
		MaxKey: maxKey,
	}

	RespJson(c, OK, ret)
}

func (s *RestSever) GetAddressTransations(c *gin.Context) {
	addr := c.Query("address")

	strLimit := c.Query("limit")
	if strLimit == "" {
		strLimit = "10"
	}
	limit, _ := strconv.Atoi(strLimit)

	strOrder := c.Query("order")
	if strOrder == "" {
		strOrder = "-1"
	}
	order, _ := strconv.Atoi(strOrder)

	prevMinKey := c.Query("prevminkey")
	prevMaxKey := c.Query("prevmaxkey")

	logger.Debugf("addr:%s, limit:%d, order:%d, prevminkey:%s, prevmaxkey:%s", addr, limit, order, prevMinKey, prevMaxKey)

	if addr == "" {
		RespJson(c, BadRequest, nil)
		return
	}

	makeUcTxkey := func(blockTime int64) string {
		return fmt.Sprintf("uctx-%d", blockTime)
	}

	parserUcTxKey := func(key string) int64 {
		blockTime := int64(0)
		fmt.Sscanf(key, "uctx-%d", &blockTime)
		return blockTime
	}

	isUcTxKey := func(key string) bool {
		return strings.HasPrefix(key, "uctx-")
	}

	var retTxs []*models.Transaction

	ret := struct {
		Txs    interface{} `json:"txs"`
		MinKey string      `json:"minkey"`
		MaxKey string      `json:"maxkey"`
	}{}

	if order == -1 {
		//往前翻页
		if isUcTxKey(prevMinKey) || prevMinKey == "" {
			txs, err := s.txmgr.GetAddressUnConfirmedTxs(addr)
			if err != nil {
				logger.Errorf("GetAddressTransations %s", err)
			}
			blockTime := time.Now().Unix()
			if prevMinKey != "" {
				blockTime = parserUcTxKey(prevMinKey)
			}

			//txs 按时间降序
			txLen := len(txs)
			for i := 0; i < txLen; i++ {
				tx := txs[i]
				if tx.BlockTime < blockTime {
					retTxs = append(retTxs, tx)
				}
				if len(retTxs) == 1 {
					ret.MaxKey = makeUcTxkey(retTxs[0].BlockTime)
				}
				if len(retTxs) >= limit {
					ret.MinKey = makeUcTxkey(retTxs[len(retTxs)-1].BlockTime)
					break
				}
			}
		}
		if len(retTxs) < limit {
			txs, minKey, maxKey, err := s.txmgr.GetAddressConfirmedTxs(addr, uint(limit-len(retTxs)), order, prevMinKey, prevMaxKey)
			if err != nil {
				logger.Errorf("GetAddressTransations %s", err)
			}
			if len(txs) > 0 {
				if len(retTxs) == 0 {
					ret.MaxKey = maxKey
				}
				retTxs = append(retTxs, txs...)
				ret.MinKey = minKey
			}
		}
	} else {
		//往后翻页
		if isUcTxKey(prevMaxKey) {
			txs, err := s.txmgr.GetAddressUnConfirmedTxs(addr)
			if err != nil {
				logger.Errorf("GetAddressTransations %s", err)
			}
			blockTime := parserUcTxKey(prevMaxKey)

			//txs 按时间降序
			txLen := len(txs)
			for i := txLen - 1; i >= 0; i-- {
				tx := txs[i]
				if tx.BlockTime > blockTime {
					retTxs = append(retTxs, tx)
				}
				if len(retTxs) >= limit {
					break
				}
			}
			//逆序
			n := len(retTxs)
			for i := 0; i < n/2; i++ {
				retTxs[i], retTxs[n-i-1] = retTxs[n-i-1], retTxs[i]
			}
			if n > 0 {
				ret.MaxKey = makeUcTxkey(retTxs[0].BlockTime)
				ret.MinKey = makeUcTxkey(retTxs[n-1].BlockTime)
			}
		} else {
			txs, minKey, maxKey, err := s.txmgr.GetAddressConfirmedTxs(addr, uint(limit), order, prevMinKey, prevMaxKey)
			if err != nil {
				logger.Errorf("GetAddressTransations %s", err)
			}
			if len(txs) > 0 {
				ret.MinKey = minKey
				ret.MaxKey = maxKey
				retTxs = append(retTxs, txs...)
			}
			if len(retTxs) < limit {
				txs, err := s.txmgr.GetAddressUnConfirmedTxs(addr)
				if err != nil {
					logger.Errorf("GetAddressTransations %s", err)
				}
				n := len(txs)
				for i := n - 1; i >= 0; i-- {
					retTxs = append(retTxs, txs[i])
					if len(retTxs) >= limit {
						ret.MaxKey = makeUcTxkey(txs[i].BlockTime)
						break
					}
				}
			}
		}
	}
	ret.Txs = retTxs

	RespJson(c, OK, ret)
}

func (s *RestSever) GetAddressUnconfirmedTransations(c *gin.Context) {
	addr := c.Query("address")

	logger.Debugf("addr:%s", addr)

	if addr == "" {
		RespJson(c, BadRequest, nil)
		return
	}

	txs, err := s.txmgr.GetAddressUnConfirmedTxs(addr)
	if err != nil {
		logger.Errorf("GetAddressTransations %s", err)
	}

	ret := struct {
		Txs interface{} `json:"txs"`
	}{
		Txs: txs,
	}
	RespJson(c, OK, ret)
}

func (s *RestSever) GetAddressUtxo(c *gin.Context) {
	addr := c.Param("addr")

	logger.Debugf("addr:%s", addr)

	utxos, _ := s.txmgr.GetAddressUtxos(addr)

	ret := struct {
		Utxos interface{} `json:"utxos"`
	}{
		Utxos: utxos,
	}

	RespJson(c, OK, ret)
}

func (s *RestSever) SendTx(c *gin.Context) {
	req := struct {
		Rawtx string `json:"rawtx" binding:"required"`
	}{}

	//if err := c.ShouldBindJSON(&req); err != nil {
	//	logger.Errorf("SendTx %s", err)
	//	RespJson(c, BadRequest, nil)
	//	return
	//}

	body, _ := ioutil.ReadAll(c.Request.Body)
	logger.Debugf("SendTx: body=%s", string(body))

	ss := strings.Split(string(body), "=")
	if len(ss) > 1 {
		req.Rawtx = ss[1]
	}

	if req.Rawtx == "" {
		logger.Errorf("SendTx: rawtx is empty")
		RespJson(c, BadRequest, nil)
		return
	}

	txid, err := s.btcCli.SendTx(req.Rawtx)
	if err != nil {
		RespJson(c, InternalServerError, nil)
		return
	}
	ret := struct {
		Txid string `json:"txid"`
	}{
		Txid: txid,
	}
	RespJson(c, OK, ret)
}

func (s *RestSever) GetCtx(c *gin.Context) {
	logger.Debugf("GetCtx")

	ctx := s.txmgr.GetCtx()
	RespJson(c, OK, ctx)
}

func (s *RestSever) GetBestHeight(c *gin.Context) {
	logger.Debugf("GetBestHeight")

	height, err := s.btcCli.WsClient.GetBlockCount()
	if err != nil {
		logger.Errorf("GetBestHeight %v", err.Error())
		RespJson(c, InternalServerError, nil)
		return
	}

	ret := struct {
		BestHeight int64 `json:"bestHeight"`
	}{
		BestHeight: height,
	}

	RespJson(c, OK, ret)
}
