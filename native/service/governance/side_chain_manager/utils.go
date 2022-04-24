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

package side_chain_manager

import (
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/polynetwork/poly/common"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/utils"
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
	}
	return nil, nil

}

func PutSideChain(native *native.NativeService, sideChain *SideChain) error {
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
		return sideChain, nil
	} else {
		return nil, nil
	}
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

func getQuitSideChain(native *native.NativeService, chainid uint64) error {
	contract := utils.SideChainManagerContractAddress
	chainidByte := utils.GetUint64Bytes(chainid)

	chainIDStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(QUIT_SIDE_CHAIN_REQUEST),
		chainidByte))
	if err != nil {
		return fmt.Errorf("getQuitSideChain, get registerSideChainRequestStore error: %v", err)
	}
	if chainIDStore != nil {
		return nil
	}
	return fmt.Errorf("getQuitSideChain, no record")
}

func putQuitSideChain(native *native.NativeService, chainid uint64) error {
	contract := utils.SideChainManagerContractAddress
	chainidByte := utils.GetUint64Bytes(chainid)

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(QUIT_SIDE_CHAIN_REQUEST), chainidByte),
		cstates.GenRawStorageItem(chainidByte))
	return nil
}

func GetContractBind(native *native.NativeService, redeemChainID, contractChainID uint64,
	redeemKey []byte) (*ContractBinded, error) {
	contract := utils.SideChainManagerContractAddress
	redeemChainIDByte := utils.GetUint64Bytes(redeemChainID)
	contractChainIDByte := utils.GetUint64Bytes(contractChainID)
	contractBindStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(REDEEM_BIND),
		redeemChainIDByte, contractChainIDByte, redeemKey))
	if err != nil {
		return nil, fmt.Errorf("GetContractBind, get contractBindStore error: %v", err)
	}
	if contractBindStore != nil {
		val, err := cstates.GetValueFromRawStorageItem(contractBindStore)
		if err != nil {
			return nil, fmt.Errorf("GetContractBind, deserialize from raw storage item err:%v", err)
		}
		cb := &ContractBinded{}
		err = cb.Deserialization(common.NewZeroCopySource(val))
		if err != nil {
			return nil, fmt.Errorf("GetContractBind, deserialize BindContract err:%v", err)
		}
		return cb, nil
	} else {
		return nil, nil
	}

}

func putContractBind(native *native.NativeService, redeemChainID, contractChainID uint64,
	redeemKey, contractAddress []byte, cver uint64) error {
	contract := utils.SideChainManagerContractAddress
	redeemChainIDByte := utils.GetUint64Bytes(redeemChainID)
	contractChainIDByte := utils.GetUint64Bytes(contractChainID)
	bc := &ContractBinded{
		Contract: contractAddress,
		Ver:      cver,
	}
	sink := common.NewZeroCopySink(nil)
	bc.Serialization(sink)

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(REDEEM_BIND),
		redeemChainIDByte, contractChainIDByte, redeemKey), cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func putBindSignInfo(native *native.NativeService, message []byte, multiSignInfo *BindSignInfo) error {
	key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(BIND_SIGN_INFO), message)
	sink := common.NewZeroCopySink(nil)
	multiSignInfo.Serialization(sink)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getBindSignInfo(native *native.NativeService, message []byte) (*BindSignInfo, error) {
	key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(BIND_SIGN_INFO), message)
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

func putBtcTxParam(native *native.NativeService, redeemKey []byte, redeemChainId uint64, detail *BtcTxParamDetial) error {
	redeemChainIdBytes := utils.GetUint64Bytes(redeemChainId)
	sink := common.NewZeroCopySink(nil)
	detail.Serialization(sink)
	native.GetCacheDB().Put(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(BTC_TX_PARAM), redeemKey,
		redeemChainIdBytes), cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func GetBtcTxParam(native *native.NativeService, redeemKey []byte, redeemChainId uint64) (*BtcTxParamDetial, error) {
	redeemChainIdBytes := utils.GetUint64Bytes(redeemChainId)
	store, err := native.GetCacheDB().Get(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(BTC_TX_PARAM), redeemKey,
		redeemChainIdBytes))
	if err != nil {
		return nil, fmt.Errorf("GetBtcTxParam, get btcTxParam error: %v", err)
	}
	if store != nil {
		detialBytes, err := cstates.GetValueFromRawStorageItem(store)
		if err != nil {
			return nil, fmt.Errorf("GetBtcTxParam, deserialize from raw storage item error: %v", err)
		}
		detial := &BtcTxParamDetial{}
		err = detial.Deserialization(common.NewZeroCopySource(detialBytes))
		if err != nil {
			return nil, fmt.Errorf("GetBtcTxParam, deserialize BtcTxParam error: %v", err)
		}
		return detial, nil
	}
	return nil, nil
}

