package models

import (
	"testing"
	"time"
	"github.com/liyue201/btc-scan/store/mockdb"
)

//go test -v -run='TestTx'
func TestTx(t *testing.T) {

	store := mockdb.NewMockDBStorage()
	tx := &Transaction{TxId: "aaaaaa", BlockTime: time.Now().Unix(), BlockHash: "bbbb", BlockHeight: 1}

	tx.Inputs = append(tx.Inputs, TxInput{TxId: "ggggg", PrvVout: 4})
	tx.Outputs = append(tx.Outputs, TxOutput{Addr:"hhh", vout: 2, Amount: 100})

	err := tx.Save(store)
	if err != nil {
		t.Error(err)
		return
	}

	tx2 := Transaction{}
	err = tx2.Load(store, tx.TxId)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("tx2=%v", tx2)

	tx3 := Transaction{}
	err = tx3.Load(store, "afsfsdaf")
	if err != nil {
		t.Logf(err.Error())
	}
}
