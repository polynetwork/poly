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
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"

	"github.com/polynetwork/poly/native/service/cross_chain_manager/consensus_vote"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/utils"
)

const (
	//function name
	REGISTER_SIDE_CHAIN         = "registerSideChain"
	APPROVE_REGISTER_SIDE_CHAIN = "approveRegisterSideChain"
	UPDATE_SIDE_CHAIN           = "updateSideChain"
	APPROVE_UPDATE_SIDE_CHAIN   = "approveUpdateSideChain"
	QUIT_SIDE_CHAIN             = "quitSideChain"
	APPROVE_QUIT_SIDE_CHAIN     = "approveQuitSideChain"
	REGISTER_REDEEM             = "registerRedeem"
	REGISTER_ASSET              = "registerAsset"
	UPDATE_FEE                  = "updateFee"
	SET_BTC_TX_PARAM            = "setBtcTxParam"

	//key prefix
	SIDE_CHAIN_APPLY          = "sideChainApply"
	UPDATE_SIDE_CHAIN_REQUEST = "updateSideChainRequest"
	QUIT_SIDE_CHAIN_REQUEST   = "quitSideChainRequest"
	SIDE_CHAIN                = "sideChain"
	REDEEM_BIND               = "redeemBind"
	BIND_SIGN_INFO            = "bindSignInfo"
	BTC_TX_PARAM              = "btcTxParam"
	REDEEM_SCRIPT             = "redeemScript"
	ASSET_BIND                = "assetBind"
	FEE                       = "fee"
	FEE_INFO                  = "feeInfo"

	UPDATE_FEE_TIMEOUT = 300
)

//Register methods of node_manager contract
func RegisterSideChainManagerContract(native *native.NativeService) {
	native.Register(REGISTER_SIDE_CHAIN, RegisterSideChain)
	native.Register(APPROVE_REGISTER_SIDE_CHAIN, ApproveRegisterSideChain)
	native.Register(UPDATE_SIDE_CHAIN, UpdateSideChain)
	native.Register(APPROVE_UPDATE_SIDE_CHAIN, ApproveUpdateSideChain)
	native.Register(QUIT_SIDE_CHAIN, QuitSideChain)
	native.Register(APPROVE_QUIT_SIDE_CHAIN, ApproveQuitSideChain)
	native.Register(REGISTER_ASSET, RegisterAsset)
	native.Register(UPDATE_FEE, UpdateFee)

	native.Register(REGISTER_REDEEM, RegisterRedeem)
	native.Register(SET_BTC_TX_PARAM, SetBtcTxParam)
}

func RegisterSideChain(native *native.NativeService) ([]byte, error) {
	params := new(RegisterSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, checkWitness error: %v", err)
	}

	registerSideChain, err := getSideChainApply(native, params.ChainId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, getRegisterSideChain error: %v", err)
	}
	if registerSideChain != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, chainid already requested")
	}
	sideChain, err := GetSideChain(native, params.ChainId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, getSideChain error: %v", err)
	}
	if sideChain != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, chainid already registered")
	}
	sideChain = &SideChain{
		Address:      params.Address,
		ChainId:      params.ChainId,
		Router:       params.Router,
		Name:         params.Name,
		BlocksToWait: params.BlocksToWait,
		CCMCAddress:  params.CCMCAddress,
		ExtraInfo:    params.ExtraInfo,
	}
	err = putSideChainApply(native, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, putRegisterSideChain error: %v", err)
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.NodeManagerContractAddress,
			States:          []interface{}{"RegisterSideChain", params.ChainId, params.Router, params.Name, params.BlocksToWait},
		})
	return utils.BYTE_TRUE, nil
}

func ApproveRegisterSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainidParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, checkWitness error: %v", err)
	}

	registerSideChain, err := getSideChainApply(native, params.Chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, getRegisterSideChain error: %v", err)
	}
	if registerSideChain == nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, chainid is not requested")
	}

	//check consensus signs
	ok, err := node_manager.CheckConsensusSigns(native, APPROVE_REGISTER_SIDE_CHAIN, utils.GetUint64Bytes(params.Chainid),
		params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}

	err = PutSideChain(native, registerSideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, putSideChain error: %v", err)
	}
	native.GetCacheDB().Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(SIDE_CHAIN_APPLY), utils.GetUint64Bytes(params.Chainid)))
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.NodeManagerContractAddress,
			States:          []interface{}{"ApproveRegisterSideChain", params.Chainid},
		})
	return utils.BYTE_TRUE, nil
}

func UpdateSideChain(native *native.NativeService) ([]byte, error) {
	params := new(RegisterSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, checkWitness error: %v", err)
	}

	sideChain, err := GetSideChain(native, params.ChainId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, getSideChain error: %v", err)
	}
	if sideChain == nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, side chain is not registered")
	}
	if sideChain.Address != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, side chain owner is wrong")
	}
	updateSideChain := &SideChain{
		Address:      params.Address,
		ChainId:      params.ChainId,
		Router:       params.Router,
		Name:         params.Name,
		BlocksToWait: params.BlocksToWait,
		CCMCAddress:  params.CCMCAddress,
		ExtraInfo:    params.ExtraInfo,
	}
	err = putUpdateSideChain(native, updateSideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, putUpdateSideChain error: %v", err)
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.NodeManagerContractAddress,
			States:          []interface{}{"UpdateSideChain", params.ChainId, params.Router, params.Name, params.BlocksToWait},
		})
	return utils.BYTE_TRUE, nil
}

func ApproveUpdateSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainidParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, checkWitness error: %v", err)
	}

	sideChain, err := getUpdateSideChain(native, params.Chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, getUpdateSideChain error: %v", err)
	}
	if sideChain == nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, chainid is not requested update")
	}

	//check consensus signs
	ok, err := node_manager.CheckConsensusSigns(native, APPROVE_UPDATE_SIDE_CHAIN, utils.GetUint64Bytes(params.Chainid),
		params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}

	err = PutSideChain(native, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, putSideChain error: %v", err)
	}
	chainidByte := utils.GetUint64Bytes(params.Chainid)
	native.GetCacheDB().Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(UPDATE_SIDE_CHAIN_REQUEST), chainidByte))
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.NodeManagerContractAddress,
			States:          []interface{}{"ApproveUpdateSideChain", params.Chainid},
		})
	return utils.BYTE_TRUE, nil
}

func QuitSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainidParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, checkWitness error: %v", err)
	}

	sideChain, err := GetSideChain(native, params.Chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, getSideChain error: %v", err)
	}
	if sideChain == nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, side chain is not registered")
	}
	if sideChain.Address != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, side chain owner is wrong")
	}

	err = putQuitSideChain(native, params.Chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, putUpdateSideChain error: %v", err)
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.NodeManagerContractAddress,
			States:          []interface{}{"QuitSideChain", params.Chainid},
		})
	return utils.BYTE_TRUE, nil
}

func ApproveQuitSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainidParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, checkWitness error: %v", err)
	}

	err = getQuitSideChain(native, params.Chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, getQuitSideChain error: %v", err)
	}

	//check consensus signs
	ok, err := node_manager.CheckConsensusSigns(native, QUIT_SIDE_CHAIN, utils.GetUint64Bytes(params.Chainid),
		params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}

	chainidByte := utils.GetUint64Bytes(params.Chainid)
	native.GetCacheDB().Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(QUIT_SIDE_CHAIN), chainidByte))
	native.GetCacheDB().Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(SIDE_CHAIN), chainidByte))
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.NodeManagerContractAddress,
			States:          []interface{}{"ApproveQuitSideChain", params.Chainid},
		})
	return utils.BYTE_TRUE, nil
}