func verifyRedeemRegister(param *RegisterRedeemParam, addrs []btcutil.Address) (map[string][]byte, error) {
	r := make([]byte, len(param.Redeem))
	copy(r, param.Redeem)
	cverBytes := utils.GetUint64Bytes(param.CVersion)
	fromChainId := utils.GetUint64Bytes(param.RedeemChainID)
	toChainId := utils.GetUint64Bytes(param.ContractChainID)
	hash := btcutil.Hash160(append(append(append(append(r, fromChainId...), param.ContractAddress...),
		toChainId...), cverBytes...))
	return verify(param.Signs, addrs, hash)
}

func verifyBtcTxParam(param *BtcTxParam, addrs []btcutil.Address) (map[string][]byte, error) {
	r := make([]byte, len(param.Redeem))
	copy(r, param.Redeem)
	fromChainId := utils.GetUint64Bytes(param.RedeemChainId)
	frBytes := utils.GetUint64Bytes(param.Detial.FeeRate)
	mcBytes := utils.GetUint64Bytes(param.Detial.MinChange)
	verBytes := utils.GetUint64Bytes(param.Detial.PVersion)
	hash := btcutil.Hash160(append(append(append(append(r, fromChainId...), frBytes...), mcBytes...), verBytes...))
	return verify(param.Sigs, addrs, hash)
}

func verify(sigs [][]byte, addrs []btcutil.Address, hash []byte) (map[string][]byte, error) {
	res := make(map[string][]byte)
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

func putBtcRedeemScript(native *native.NativeService, redeemScriptKey string, redeemScriptBytes []byte, redeemChainId uint64) error {
	chainIDBytes := utils.GetUint64Bytes(redeemChainId)
	key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(REDEEM_SCRIPT), chainIDBytes, []byte(redeemScriptKey))

	cls := txscript.GetScriptClass(redeemScriptBytes)
	if cls.String() != "multisig" {
		return fmt.Errorf("putBtcRedeemScript, wrong type of redeem: %s", cls)
	}
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(redeemScriptBytes))
	return nil
}

func GetBtcRedeemScriptBytes(native *native.NativeService, redeemScriptKey string, redeemChainId uint64) ([]byte, error) {
	chainIDBytes := utils.GetUint64Bytes(redeemChainId)
	key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(REDEEM_SCRIPT), chainIDBytes, []byte(redeemScriptKey))
	redeemStore, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("getBtcRedeemScript, get btcProofStore error: %v", err)
	}
	if redeemStore == nil {
		return nil, fmt.Errorf("getBtcRedeemScript, can not find any records")
	}
	redeemBytes, err := cstates.GetValueFromRawStorageItem(redeemStore)
	if err != nil {
		return nil, fmt.Errorf("getBtcRedeemScript, deserialize from raw storage item err:%v", err)
	}
	return redeemBytes, nil
}

func PutAssetMap(native *native.NativeService, param *RegisterAssetParam) {
	chainIDBytes := utils.GetUint64Bytes(param.ChainId)
	key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(ASSET_MAP), chainIDBytes)
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
}

func GetAssetMap(native *native.NativeService, chainId uint64) (*RegisterAssetParam, error) {
	chainIDBytes := utils.GetUint64Bytes(chainId)
	key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(ASSET_MAP), chainIDBytes)
	store, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("GetAssetMap, get asset map store error: %v", err)
	}
	registerAssetParam := &RegisterAssetParam{
		AssetMap: make(map[uint64][]byte),
	}
	if store != nil {
		assetMapBytes, err := cstates.GetValueFromRawStorageItem(store)
		if err != nil {
			return nil, fmt.Errorf("GetAssetMap, deserialize from raw storage item err:%v", err)
		}
		err = registerAssetParam.Deserialization(common.NewZeroCopySource(assetMapBytes))
		if err != nil {
			return nil, fmt.Errorf("GetAssetMap, deserialize asset map err:%v", err)
		}
	}
	return registerAssetParam, nil
}

func PutAssetMapIndexes(native *native.NativeService, param *RegisterAssetParam) {
	chainIDBytes := utils.GetUint64Bytes(param.ChainId)
	for k, v := range param.AssetMap {
		kBytes := utils.GetUint64Bytes(k)
		key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(ASSET_MAP_INDEX), kBytes, v)
		native.GetCacheDB().Put(key, cstates.GenRawStorageItem(chainIDBytes))
	}
}

