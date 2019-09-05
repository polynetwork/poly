package inf

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native/service/native"
	"github.com/ontio/multi-chain/native/service/native/utils"
	"math/big"
	"sort"
)

var (
	KEY_PREFIX_BTC = "btc"
	KEY_PREFIX_ETH = "eth"

	KEY_PREFIX_BTC_VOTE = "btcVote"
	KEY_PREFIX_ETH_VOTE = "ethVote"
)

type ChainHandler interface {
	Vote(service *native.NativeService) (bool, *MakeTxParam, error)
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
	sourceChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("EntranceParam deserialize sourcechainid error:%s", err)
	}
	txData, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("EntranceParam deserialize txdata error:%s", err)
	}
	height, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("EntranceParam deserialize height error:%s", err)
	}
	proof, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("EntranceParam deserialize proof error:%s", err)
	}
	relayerAddr, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("EntranceParam deserialize relayerAddr error:%s", err)
	}
	targetChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("EntranceParam deserialize targetchainid error:%s", err)
	}
	value, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("EntranceParam deserialize value error:%s", err)
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
	utils.EncodeVarUint(sink, this.SourceChainID)
	utils.EncodeString(sink, this.TxData)
	utils.EncodeVarUint(sink, uint64(this.Height))
	utils.EncodeString(sink, this.Proof)
	utils.EncodeString(sink, this.RelayerAddress)
	utils.EncodeVarUint(sink, this.TargetChainID)
	utils.EncodeString(sink, this.Value)
}

type MakeTxParam struct {
	FromChainID         uint64
	FromContractAddress string
	ToChainID           uint64
	ToAddress           string
	Amount              *big.Int
}

func (this *MakeTxParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.FromChainID)
	utils.EncodeString(sink, this.FromContractAddress)
	utils.EncodeVarUint(sink, this.ToChainID)
	utils.EncodeString(sink, this.ToAddress)
	utils.EncodeVarBytes(sink, this.Amount.Bytes())
}

func (this *MakeTxParam) Deserialization(source *common.ZeroCopySource) error {
	fromChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("MakeTxParam deserialize fromChainID error:%s", err)
	}
	fromContractAddress, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("MakeTxParam deserialize fromContractAddress error:%s", err)
	}
	toChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("MakeTxParam deserialize toChainID error:%s", err)
	}
	toAddress, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("MakeTxParam deserialize toAddress error:%s", err)
	}
	amount, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("MakeTxParam deserialize amount error:%s", err)
	}

	this.FromChainID = fromChainID
	this.FromContractAddress = fromContractAddress
	this.ToChainID = toChainID
	this.ToAddress = toAddress
	this.Amount = new(big.Int).SetBytes(amount)
	return nil
}

type VoteParam struct {
	FromChainID uint64
	Address     string
	TxHash      []byte
}

func (this *VoteParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.FromChainID)
	utils.EncodeString(sink, this.Address)
	utils.EncodeVarBytes(sink, this.TxHash)
}

func (this *VoteParam) Deserialization(source *common.ZeroCopySource) error {
	fromChainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("VoteParam deserialize fromChainID error:%s", err)
	}
	address, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("VoteParam deserialize address error:%s", err)
	}
	txHash, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("VoteParam deserialize txHash error:%s", err)
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
	utils.EncodeVarUint(sink, uint64(len(this.VoteMap)))
	var voteList []string
	for _, v := range this.VoteMap {
		voteList = append(voteList, v)
	}
	sort.SliceStable(voteList, func(i, j int) bool {
		return voteList[i] > voteList[j]
	})
	for _, v := range voteList {
		utils.EncodeString(sink, v)
	}
}

func (this *Vote) Deserialization(source *common.ZeroCopySource) error {
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize VoteMap length error: %v", err)
	}
	voteMap := make(map[string]string)
	for i := 0; uint64(i) < n; i++ {
		v, err := utils.DecodeString(source)
		if err != nil {
			return fmt.Errorf("deserialize VoteMap error: %v", err)
		}
		voteMap[v] = v
	}
	this.VoteMap = voteMap
	return nil
}