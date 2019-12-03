package common

import (
	"fmt"
	"sort"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
)

var (
	KEY_PREFIX_BTC         = "btc"
	KEY_PREFIX_BTC_RELAYER = "btcRelayer"

	KEY_PREFIX_BTC_VOTE = "btcVote"
	REQUEST             = "request"
	DONE_TX             = "doneTx"

	NOTIFY_MAKE_PROOF = "makeProof"
)

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
	this.Height = height
	this.Proof = proof
	this.RelayerAddress = relayerAddr
	this.Extra = extra
	return nil
}

func (this *EntranceParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.SourceChainID)
	sink.WriteVarBytes(this.TxHash)
	sink.WriteUint32(this.Height)
	sink.WriteVarBytes(this.Proof)
	sink.WriteVarBytes(this.RelayerAddress)
	sink.WriteVarBytes(this.Extra)
}

type MakeTxParam struct {
	TxHash                   []byte
	FromContractAddress      []byte
	ToChainID                uint64
	ToContractAddress        []byte
	Fee                      uint64
	SourceFeeContractAddress []byte
	DestFeeContractAddress   []byte
	Method                   string
	Args                     []byte
}

func (this *MakeTxParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.TxHash)
	sink.WriteVarBytes(this.FromContractAddress)
	sink.WriteUint64(this.ToChainID)
	sink.WriteVarBytes(this.ToContractAddress)
	sink.WriteUint64(this.Fee)
	sink.WriteVarBytes(this.SourceFeeContractAddress)
	sink.WriteVarBytes(this.DestFeeContractAddress)
	sink.WriteVarBytes([]byte(this.Method))
	sink.WriteVarBytes(this.Args)
}

func (this *MakeTxParam) Deserialization(source *common.ZeroCopySource) error {
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize txHash error")
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
	fee, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize fee error")
	}
	sourceFeeContractAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize sourceFeeContractAddress error")
	}
	destFeeContractAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize destFeeContractAddress error")
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
	this.FromContractAddress = fromContractAddress
	this.ToChainID = toChainID
	this.ToContractAddress = toContractAddress
	this.Fee = fee
	this.SourceFeeContractAddress = sourceFeeContractAddress
	this.DestFeeContractAddress = destFeeContractAddress
	this.Method = method
	this.Args = args
	return nil
}

type VoteParam struct {
	FromChainID uint64
	Address     string
	TxHash      []byte
}

func (this *VoteParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.FromChainID)
	sink.WriteVarBytes([]byte(this.Address))
	sink.WriteVarBytes(this.TxHash)
}

func (this *VoteParam) Deserialization(source *common.ZeroCopySource) error {
	fromChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("VoteParam deserialize fromChainID error")
	}
	address, eof := source.NextString()
	if eof {
		return fmt.Errorf("VoteParam deserialize address error")
	}
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("VoteParam deserialize txHash error")
	}

	this.FromChainID = fromChainID
	this.Address = address
	this.TxHash = txHash
	return nil
}

type MultiSignParam struct {
	ChainID uint64
	TxHash  []byte
	Address string
	Signs   [][]byte
}

func (this *MultiSignParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ChainID)
	sink.WriteVarBytes(this.TxHash)
	sink.WriteVarBytes([]byte(this.Address))
	sink.WriteUint64(uint64(len(this.Signs)))
	for _, v := range this.Signs {
		sink.WriteVarBytes(v)
	}
}

func (this *MultiSignParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MultiSignParam deserialize txHash error")
	}
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

	this.ChainID = chainID
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
	FromChainID uint64
	Relayer     []byte
	Ratio       uint64
	MakeTxParam *MakeTxParam
}

func (this *ToMerkleValue) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.TxHash)
	sink.WriteUint64(this.FromChainID)
	sink.WriteVarBytes(this.Relayer)
	sink.WriteUint64(this.Ratio)
	this.MakeTxParam.Serialization(sink)
}

func (this *ToMerkleValue) Deserialization(source *common.ZeroCopySource) error {
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("ToMerkleValue deserialize txHash error")
	}
	fromChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("ToMerkleValue deserialize fromChainID error")
	}
	relayer, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("ToMerkleValue deserialize relayer error")
	}
	ratio, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("ToMerkleValue deserialize ratio error")
	}

	makeTxParam := new(MakeTxParam)
	err := makeTxParam.Deserialization(source)
	if err != nil {
		return fmt.Errorf("ToMerkleValue deserialize makeTxParam error:%s", err)
	}

	this.TxHash = txHash
	this.FromChainID = fromChainID
	this.Relayer = relayer
	this.Ratio = ratio
	this.MakeTxParam = makeTxParam
	return nil
}