func GetAsset(native *native.NativeService, chainID uint64, assetAddress []byte) (uint64, error) {
	chainIDBytes := utils.GetUint64Bytes(chainID)
	key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(ASSET_MAP_INDEX), chainIDBytes, assetAddress)
	store, err := native.GetCacheDB().Get(key)
	if err != nil {
		return 0, fmt.Errorf("GetAssetMapIndex, get asset map index store error: %v", err)
	}
	if store == nil {
		return 0, fmt.Errorf("GetAssetMapIndex, cannot find any record")
	}
	r, err := cstates.GetValueFromRawStorageItem(store)
	if err != nil {
		return 0, fmt.Errorf("GetAssetMap, deserialize from raw storage item err:%v", err)
	}
	asset := utils.GetBytesUint64(r)
	return asset, nil
}

func PutFee(native *native.NativeService, chainId uint64, fee *Fee) {
	chainIdBytes := utils.GetUint64Bytes(chainId)
	key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(FEE), chainIdBytes)
	sink := common.NewZeroCopySink(nil)
	fee.Serialization(sink)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
}

func GetFee(native *native.NativeService, chainID uint64) (*Fee, error) {
	chainIDBytes := utils.GetUint64Bytes(chainID)
	key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(FEE), chainIDBytes)
	store, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("GetFee, get fee info store error: %v", err)
	}
	fee := new(Fee)
	if store != nil {
		feeBytes, err := cstates.GetValueFromRawStorageItem(store)
		if err != nil {
			return nil, fmt.Errorf("GetFee, deserialize from raw storage item err:%v", err)
		}
		err = fee.Deserialization(common.NewZeroCopySource(feeBytes))
		if err != nil {
			return nil, fmt.Errorf("GetFee, deserialize fee info err:%v", err)
		}
	}
	return fee, nil
}

func PutFeeInfo(native *native.NativeService, chainId, view uint64, feeInfo *FeeInfo) {
	chainIdBytes := utils.GetUint64Bytes(chainId)
	viewBytes := utils.GetUint64Bytes(view)
	key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(FEE_INFO), chainIdBytes, viewBytes)
	sink := common.NewZeroCopySink(nil)
	feeInfo.Serialization(sink)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
}

func GetFeeInfo(native *native.NativeService, chainID, view uint64) (*FeeInfo, error) {
	chainIDBytes := utils.GetUint64Bytes(chainID)
	viewBytes := utils.GetUint64Bytes(view)
	key := utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(FEE_INFO), chainIDBytes, viewBytes)
	store, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("GetFeeInfo, get fee info store error: %v", err)
	}
	feeInfo := &FeeInfo{
		FeeInfo: make(map[common.Address]*big.Int),
	}
	if store != nil {
		feeInfoBytes, err := cstates.GetValueFromRawStorageItem(store)
		if err != nil {
			return nil, fmt.Errorf("GetFeeInfo, deserialize from raw storage item err:%v", err)
		}
		err = feeInfo.Deserialization(common.NewZeroCopySource(feeInfoBytes))
		if err != nil {
			return nil, fmt.Errorf("GetFeeInfo, deserialize fee info err:%v", err)
		}
	}
	return feeInfo, nil
}

func GetRippleExtraInfo(native *native.NativeService, chainId uint64) (*RippleExtraInfo, error){
	sideChainInfo, err := GetSideChain(native, chainId)
	if err != nil {
		return nil, fmt.Errorf("GetRippleExtraInfo, GetSideChain error: %v", err)
	}
	rippleExtraInfo := &RippleExtraInfo{
		Pks: make([][]byte, 0),
	}
	err = rippleExtraInfo.Deserialization(common.NewZeroCopySource(sideChainInfo.ExtraInfo))
	if err != nil {
		return nil, fmt.Errorf("GetRippleExtraInfo, rippleExtraInfo.Deserialization error: %v", err)
	}
	return rippleExtraInfo, nil
}

func PutRippleExtraInfo(native *native.NativeService, chainId uint64, rippleExtraInfo *RippleExtraInfo) error {
	sink := common.NewZeroCopySink(nil)
	rippleExtraInfo.Serialization(sink)
	sideChainInfo, err := GetSideChain(native, chainId)
	if err != nil {
		return fmt.Errorf("PutRippleExtraInfo, GetSideChain error: %v", err)
	}
	sideChainInfo.ExtraInfo = sink.Bytes()
	err = PutSideChain(native, sideChainInfo)
	if err != nil {
		return fmt.Errorf("PutRippleExtraInfo, PutSideChain error: %v", err)
	}
	return nil
}