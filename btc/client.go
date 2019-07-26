package btc

import (
	"bytes"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	_ "github.com/btcsuite/btcutil"
	log "github.com/liyue201/go-logger"
	"io/ioutil"
	"github.com/liyue201/btc-scan/config"
	"github.com/liyue201/btc-scan/uitils"
	"encoding/hex"
)

//测试发现rpcclient的websocket接口存在bug，调接口时经常卡死
//所以正常流程使用https，回调使用websocket
type Client struct {
	RpcClient     *rpcclient.Client
	WsClient      *rpcclient.Client
	MempoolTxChan chan *btcjson.TxRawResult
	MempoolOpened bool
}

func NewClient(nodeCfg *config.BtcNodeCfg) (*Client, error) {

	rpcConfig := &rpcclient.ConnConfig{
		Host:         nodeCfg.Address,
		User:         nodeCfg.User,
		Pass:         nodeCfg.Password,
		Endpoint:     "https",
		Certificates: getCertificate(nodeCfg.Certificate),
		HTTPPostMode: true,  // Bitcoin core only supports HTTP POST mode
		DisableTLS:   false, // Bitcoin core does not provide TLS by default
	}

	wsConfig := &rpcclient.ConnConfig{
		Host:         nodeCfg.Address,
		User:         nodeCfg.User,
		Pass:         nodeCfg.Password,
		Endpoint:     "ws",
		Certificates: getCertificate(nodeCfg.Certificate),
		HTTPPostMode: false, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   false, // Bitcoin core does not provide TLS by default
	}

	c := &Client{
		MempoolTxChan: make(chan *btcjson.TxRawResult, 5),
	}
	if nodeCfg.Mempool > 1 {
		c.MempoolOpened = true
	}

	ntfnHandlers := rpcclient.NotificationHandlers{
		OnTxAcceptedVerbose: func(rawtx *btcjson.TxRawResult) {
			//log.Debugf("OnTxAcceptedVerbose txid = %s", rawtx.Txid)
			c.MempoolTxChan <- rawtx
		},
	}

	rpcClient, err := rpcclient.New(rpcConfig, nil)
	if err != nil {
		log.Errorf("NewClient(): rpcclient.New %s\n", err.Error())
		return nil, err
	}

	wsClient, err := rpcclient.New(wsConfig, &ntfnHandlers)
	if err != nil {
		log.Errorf("NewClient(): rpcclient.New %s\n", err.Error())
		return nil, err
	}

	log.Info("NewClient rpcclient.New ok")

	c.RpcClient = rpcClient
	c.WsClient = wsClient

	return c, nil
}

func (c *Client) Run() error {
	log.Info("Run")

	if c.MempoolOpened {
		if err := c.WsClient.NotifyNewTransactions(true); err != nil {
			log.Errorf("Run(): WsClient.NotifyNewTransactions %s\n", err.Error())
			return err
		}
	}

	c.RpcClient.WaitForShutdown()
	return nil
}

func (c *Client) Stop() {
	c.RpcClient.Shutdown()
	c.WsClient.Shutdown()
}

func getCertificate(certFile string) []byte {
	certFile = utils.AbsPath(certFile)
	cert, err := ioutil.ReadFile(certFile)
	cert = bytes.Trim(cert, "\x00")

	if err != nil {
		log.Errorf("get certificate: %s", err.Error())
		return []byte{}
	}
	if len(cert) > 1 {
		return cert
	}
	log.Errorf("get certificate: empty certificate")
	return []byte{}
}

func (c *Client) SendTx(rawtx string) (string, error) {
	var tx wire.MsgTx

	data, err := hex.DecodeString(rawtx)
	if err != nil {
		log.Errorf("SendTx DecodeString %s", err)
		return "", err
	}
	rbuf := bytes.NewReader([]byte(data))

	err = tx.Deserialize(rbuf)
	if err != nil {
		log.Errorf("SendTx tx.DeserializeNoWitness %s", err)
		return "", err
	}
	txid, err := c.RpcClient.SendRawTransaction(&tx, true)
	if err != nil {
		log.Errorf("SendTx c.RpcClient.SendRawTransaction %s", err)
		return "", err
	}
	return txid.String(), err
}
