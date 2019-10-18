package common

import (
	"fmt"
	"github.com/ontio/multi-chain/native/service/utils"
	"sort"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
)

var (
	KEY_PREFIX_BTC = "btc"
	KEY_PREFIX_ETH = "eth"

	KEY_PREFIX_BTC_VOTE = "btcVote"
	KEY_PREFIX_ETH_VOTE = "ethVote"
	REQUEST             = "request"
	DONE_TX             = "doneTx"
)

var NotifyMakeProofInfo = map[uint64]string{
	utils.ETH_CHAIN_ID: "makeToEthProof",
	utils.ONT_CHAIN_ID: "makeToOntProof",
	utils.NEO_CHAIN_ID: "makeToNeoProof",
}

type ChainHandler interface {
	MakeDepositProposal(service *native.NativeService) (*MakeTxParam, error)
}

type InitRedeemScriptParam struct {
	RedeemScript string
}

func (this *InitRedeemScriptParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.RedeemScript)
}

func (this *InitRedeemScriptParam) Deserialization(source *common.ZeroCopySource) error {
	redeemScript, eof := source.NextString()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize redeemScript error")
	}

	this.RedeemScript = redeemScript
	return nil
}

type EntranceParam struct {
	SourceChainID  uint64 `json:"sourceChainId"`
	TxHash         []byte `json:"txHash"`
	Height         uint32 `json:"height"`
	Proof          []byte `json:"proof"`
	RelayerAddress []byte `json:"relayerAddress"`
	Extra          []byte `json:"extra"`
}

func (this *EntranceParam) Deserialization(source *common.ZeroCopySource) error {
	sourceChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("EntranceParam deserialize sourcechainid error")
	}

	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("EntranceParam deserialize sourcechainid error")
	}

	height, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("EntranceParam deserialize height error")
	}
	proof, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("EntranceParam deserialize proof error")
	}
	relayerAddr, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("EntranceParam deserialize relayerAddr error")
	}
	extra, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("EntranceParam deserialize txdata error")
	}
	this.SourceChainID = sourceChainID
	this.TxHash = txHash
	this.Height = uint32(height)
	this.Proof = proof
	this.RelayerAddress = relayerAddr
	this.Extra = extra
	return nil
}

func (this *EntranceParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.SourceChainID)
	sink.WriteUint32(this.Height)
	sink.WriteVarBytes([]byte(this.Proof))
	sink.WriteVarBytes([]byte(this.RelayerAddress))
	sink.WriteVarBytes([]byte(this.Extra))
}

type MakeTxParam struct {
	TxHash              []byte
	FromChainID         uint64
	FromContractAddress []byte
	ToChainID           uint64
	ToContractAddress   []byte
	Method              string
	Args                []byte
}

func (this *MakeTxParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes([]byte(this.TxHash))
	sink.WriteUint64(this.FromChainID)
	sink.WriteVarBytes([]byte(this.FromContractAddress))
	sink.WriteUint64(this.ToChainID)
	sink.WriteVarBytes([]byte(this.ToContractAddress))
	sink.WriteVarBytes([]byte(this.Method))
	sink.WriteVarBytes([]byte(this.Args))
}

func (this *MakeTxParam) Deserialization(source *common.ZeroCopySource) error {
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize txHash error")
	}
	fromChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize fromChainID error")
	}
	fromContractAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize fromContractAddress error")
	}
	toChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize toChainID error")
	}
	toContractAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize toContractAddress error")
	}
	method, eof := source.NextString()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize method error")
	}
	args, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize args error")
	}

	this.TxHash = txHash
	this.FromChainID = fromChainID
	this.FromContractAddress = fromContractAddress
	this.ToChainID = toChainID
	this.ToContractAddress = toContractAddress
	this.Method = method
	this.Args = args
	return nil
}

type VoteParam struct {
	Address string
	TxHash  []byte
}

func (this *VoteParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes([]byte(this.Address))
	sink.WriteVarBytes(this.TxHash)
}

func (this *VoteParam) Deserialization(source *common.ZeroCopySource) error {
	address, eof := source.NextString()
	if eof {
		return fmt.Errorf("VoteParam deserialize address error")
	}
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("VoteParam deserialize txHash error")
	}

	this.Address = address
	this.TxHash = txHash
	return nil
}

type MultiSignParam struct {
	TxHash  []byte
	Address string
	Signs   [][]byte
}

func (this *MultiSignParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.TxHash)
	sink.WriteVarBytes([]byte(this.Address))
	sink.WriteUint64(uint64(len(this.Signs)))
	for _, v := range this.Signs {
		sink.WriteVarBytes(v)
	}
}

func (this *MultiSignParam) Deserialization(source *common.ZeroCopySource) error {
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize txHash error")
	}
	address, eof := source.NextString()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize address error")
	}
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize signs length error")
	}
	signs := make([][]byte, 0)
	for i := 0; uint64(i) < n; i++ {
		v, eof := source.NextVarBytes()
		if eof {
			return fmt.Errorf("deserialize Signs error")
		}
		signs = append(signs, v)
	}

	this.TxHash = txHash
	this.Address = address
	this.Signs = signs
	return nil
}

type Vote struct {
	VoteMap map[string]string
}

func (this *Vote) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(uint64(len(this.VoteMap)))
	var voteList []string
	for _, v := range this.VoteMap {
		voteList = append(voteList, v)
	}
	sort.SliceStable(voteList, func(i, j int) bool {
		return voteList[i] > voteList[j]
	})
	for _, v := range voteList {
		sink.WriteVarBytes([]byte(v))
	}
}

func (this *Vote) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize VoteMap length error")
	}
	voteMap := make(map[string]string)
	for i := 0; uint64(i) < n; i++ {
		v, eof := source.NextString()
		if eof {
			return fmt.Errorf("deserialize VoteMap error")
		}
		voteMap[v] = v
	}
	this.VoteMap = voteMap
	return nil
}

type ToMerkleValue struct {
	TxHash      []byte
	MakeTxParam *MakeTxParam
}

func (this *ToMerkleValue) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.TxHash)
	this.MakeTxParam.Serialization(sink)
}

func (this *ToMerkleValue) Deserialization(source *common.ZeroCopySource) error {
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MerkleValue deserialize txHash error")
	}
	makeTxParam := new(MakeTxParam)
	err := makeTxParam.Deserialization(source)
	if err != nil {
		return fmt.Errorf("MerkleValue deserialize makeTxParam error:%s", err)
	}

	this.TxHash = txHash
	this.MakeTxParam = makeTxParam
	return nil
}
