package eth

import (
	"fmt"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

func putEthProof(native *native.NativeService, txHash, proof []byte) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(inf.KEY_PREFIX_ETH), txHash)
	native.CacheDB.Put(key, states.GenRawStorageItem(proof))
}

func getEthProof(native *native.NativeService, txHash []byte) ([]byte, error) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(inf.KEY_PREFIX_ETH), txHash)
	btcProofStore, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getEthProof, get ethProofStore error: %v", err)
	}
	if btcProofStore == nil {
		return nil, fmt.Errorf("getEthProof, can not find any records")
	}
	btcProofBytes, err := states.GetValueFromRawStorageItem(btcProofStore)
	if err != nil {
		return nil, fmt.Errorf("getEthProof, deserialize from raw storage item err:%v", err)
	}
	return btcProofBytes, nil
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
	btcVoteStore, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getEthVote, get btcTxStore error: %v", err)
	}
	vote := &inf.Vote{
		VoteMap: make(map[string]string),
	}
	if btcVoteStore != nil {
		btcVoteBytes, err := states.GetValueFromRawStorageItem(btcVoteStore)
		if err != nil {
			return nil, fmt.Errorf("getEthVote, deserialize from raw storage item err:%v", err)
		}
		err = vote.Deserialization(common.NewZeroCopySource(btcVoteBytes))
		if err != nil {
			return nil, fmt.Errorf("getEthVote, vote.Deserialization err:%v", err)
		}
	}
	return vote, nil
}