func RegisterRedeem(native *native.NativeService) ([]byte, error) {
	params := new(RegisterRedeemParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRedeem, contract params deserialize error: %v", err)
	}
	ty, addrs, m, err := txscript.ExtractPkScriptAddrs(params.Redeem, netParam)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRedeem, failed to extract addrs: %v", err)
	}
	if ty != txscript.MultiSigTy {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRedeem, wrong type of redeem: %s", ty.String())
	}
	rk := btcutil.Hash160(params.Redeem)
	contract, err := GetContractBind(native, params.RedeemChainID, params.ContractChainID, rk)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRedeem, failed to get contract and version: %v", err)
	}
	if contract != nil && contract.Ver+1 != params.CVersion {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRedeem, previous version is %d and your version should "+
			"be %d not %d", contract.Ver, contract.Ver+1, params.CVersion)
	}
	verified, err := verifyRedeemRegister(params, addrs)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRedeem, failed to verify: %v", err)
	}
	key := append(append(append(rk, utils.GetUint64Bytes(params.RedeemChainID)...),
		params.ContractAddress...), utils.GetUint64Bytes(params.ContractChainID)...)
	bindSignInfo, err := getBindSignInfo(native, key)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRedeem, getBindSignInfo error: %v", err)
	}
	for k, v := range verified {
		bindSignInfo.BindSignInfo[k] = v
	}
	err = putBindSignInfo(native, key, bindSignInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRedeem, failed to putBindSignInfo: %v", err)
	}
	if len(bindSignInfo.BindSignInfo) >= m {
		err = putContractBind(native, params.RedeemChainID, params.ContractChainID, rk, params.ContractAddress, params.CVersion)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("RegisterRedeem, putContractBind error: %v", err)
		}
		if err = putBtcRedeemScript(native, hex.EncodeToString(rk), params.Redeem, params.RedeemChainID); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("RegisterRedeem, failed to save redeemscript %v with key %v, error: %v", hex.EncodeToString(params.Redeem), rk, err)
		}
		native.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: utils.SideChainManagerContractAddress,
				States:          []interface{}{"RegisterRedeem", hex.EncodeToString(rk), hex.EncodeToString(params.ContractAddress)},
			})
	}

	return utils.BYTE_TRUE, nil
}

func SetBtcTxParam(native *native.NativeService) ([]byte, error) {
	params := new(BtcTxParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetBtcTxParam, deserialize BtcTxParam error: %v", err)
	}
	if params.Detial.FeeRate == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("SetBtcTxParam, fee rate can't be zero")
	}
	if params.Detial.MinChange < 2000 {
		return utils.BYTE_FALSE, fmt.Errorf("SetBtcTxParam, min-change can't less than 2000")
	}
	cls, addrs, m, err := txscript.ExtractPkScriptAddrs(params.Redeem, netParam)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetBtcTxParam, extract addrs from redeem %v", err)
	}
	if cls != txscript.MultiSigTy {
		return utils.BYTE_FALSE, fmt.Errorf("SetBtcTxParam, redeem script is not multisig script: %s", cls.String())
	}
	rk := btcutil.Hash160(params.Redeem)
	prev, err := GetBtcTxParam(native, rk, params.RedeemChainId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetBtcTxParam, get previous param error: %v", err)
	}
	if prev != nil && params.Detial.PVersion != prev.PVersion+1 {
		return utils.BYTE_FALSE, fmt.Errorf("SetBtcTxParam, previous version is %d and your version should "+
			"be %d not %d", prev.PVersion, prev.PVersion+1, params.Detial.PVersion)
	}
	sink := common.NewZeroCopySink(nil)
	params.Detial.Serialization(sink)
	key := append(append(rk, utils.GetUint64Bytes(params.RedeemChainId)...), sink.Bytes()...)
	info, err := getBindSignInfo(native, key)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetBtcTxParam, getBindSignInfo error: %v", err)
	}
	if len(info.BindSignInfo) >= m {
		return utils.BYTE_FALSE, fmt.Errorf("SetBtcTxParam, the signatures are already enough")
	}
	verified, err := verifyBtcTxParam(params, addrs)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetBtcTxParam, failed to verify: %v", err)
	}
	for k, v := range verified {
		info.BindSignInfo[k] = v
	}
	if err = putBindSignInfo(native, key, info); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetBtcTxParam, failed to put bindSignInfo: %v", err)
	}
	if len(info.BindSignInfo) >= m {
		if err = putBtcTxParam(native, rk, params.RedeemChainId, params.Detial); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("SetBtcTxParam, failed to put btcTxParam: %v", err)
		}
		native.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: utils.SideChainManagerContractAddress,
				States: []interface{}{"SetBtcTxParam", hex.EncodeToString(rk), params.RedeemChainId,
					params.Detial.FeeRate, params.Detial.MinChange},
			})
	}
	return utils.BYTE_TRUE, nil
}

