/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package side_chain_manager

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/ontio/multi-chain/common"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
)

var netParam = &chaincfg.TestNet3Params

func getSideChainApply(native *native.NativeService, chanid uint64) (*SideChain, error) {
	contract := utils.SideChainManagerContractAddress
	chainidByte := utils.GetUint64Bytes(chanid)

	sideChainStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(SIDE_CHAIN_APPLY),
		chainidByte))
	if err != nil {
		return nil, fmt.Errorf("getRegisterSideChain,get registerSideChainRequestStore error: %v", err)
	}
	sideChain := new(SideChain)
	if sideChainStore != nil {
		sideChainBytes, err := cstates.GetValueFromRawStorageItem(sideChainStore)
		if err != nil {
			return nil, fmt.Errorf("getRegisterSideChain, deserialize from raw storage item err:%v", err)
		}
		if err := sideChain.Deserialization(common.NewZeroCopySource(sideChainBytes)); err != nil {
			return nil, fmt.Errorf("getRegisterSideChain, deserialize sideChain error: %v", err)
		}
		return sideChain, nil
	} else {
		return nil, nil
	}
}

func putSideChainApply(native *native.NativeService, sideChain *SideChain) error {
	contract := utils.SideChainManagerContractAddress
	chainidByte := utils.GetUint64Bytes(sideChain.ChainId)

	sink := common.NewZeroCopySink(nil)
	err := sideChain.Serialization(sink)
	if err != nil {
		return fmt.Errorf("putRegisterSideChain, sideChain.Serialization error: %v", err)
	}

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(SIDE_CHAIN_APPLY), chainidByte),
		cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func GetSideChain(native *native.NativeService, chainID uint64) (*SideChain, error) {
	contract := utils.SideChainManagerContractAddress
	chainIDByte := utils.GetUint64Bytes(chainID)

	sideChainStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(SIDE_CHAIN),
		chainIDByte))
	if err != nil {
		return nil, fmt.Errorf("getSideChain,get registerSideChainRequestStore error: %v", err)
	}
	sideChain := new(SideChain)
	if sideChainStore != nil {
		sideChainBytes, err := cstates.GetValueFromRawStorageItem(sideChainStore)
		if err != nil {
			return nil, fmt.Errorf("getSideChain, deserialize from raw storage item err:%v", err)
		}
		if err := sideChain.Deserialization(common.NewZeroCopySource(sideChainBytes)); err != nil {
			return nil, fmt.Errorf("getSideChain, deserialize sideChain error: %v", err)
		}
		return sideChain, nil
	} else {
		return nil, nil
	}

}

func putSideChain(native *native.NativeService, sideChain *SideChain) error {
	contract := utils.SideChainManagerContractAddress
	chainidByte := utils.GetUint64Bytes(sideChain.ChainId)

	sink := common.NewZeroCopySink(nil)
	err := sideChain.Serialization(sink)
	if err != nil {
		return fmt.Errorf("putSideChain, sideChain.Serialization error: %v", err)
	}

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(SIDE_CHAIN), chainidByte),
		cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getUpdateSideChain(native *native.NativeService, chanid uint64) (*SideChain, error) {
	contract := utils.SideChainManagerContractAddress
	chainidByte := utils.GetUint64Bytes(chanid)

	sideChainStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(UPDATE_SIDE_CHAIN_REQUEST),
		chainidByte))
	if err != nil {
		return nil, fmt.Errorf("getUpdateSideChain,get registerSideChainRequestStore error: %v", err)
	}
	sideChain := new(SideChain)
	if sideChainStore != nil {
		sideChainBytes, err := cstates.GetValueFromRawStorageItem(sideChainStore)
		if err != nil {
			return nil, fmt.Errorf("getUpdateSideChain, deserialize from raw storage item err:%v", err)
		}
		if err := sideChain.Deserialization(common.NewZeroCopySource(sideChainBytes)); err != nil {
			return nil, fmt.Errorf("getUpdateSideChain, deserialize sideChain error: %v", err)
		}
	}
	return sideChain, nil
}

