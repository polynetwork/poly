package btc

import (
	"bytes"
	"fmt"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/gcash/bchd/chaincfg/chainhash"
	"github.com/ontio/multi-chain/common"
	"sort"
	"strconv"
)

type BtcProof struct {
	Tx           []byte
	Proof        []byte
	Height       uint32
	BlocksToWait uint64
}

func (this *BtcProof) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.Tx)
	sink.WriteVarBytes(this.Proof)
	sink.WriteUint32(this.Height)
	sink.WriteUint64(this.BlocksToWait)
}

func (this *BtcProof) Deserialization(source *common.ZeroCopySource) error {
	tx, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("BtcProof deserialize tx error")
	}
	proof, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("BtcProof deserialize proof error")
	}
	height, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("BtcProof deserialize height error")
	}
	blocksToWait, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("BtcProof deserialize blocksToWait error:")
	}

	this.Tx = tx
	this.Proof = proof
	this.Height = uint32(height)
	this.BlocksToWait = blocksToWait
	return nil
}

type Utxos struct {
	Utxos []*Utxo
}

func (this *Utxos) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(uint64(len(this.Utxos)))
	for _, v := range this.Utxos {
		v.Serialization(sink)
	}
}

func (this *Utxos) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize Utxos length error")
	}
	utxos := make([]*Utxo, 0)
	for i := 0; uint64(i) < n; i++ {
		utxo := new(Utxo)
		if err := utxo.Deserialization(source); err != nil {
			return fmt.Errorf("deserialize utxo error: %v", err)
		}
		utxos = append(utxos, utxo)
	}

	this.Utxos = utxos
	return nil
}

func (this *Utxos) Len() int {
	return len(this.Utxos)
}

func (this *Utxos) Less(i, j int) bool {
	if this.Utxos[i].Value == this.Utxos[j].Value {
		return bytes.Compare(this.Utxos[i].Op.Hash, this.Utxos[j].Op.Hash) == -1
	}
	return this.Utxos[i].Value < this.Utxos[j].Value
}

func (this *Utxos) Swap(i, j int) {
	temp := this.Utxos[i]
	this.Utxos[i] = this.Utxos[j]
	this.Utxos[j] = temp
}

type Utxo struct {
	// Previous txid and output index
	Op *OutPoint

	// Block height where this tx was confirmed, 0 for unconfirmed
	AtHeight uint32 // TODO: del ??

	// The higher the better
	Value uint64

	// Output script
	ScriptPubkey []byte
}

func (this *Utxo) Serialization(sink *common.ZeroCopySink) {
	this.Op.Serialization(sink)
	sink.WriteUint32(this.AtHeight)
	sink.WriteUint64(this.Value)
	sink.WriteVarBytes(this.ScriptPubkey)
}

func (this *Utxo) Deserialization(source *common.ZeroCopySource) error {
	op := new(OutPoint)
	err := op.Deserialization(source)
	if err != nil {
		return fmt.Errorf("Utxo deserialize OutPoint error:%s", err)
	}
	atHeight, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("OutPoint deserialize atHeight error")
	}
	value, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("OutPoint deserialize value error")
	}
	scriptPubkey, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("OutPoint deserialize scriptPubkey error")
	}

	this.Op = op
	this.AtHeight = atHeight
	this.Value = value
	this.ScriptPubkey = scriptPubkey
	return nil
}

type CoinSelector struct {
	sortedUtxos *Utxos
	mc          uint64
	target      uint64
	maxP        float64
	txOuts      []*wire.TxOut
	k           float64
	tries       int64
	feeRate     uint64
	m           int
	n           int
}

func (selector *CoinSelector) Select() ([]*Utxo, uint64, uint64) {
	if selector.sortedUtxos == nil || len(selector.sortedUtxos.Utxos) == 0 {
		return nil, 0, 0
	}
	//selector.mixUpUtxos()
	result, sum, fee := selector.SimpleBnbSearch(0, make([]*Utxo, 0), 0)
	if result != nil {
		return result, sum, fee
	}
	//sort.Sort(sort.Reverse(selector.SortedUtxos))
	result, sum, fee = selector.SortedSearch()
	return result, sum, fee
}