func RegisterAsset(native *native.NativeService) ([]byte, error) {
	params := new(RegisterAssetParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterAssetMap, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.OperatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterAssetMap, checkWitness error: %v", err)
	}

	rippleExtraInfo, err := GetRippleExtraInfo(native, params.ChainId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterAssetMap, GetRippleExtraInfo error: %v", err)
	}
	if rippleExtraInfo.Operator != params.OperatorAddress {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterAssetMap, caller is not operator")
	}

	assetBind, err := GetAssetBind(native, params.ChainId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterAssetMap, GetAssetMap error: %v", err)
	}
	for k, v := range params.AssetMap {
		assetBind.AssetMap[k] = v
	}
	for k, v := range params.LockProxyMap {
		assetBind.LockProxyMap[k] = v
	}

	PutAssetBind(native, params.ChainId, assetBind)
	return utils.BYTE_TRUE, nil
}

func UpdateFee(native *native.NativeService) ([]byte, error) {
	params := new(UpdateFeeParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateFee, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateFee, checkWitness error: %v", err)
	}

	//get fee
	fee, err := GetFee(native, params.ChainId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateFee, GetFee error: %v", err)
	}
	if fee.View != params.View {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateFee, poly view: %d, params view: %d not match",
			fee.View, params.View)
	}

	//add fee info
	feeInfo, err := GetFeeInfo(native, params.ChainId, fee.View)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateFee, GetFeeInfo error: %v", err)
	}
	if feeInfo.StartTime == 0 {
		feeInfo.StartTime = native.GetTime()
	} else if native.GetTime()-feeInfo.StartTime > UPDATE_FEE_TIMEOUT {
		// if time out view + 1
		fee.View = fee.View + 1
		PutFee(native, params.ChainId, fee)
		feeInfo = &FeeInfo{
			StartTime: native.GetTime(),
			FeeInfo:   make(map[common.Address]*big.Int),
		}
	}
	feeInfo.FeeInfo[params.Address] = params.Fee
	PutFeeInfo(native, params.ChainId, fee.View, feeInfo)

	//check consensus signs
	id := append([]byte(UPDATE_FEE), append(utils.GetUint64Bytes(params.ChainId), utils.GetUint64Bytes(fee.View)...)...)
	ok, err := consensus_vote.CheckVotes(native, id, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateFee, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}
	//vote enough
	feeInfoList := make([]*big.Int, 0, len(feeInfo.FeeInfo))
	for _, v := range feeInfo.FeeInfo {
		feeInfoList = append(feeInfoList, v)
	}
	sort.SliceStable(feeInfoList, func(i, j int) bool {
		return feeInfoList[i].Cmp(feeInfoList[j]) >= 1
	})
	l := len(feeInfoList)
	if l%2 == 0 {
		//even: (a + b)*5 / 2
		fee.Fee = new(big.Int).Div(new(big.Int).Mul(new(big.Int).Add(feeInfoList[l/2], feeInfoList[l/2-1]),
			new(big.Int).SetUint64(5)), new(big.Int).SetUint64(2))
	} else {
		//odd：a * 5
		fee.Fee = new(big.Int).Mul(feeInfoList[(l-1)/2], new(big.Int).SetUint64(5))
	}
	fee.View = fee.View + 1
	PutFee(native, params.ChainId, fee)
	return utils.BYTE_TRUE, nil
}
