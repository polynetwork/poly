package common

import (
	"fmt"
	"sort"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
)

var (
	KEY_PREFIX_BTC = "btc"
	KEY_PREFIX_ETH = "eth"

	KEY_PREFIX_BTC_VOTE = "btcVote"
	KEY_PREFIX_ETH_VOTE = "ethVote"
)

type ChainHandler interface {
	MakeDepositProposal(service *native.NativeService) (*MakeTxParam, error)
	MakeTransaction(service *native.NativeService, param *MakeTxParam) error
}

type EntranceParam struct {
	SourceChainID  uint64 `json:"sourceChainId"`
	TxData         string `json:"txData"`
	Height         uint32 `json:"height"`
	Proof          string `json:"proof"`
	RelayerAddress string `json:"relayerAddress"`
	TargetChainID  uint64 `json:"targetChainId"`
	Value          string `json:"value"`
}

func (this *EntranceParam) Deserialization(source *common.ZeroCopySource) error {
	sourceChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("EntranceParam deserialize sourcechainid error")
	}
	txData, eof := source.NextString()
	if eof {
		return fmt.Errorf("EntranceParam deserialize txdata error")
	}
	height, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("EntranceParam deserialize height error")
	}
	proof, eof := source.NextString()
	if eof {
		return fmt.Errorf("EntranceParam deserialize proof error")
	}
	relayerAddr, eof := source.NextString()
	if eof {
		return fmt.Errorf("EntranceParam deserialize relayerAddr error")
	}
	targetChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("EntranceParam deserialize targetchainid error")
	}
	value, eof := source.NextString()
	if eof {
		return fmt.Errorf("EntranceParam deserialize value error")
	}

	this.SourceChainID = sourceChainID
	this.TxData = txData
	this.Height = uint32(height)
	this.Proof = proof
	this.RelayerAddress = relayerAddr
	this.TargetChainID = targetChainID
	this.Value = value
	return nil
}

func (this *EntranceParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.SourceChainID)
	sink.WriteVarBytes([]byte(this.TxData))
	sink.WriteUint32(this.Height)
	sink.WriteVarBytes([]byte(this.Proof))
	sink.WriteVarBytes([]byte(this.RelayerAddress))
	sink.WriteUint64(this.TargetChainID)
	sink.WriteVarBytes([]byte(this.Value))
}

type MakeTxParam struct {
	TxHash              string
	FromChainID         uint64
	FromContractAddress string
	ToChainID           uint64
	Method              string
	Args                []byte
}

func (this *MakeTxParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes([]byte(this.TxHash))
	sink.WriteUint64(this.FromChainID)
	sink.WriteVarBytes([]byte(this.FromContractAddress))
	sink.WriteUint64(this.ToChainID)
	sink.WriteVarBytes([]byte(this.Method))
	sink.WriteVarBytes([]byte(this.Args))
}

func (this *MakeTxParam) Deserialization(source *common.ZeroCopySource) error {
	txHash, eof := source.NextString()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize txHash error")
	}
	fromChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize fromChainID error")
	}
	fromContractAddress, eof := source.NextString()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize fromContractAddress error")
	}
	toChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize toChainID error")
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