func (selector *CoinSelector) SimpleBnbSearch(depth int, selection []*Utxo, sum uint64) ([]*Utxo, uint64, uint64) {
	fee, lr := selector.getLossRatio(selection)
	switch {
	case lr >= selector.maxP, float64(sum) > selector.k*float64(selector.target):
		return nil, 0, 0
	case sum == selector.target || (sum >= selector.target+selector.mc && float64(sum) <= selector.k*float64(selector.target)):
		return selection, sum, fee
	case selector.tries <= 0, depth == -1:
		return nil, 0, 0
	default:
		selector.tries--
		var next int
		if depth > selector.sortedUtxos.Len()/2 {
			next = selector.sortedUtxos.Len() - depth
		} else if depth < selector.sortedUtxos.Len()/2 {
			next = selector.sortedUtxos.Len() - depth - 1
		} else {
			next = -1
		}
		result, resSum, fee := selector.SimpleBnbSearch(next, append(selection, selector.sortedUtxos.Utxos[depth]),
			sum+selector.sortedUtxos.Utxos[depth].Value)
		if result != nil {
			return result, resSum, fee
		}
		if next == -1 {
			return nil, 0, 0
		}
		result, resSum, fee = selector.SimpleBnbSearch(next, selection, sum)
		return result, resSum, fee
	}
}

func (selector *CoinSelector) SortedSearch() ([]*Utxo, uint64, uint64) {
	selection := make([]*Utxo, 0)
	sum := uint64(0)
	pass := 0
	fee := uint64(0)
	lr := 0.0
	for _, u := range selector.sortedUtxos.Utxos {
		switch pass {
		case 0:
			selection = append(selection, u)
			sum += u.Value
			fee, lr = selector.getLossRatio(selection)
			if lr >= selector.maxP {
				if txscript.IsPayToScriptHash(u.ScriptPubkey) {
					selection = selection[:len(selection)-1]
					continue
				}
				return nil, 0, 0
			}
			if sum == selector.target || sum >= selector.target+selector.mc {
				pass = 1
			}
		case 1:
			feeReplaced, lr := selector.getLossRatio(append(selection[:len(selection)-1:cap(selection)-1], u))
			if sumTemp := sum - selection[len(selection)-1].Value + u.Value; (sumTemp == selector.target ||
				sumTemp >= selector.target+selector.mc) && lr < selector.maxP {
				fee, sum = feeReplaced, sumTemp
				selection[len(selection)-1] = u
			} else {
				return selection, sum, fee
			}
		}
	}
	if pass == 1 {
		return selection, sum, fee
	}
	return nil, 0, 0
}

//func (selector *CoinSelector) mixUpUtxos() {
//	length := selector.SortedUtxos.Len()
//	if length <= 2 {
//		return
//	}
//	mid := func() int {
//		if length%2 == 0 {
//			return selector.SortedUtxos.Len()/2 - 1
//		} else {
//			return selector.SortedUtxos.Len()/2
//		}
//	}()
//
//	last := selector.SortedUtxos.Utxos[length-1]
//	selector.swapUtxo(1, mid)
//	selector.SortedUtxos.Utxos[1] = last
//}
//
//func (selector *CoinSelector) swapUtxo(n, mid int) {
//	if n == selector.SortedUtxos.Len()-1 {
//		return
//	}
//	var next int
//	if n <= mid {
//		next = 2*n
//	} else {
//		next = 2*(selector.SortedUtxos.Len()-n)-1
//	}
//	selector.swapUtxo(next, mid)
//	selector.SortedUtxos.Swap(n, next)
//}

func (selector *CoinSelector) getLossRatio(selection []*Utxo) (uint64, float64) {
	fee := selector.estimateTxFee(selection)
	return fee, float64(fee) / float64(selector.target)
}

func (selector *CoinSelector) estimateTxFee(selection []*Utxo) uint64 {
	size := uint64(selector.estimateTxSize(selection))
	return size * selector.feeRate
}

