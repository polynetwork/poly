/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package consensus_vote

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/common"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/utils"
)

const (
	VOTE_INFO = "voteInfo"
)

func getVoteInfo(native *native.NativeService, id []byte) (*VoteInfo, error) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(VOTE_INFO), id)
	voteInfoStore, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("getVoteInfo, get getVoteInfoStore error: %v", err)
	}

	voteInfo := &VoteInfo{
		VoteInfo: make(map[string]bool),
	}
	if voteInfoStore != nil {
		voteInfoBytes, err := cstates.GetValueFromRawStorageItem(voteInfoStore)
		if err != nil {
			return nil, fmt.Errorf("getVoteInfo, deserialize from raw storage item err:%v", err)
		}
		err = voteInfo.Deserialization(common.NewZeroCopySource(voteInfoBytes))
		if err != nil {
			return nil, fmt.Errorf("getVoteInfo, deserialize VoteInfo err:%v", err)
		}
	}
	return voteInfo, nil
}

func putVoteInfo(native *native.NativeService, id []byte, voteInfo *VoteInfo) {
	contract := utils.CrossChainManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	voteInfo.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(VOTE_INFO), id), cstates.GenRawStorageItem(sink.Bytes()))
}

func CheckVotes(native *native.NativeService, id []byte, address common.Address) (bool, error) {
	voteInfo, err := getVoteInfo(native, id)
	if err != nil {
		return false, fmt.Errorf("CheckVotes, getVoteInfo error: %v", err)
	}

	//check voteInfo status
	if voteInfo.Status {
		return false, nil
	}

	//check signs num
	//get view
	view, err := node_manager.GetView(native)
	if err != nil {
		return false, fmt.Errorf("CheckVotes, GetView error: %v", err)
	}
	//get consensus peer
	peerPoolMap, err := node_manager.GetPeerPoolMap(native, view)
	if err != nil {
		return false, fmt.Errorf("CheckVotes, GetPeerPoolMap error: %v", err)
	}
	num := 0
	sum := 0
	flag := false
	for key, v := range peerPoolMap.PeerPoolMap {
		if v.Status == node_manager.ConsensusStatus {
			k, err := hex.DecodeString(key)
			if err != nil {
				return false, fmt.Errorf("CheckVotes, hex.DecodeString public key error: %v", err)
			}
			publicKey, err := keypair.DeserializePublicKey(k)
			if err != nil {
				return false, fmt.Errorf("CheckVotes, keypair.DeserializePublicKey error: %v", err)
			}
			addr := types.AddressFromPubKey(publicKey)
			_, ok := voteInfo.VoteInfo[addr.ToBase58()]
			if ok {
				num = num + 1
			}
			sum = sum + 1
			_, ok = voteInfo.VoteInfo[address.ToBase58()]
			if addr == address && !ok {
				flag = true
			}
		}
	}
	if flag {
		voteInfo.VoteInfo[address.ToBase58()] = true
		num = num + 1
		putVoteInfo(native, id, voteInfo)
	}
	if num >= (2*sum+2)/3 {
		voteInfo.Status = true
		putVoteInfo(native, id, voteInfo)
		return true, nil
	} else {
		return false, nil
	}
}
