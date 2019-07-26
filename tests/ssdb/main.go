package main

import (
	"fmt"
	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
	"github.com/liyue201/btc-scan/models"
	"time"
	"github.com/liyue201/go-logger"
)

func NewSSDBStorage(config *conf.Config) (*SSDBStorage, error) {
	s := SSDBStorage{}
	p, err := gossdb.NewPool(config)
	if err != nil {
		return nil, err
	}
	s.pool = p
	return &s, nil
}

type SSDBStorage struct {
	pool *gossdb.Connectors
}

func (s *SSDBStorage) Set(key string, value string) error {
	c, err := s.pool.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()
	err = c.Set(key, value)
	return err
}

func (s *SSDBStorage) Get(key string) (string, error) {
	c, err := s.pool.NewClient()
	if err != nil {
		return "", nil
	}
	defer c.Close()

	value, err := c.Get(key)
	if err != nil {
		return "", err
	}
	return value.String(), nil
}

func (s *SSDBStorage) Delete(key string) error {
	c, err := s.pool.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()
	err = c.Del(key)
	return err
}

func (s *SSDBStorage) Scan(keyStart, keyEnd string, limit int64) (map[string]string, error) {
	m := map[string]string{}

	c, err := s.pool.NewClient()
	if err != nil {
		return m, nil
	}
	defer c.Close()

	values, err := c.Scan(keyStart, keyEnd, limit)
	if err != nil {
		return m, err
	}
	for key, v := range values{
		m[key] = v.String()
	}
	return m, nil
}

func (s *SSDBStorage) RScan(keyStart, keyEnd string, limit int64) (map[string]string, error) {
	m := map[string]string{}

	c, err := s.pool.NewClient()
	if err != nil {
		return m, nil
	}
	defer c.Close()

	values, err := c.Rscan(keyStart, keyEnd, limit)
	if err != nil {
		return m, err
	}
	for key, v := range values{
		m[key] = v.String()
	}
	return m, nil
}

func  (s *SSDBStorage) GetKeysCount(start, end string) int64  {
	c, _ := s.pool.NewClient()

	count := int64(0)
	limit := int64(5000)

	for {
		keys, err := c.Keys(start, end, limit)
		if err != nil {
			break
		}
		if len(keys) == 0 {
			break
		}

		count += int64(len(keys))

		fmt.Printf("count=%d, key=%s\n", count, start)
		//for _, key := range keys {
		//	fmt.Println(key)
		//}

		start = keys[len(keys)-1]

	}
	return count
}


func (s *SSDBStorage) GetKeyCount() int64 {
	return s.GetKeysCount("", "")
}


func (s *SSDBStorage) GetAddrCount() int64 {
	return s.GetKeysCount("addr-", "addr-zzzzzzzzzzzz")
}


func (s *SSDBStorage) GetTxCount() int64 {
	return s.GetKeysCount("tx-", "tx-zzzzzzzzzzz")
}

func (s *SSDBStorage) GetTxNodeCount() int64 {
	return s.GetKeysCount("txnode-", "txnode-zzzzzzzzzz")
}

func (s *SSDBStorage) GetUtxoCount() int64 {
	return s.GetKeysCount("utxo-", "utxo-zzzzzzzzzz")
}


func (s * SSDBStorage) GetTx(txid string) (*models.Transaction, error) {
	//fmt.Println("gettx:" , txid)
	data, err := s.Get(models.TxKey(txid))
	if err != nil{
		fmt.Printf("GetTx %s\n" , err)
		return nil, err
	}
	tx := &models.Transaction{}

	err = models.Decode([]byte(data), tx)
	if err != nil{
		fmt.Printf("GetTx Decode %s\n" , err)
		return nil, err
	}
	return  tx, nil
}

type TxAddr struct {
	addr *models.Addr
	Type uint32
}

func (s * SSDBStorage) ProcessTx()  {
	t0 := time.Now().UnixNano() / 1000000

	defer func() {
		t1 := time.Now().UnixNano()/1000000
		fmt.Printf("GetTx take %f\n", float64(t1-t0) /1000)
	}()

	txid := "2cb25da415724a396070a99afcf5db0d9635987f474bbbaa324d46c4464c533a"
	tx, err := s.GetTx(txid)
	if err != nil{
		fmt.Printf("ProcessTx GetTx %s\n" , err)
		return
	}

	curTxAddr := map[string]*TxAddr{}

	for _, in := range tx.Inputs {
		preTx := &models.Transaction{}
		err := preTx.Load(s, in.TxId)
		if err != nil {
			logger.Errorf("ProcessBlockTx preTx.Load %s", err)
			return
		}

		for _, preOutput := range preTx.Outputs {
			if preOutput.Vout == in.Vout {
				in.Amount = preOutput.Amount
				in.Addr = preOutput.Addr

				strAddr := preOutput.Addr
				txAddr, exist := curTxAddr[strAddr]
				if !exist {
					addr := &models.Addr{}
					err = addr.Load(s, strAddr)
					if err != nil {
						logger.Errorf("ProcessBlockTx addr.Load %s", err)
						return
					}
					txAddr = &TxAddr{
						addr: addr,
						Type: models.TxTypeInput,
					}
					curTxAddr[strAddr] = txAddr
				}

				//删除utxo
				//fmt.Println("delete:", txAddr.addr.Id, preTx.Id, in.Vout)
				//err := models.DeleteUtxo(s, txAddr.addr.Id, preTx.Id, in.Vout)
				//if err != nil {
				//	logger.Errorf("ProcessBlockTx models.DeleteUtxo %s", err)
				//}
				break
			}
		}
	}


	//fmt.Printf("tx: %v\n", tx)
}


func main() {
	s, _ := NewSSDBStorage(&conf.Config{
		Host: "192.168.8.54",
		Port: 8666,
		//Port:3561,
	})

	//s.ProcessTx()

	fmt.Println("keys count=", s.GetKeyCount())

	//fmt.Println("address count=", s.GetAddrCount())

	//fmt.Println("tx count=", s.GetTxCount())

	//fmt.Println("txnode count=", s.GetTxNodeCount())

	//fmt.Println("utxo count=", s.GetUtxoCount())
}

//testnet3
//Chain state (height 1325132, hash 00000000308ee4816c4842598a96520f34f633dce9622ead4490a8f53ebd7f99,
//totaltx 18381330, work 1433119409278111028294)
//400000
//address count= 5250047
//tx count= 3684574
//txnode count= 16466594