func (selector *CoinSelector) estimateTxSize(selection []*Utxo) int {
	redeemSize := 1 + selector.m*(1+75) + 1 + 1 + selector.n*(1+33) + 1 + 1
	p2shInputSize := 43 + redeemSize
	witnessInputSize := 41 + redeemSize/blockchain.WitnessScaleFactor
	outsSize := 0
	for _, txOut := range selector.txOuts {
		outsSize += txOut.SerializeSize()
	}
	witNum := 0
	for _, u := range selection {
		switch txscript.GetScriptClass(u.ScriptPubkey) {
		case txscript.WitnessV0ScriptHashTy:
			witNum++
		}
	}
	return 10 + 2 + wire.VarIntSerializeSize(uint64(len(selection))) +
		wire.VarIntSerializeSize(uint64(len(selector.txOuts)+1)) + (len(selection)-witNum)*p2shInputSize +
		witNum*witnessInputSize + outsSize
}

type OutPoint struct {
	Hash  []byte
	Index uint32
}

func (this *OutPoint) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.Hash)
	sink.WriteUint32(this.Index)
}

func (this *OutPoint) Deserialization(source *common.ZeroCopySource) error {
	hash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("OutPoint deserialize hash error")
	}
	index, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("OutPoint deserialize height error")
	}

	this.Hash = hash
	this.Index = index
	return nil
}

func (this *OutPoint) String() string {
	hash, err := chainhash.NewHash(this.Hash)
	if err != nil {
		return ""
	}

	return hash.String() + ":" + strconv.FormatUint(uint64(this.Index), 10)
}

type MultiSignInfo struct {
	MultiSignInfo map[string][][]byte
}

func (this *MultiSignInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(uint64(len(this.MultiSignInfo)))
	var MultiSignInfoList []string
	for k := range this.MultiSignInfo {
		MultiSignInfoList = append(MultiSignInfoList, k)
	}
	sort.SliceStable(MultiSignInfoList, func(i, j int) bool {
		return MultiSignInfoList[i] > MultiSignInfoList[j]
	})
	for _, k := range MultiSignInfoList {
		sink.WriteString(k)
		v := this.MultiSignInfo[k]
		sink.WriteUint64(uint64(len(v)))
		for _, b := range v {
			sink.WriteVarBytes(b)
		}
	}
}

func (this *MultiSignInfo) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MultiSignInfo deserialize MultiSignInfo length error")
	}
	multiSignInfo := make(map[string][][]byte)
	for i := 0; uint64(i) < n; i++ {
		k, eof := source.NextString()
		if eof {
			return fmt.Errorf("MultiSignInfo deserialize public key error")
		}
		m, eof := source.NextUint64()
		if eof {
			return fmt.Errorf("MultiSignInfo deserialize MultiSignItem length error")
		}
		multiSignItem := make([][]byte, 0)
		for j := 0; uint64(j) < m; j++ {
			b, eof := source.NextVarBytes()
			if eof {
				return fmt.Errorf("MultiSignInfo deserialize []byte error")
			}
			multiSignItem = append(multiSignItem, b)
		}
		multiSignInfo[k] = multiSignItem
	}
	this.MultiSignInfo = multiSignInfo
	return nil
}

type Args struct {
	ToChainID uint64
	Fee       int64
	Address   []byte
}

func (this *Args) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ToChainID)
	sink.WriteInt64(this.Fee)
	sink.WriteVarBytes(this.Address)
}

func (this *Args) Deserialization(source *common.ZeroCopySource) error {
	toChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("Args deserialize toChainID error")
	}
	fee, eof := source.NextInt64()
	if eof {
		return fmt.Errorf("Args deserialize fee error")
	}
	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Args deserialize address error")
	}

	this.ToChainID = toChainID
	this.Fee = fee
	this.Address = address
	return nil
}

type BtcFromInfo struct {
	FromTxHash  []byte
	FromChainID uint64
}

func (this *BtcFromInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.FromTxHash)
	sink.WriteUint64(this.FromChainID)
}

func (this *BtcFromInfo) Deserialization(source *common.ZeroCopySource) error {
	fromTxHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("BtcProof deserialize fromTxHash error")
	}
	fromChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("BtcProof deserialize fromChainID error:")
	}

	this.FromTxHash = fromTxHash
	this.FromChainID = fromChainID
	return nil
}
