package models

import (
	"fmt"
	"github.com/golang/protobuf/proto"
)

//db store key
const (
	TxKeyPrefix        = "tx-" //确认的交易key前缀
	TxNodeKeyPrefix    = "txnode-"
	UtxoKeyPrefix      = "utxo-"
	ContextKey         = "context"
	BlockContextPrefix = "bctx-"
)

const (
	TxTypeInput  = 0x01
	TxTypeOutput = 0x02
)

//key:  tx-{tx.txid}
func TxKey(txid string) string {
	return TxKeyPrefix + txid
}

//key: txnode-{addr}-{TxNode.id}
func TxNodeKey(addr string, txnodeId int64) string {
	return TxNodeKeyPrefix + fmt.Sprintf("%s-%012d", addr, txnodeId)
}

//key:  utxo-{addr}-{Transaction.id}-{Transaction.vout}
func UtxoKey(addr string, txId int64, vout uint32) string {
	return UtxoKeyPrefix + fmt.Sprintf("%s-%012d-%05d", addr, txId, vout)
}

func TxNodeKeyStart(addr string) string {
	return TxNodeKeyPrefix + fmt.Sprintf("%s-", addr)
}

func TxNodeKeyEnd(addr string) string {
	return TxNodeKeyPrefix + fmt.Sprintf("%s-zzzzz", addr)
}

func UtxoKeyStart(addr string) string {
	return UtxoKeyPrefix + fmt.Sprintf("%s-", addr)
}

func UtxoKeyEnd(addr string) string {
	return UtxoKeyPrefix + fmt.Sprintf("%s-zzzzzz", addr)
}

func BlockCtxKey(blockHeight int64) string {
	return BlockContextPrefix + fmt.Sprintf("%d", blockHeight)
}

func Encode(v proto.Message) ([]byte, error) {
	return proto.Marshal(v)
}

func Decode(data []byte, v proto.Message) error {
	return proto.Unmarshal(data, v)
}
