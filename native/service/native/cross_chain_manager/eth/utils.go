package eth

import (
	"fmt"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native/service/native"
	"github.com/ontio/multi-chain/native/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/native/service/native/utils"
)

func putEthProof(native *native.NativeService, txHash, proof []byte) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(inf.KEY_PREFIX_ETH), txHash)
	native.CacheDB.Put(key, states.GenRawStorageItem(proof))
}

func getEthProof(native *native.NativeService, txHash []byte) ([]byte, error) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(inf.KEY_PREFIX_ETH), txHash)
	ethProofStore, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getEthProof, get ethProofStore error: %v", err)
	}
	if ethProofStore == nil {
		return nil, fmt.Errorf("getEthProof, can not find any records")
	}
	ethProofBytes, err := states.GetValueFromRawStorageItem(ethProofStore)
	if err != nil {
		return nil, fmt.Errorf("getEthProof, deserialize from raw storage item err:%v", err)
	}
	return ethProofBytes, nil
}

func putEthVote(native *native.NativeService, txHash []byte, vote *inf.Vote) error {
	sink := common.NewZeroCopySink(nil)
	vote.Serialization(sink)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(inf.KEY_PREFIX_ETH_VOTE), txHash)
	native.CacheDB.Put(key, states.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getEthVote(native *native.NativeService, txHash []byte) (*inf.Vote, error) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(inf.KEY_PREFIX_ETH_VOTE), txHash)
	ethVoteStore, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getEthVote, get ethTxStore error: %v", err)
	}
	vote := &inf.Vote{
		VoteMap: make(map[string]string),
	}
	if ethVoteStore != nil {
		ethVoteBytes, err := states.GetValueFromRawStorageItem(ethVoteStore)
		if err != nil {
			return nil, fmt.Errorf("getEthVote, deserialize from raw storage item err:%v", err)
		}
		err = vote.Deserialization(common.NewZeroCopySource(ethVoteBytes))
		if err != nil {
			return nil, fmt.Errorf("getEthVote, vote.Deserialization err:%v", err)
		}
	}
	return vote, nil
}