func putUpdateSideChain(native *native.NativeService, sideChain *SideChain) error {
	contract := utils.SideChainManagerContractAddress
	chainidByte := utils.GetUint64Bytes(sideChain.ChainId)

	sink := common.NewZeroCopySink(nil)
	err := sideChain.Serialization(sink)
	if err != nil {
		return fmt.Errorf("putUpdateSideChain, sideChain.Serialization error: %v", err)
	}

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(UPDATE_SIDE_CHAIN_REQUEST), chainidByte),
		cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func GetContractBind(native *native.NativeService, redeemChainID, contractChainID uint64,
	redeemKey string) ([]byte, error) {
	contract := utils.SideChainManagerContractAddress
	redeemChainIDByte := utils.GetUint64Bytes(redeemChainID)
	contractChainIDByte := utils.GetUint64Bytes(contractChainID)
	redeemKeyByte, err := hex.DecodeString(redeemKey)
	if err != nil {
		return nil, fmt.Errorf("GetContractBind, hex.DecodeString error: %v", err)
	}

	contractBindStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(REDEEM_BIND),
		redeemChainIDByte, contractChainIDByte, redeemKeyByte))
	if err != nil {
		return nil, fmt.Errorf("GetContractBind, get contractBindStore error: %v", err)
	}
	if contractBindStore != nil {
		contractBind, err := cstates.GetValueFromRawStorageItem(contractBindStore)
		if err != nil {
			return nil, fmt.Errorf("GetContractBind, deserialize from raw storage item err:%v", err)
		}
		return contractBind, nil
	} else {
		return nil, nil
	}

}

func putContractBind(native *native.NativeService, redeemChainID, contractChainID uint64,
	redeemKey, contractAddress []byte) error {
	contract := utils.SideChainManagerContractAddress
	redeemChainIDByte := utils.GetUint64Bytes(redeemChainID)
	contractChainIDByte := utils.GetUint64Bytes(contractChainID)

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(REDEEM_BIND),
		redeemChainIDByte, contractChainIDByte, redeemKey), cstates.GenRawStorageItem(contractAddress))
	return nil
}

func putBindSignInfo(native *native.NativeService, message []byte, multiSignInfo *BindSignInfo) error {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(BIND_SIGN_INFO), message)
	sink := common.NewZeroCopySink(nil)
	multiSignInfo.Serialization(sink)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getBindSignInfo(native *native.NativeService, message []byte) (*BindSignInfo, error) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(BIND_SIGN_INFO), message)
	bindSignInfoStore, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("getBtcMultiSignInfo, get multiSignInfoStore error: %v", err)
	}

	bindSignInfo := &BindSignInfo{
		BindSignInfo: make(map[string][]byte),
	}
	if bindSignInfoStore != nil {
		bindSignInfoBytes, err := cstates.GetValueFromRawStorageItem(bindSignInfoStore)
		if err != nil {
			return nil, fmt.Errorf("getBtcMultiSignInfo, deserialize from raw storage item err:%v", err)
		}
		err = bindSignInfo.Deserialization(common.NewZeroCopySource(bindSignInfoBytes))
		if err != nil {
			return nil, fmt.Errorf("getBtcMultiSignInfo, deserialize multiSignInfo err:%v", err)
		}
	}
	return bindSignInfo, nil
}

func verifyBtcSigs(sigs [][]byte, addrs []btcutil.Address, contract, redeem []byte) (map[string][]byte, error) {
	res := make(map[string][]byte)

	c := make([]byte, len(contract))
	copy(c, contract)
	r := make([]byte, len(redeem))
	copy(r, redeem)
	hash := btcutil.Hash160(append(r, c...))
	for i, sig := range sigs {
		if len(sig) < 1 {
			return nil, fmt.Errorf("length of no.%d sig is less than 1", i)
		}
		pSig, err := btcec.ParseDERSignature(sig, btcec.S256())
		if err != nil {
			return nil, fmt.Errorf("failed to parse no.%d sig: %v", i, err)
		}
		for _, addr := range addrs {
			if pSig.Verify(hash, addr.(*btcutil.AddressPubKey).PubKey()) {
				res[addr.EncodeAddress()] = sig
			}
		}
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("no sigs is verified")
	}
	return res, nil
}
