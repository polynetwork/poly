/*
 * Copyright (C) 2022 The poly network Authors
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
package signature_manager

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
	SIG_INFO = "sigInfo"
)

func CheckSigns(native *native.NativeService, id, sig []byte, address common.Address) (bool, error) {
	sigInfo, err := getSigInfo(native, id)
	if err != nil {
		return false, fmt.Errorf("CheckSigs, getSigInfo error: %v", err)
	}

	//get view
	view, err := node_manager.GetView(native)
	if err != nil {
		return false, fmt.Errorf("CheckSigs, GetView error: %v", err)
	}
	//get consensus peer
	peerPoolMap, err := node_manager.GetPeerPoolMap(native, view)
	if err != nil {
		return false, fmt.Errorf("CheckSigs, GetPeerPoolMap error: %v", err)
	}

	//check if signer is consensus peer
	consensus := false
	for key, v := range peerPoolMap.PeerPoolMap {
		if v.Status == node_manager.ConsensusStatus {
			k, err := hex.DecodeString(key)
			if err != nil {
				return false, fmt.Errorf("CheckSigs, hex.DecodeString public key error: %v", err)
			}
			publicKey, err := keypair.DeserializePublicKey(k)
			if err != nil {
				return false, fmt.Errorf("CheckSigs, keypair.DeserializePublicKey error: %v", err)
			}
			addr := types.AddressFromPubKey(publicKey)

			if addr == address {
				consensus = true
				break
			}
		}
	}
	if !consensus {
		return false, fmt.Errorf("CheckSigs, signer is not consensus peer")
	}

	//check signs num
	num := 0
	sum := 0
	flag := false
	for key, v := range peerPoolMap.PeerPoolMap {
		if v.Status == node_manager.ConsensusStatus {
			k, err := hex.DecodeString(key)
			if err != nil {
				return false, fmt.Errorf("CheckSigs, hex.DecodeString public key error: %v", err)
			}
			publicKey, err := keypair.DeserializePublicKey(k)
			if err != nil {
				return false, fmt.Errorf("CheckSigs, keypair.DeserializePublicKey error: %v", err)
			}
			addr := types.AddressFromPubKey(publicKey)
			_, ok := sigInfo.SigInfo[addr.ToBase58()]
			if ok {
				num = num + 1
			}
			sum = sum + 1

			//check if voted
			_, ok = sigInfo.SigInfo[address.ToBase58()]
			if !ok {
				flag = true
			}
		}
	}
	if flag {
		sigInfo.SigInfo[address.ToBase58()] = sig
		num = num + 1
		if num < (2*sum+2)/3 {
			putSigInfo(native, id, sigInfo)
		}
	}
	if num >= (2*sum+2)/3 {
		shouldEmit := !sigInfo.Status
		sigInfo.Status = true
		putSigInfo(native, id, sigInfo)
		return shouldEmit, nil
	} else {
		return false, nil
	}
}

func getSigInfo(native *native.NativeService, id []byte) (*SigInfo, error) {
	key := utils.ConcatKey(utils.SignatureManagerContractAddress, []byte(SIG_INFO), id)
	sigInfoStore, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("getSigInfo, get getSigInfoStore error: %v", err)
	}

	sigInfo := &SigInfo{
		SigInfo: make(map[string][]byte),
	}
	if sigInfoStore != nil {
		sigInfoBytes, err := cstates.GetValueFromRawStorageItem(sigInfoStore)
		if err != nil {
			return nil, fmt.Errorf("getSigInfo, deserialize from raw storage item err:%v", err)
		}
		err = sigInfo.Deserialization(common.NewZeroCopySource(sigInfoBytes))
		if err != nil {
			return nil, fmt.Errorf("getSigInfo, deserialize SigInfo err:%v", err)
		}
	}
	return sigInfo, nil
}

func putSigInfo(native *native.NativeService, id []byte, sigInfo *SigInfo) {
	contract := utils.SignatureManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	sigInfo.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(SIG_INFO), id), cstates.GenRawStorageItem(sink.Bytes()))
}
