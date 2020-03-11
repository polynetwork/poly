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
	"math"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/governance/node_manager"
	"github.com/ontio/multi-chain/native/service/utils"
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
	SET_BTC_TX_PARAM            = "setBtcTxParam"

	//key prefix
	SIDE_CHAIN_APPLY          = "sideChainApply"
	UPDATE_SIDE_CHAIN_REQUEST = "updateSideChainRequest"
	QUIT_SIDE_CHAIN_REQUEST   = "quitSideChainRequest"
	SIDE_CHAIN                = "sideChain"
	REDEEM_BIND               = "redeemBind"
	BIND_SIGN_INFO            = "bindSignInfo"
	BTC_TX_PARAM              = "btcTxParam"
)

//Register methods of node_manager contract
func RegisterSideChainManagerContract(native *native.NativeService) {
	native.Register(REGISTER_SIDE_CHAIN, RegisterSideChain)
	native.Register(APPROVE_REGISTER_SIDE_CHAIN, ApproveRegisterSideChain)
	native.Register(UPDATE_SIDE_CHAIN, UpdateSideChain)
	native.Register(APPROVE_UPDATE_SIDE_CHAIN, ApproveUpdateSideChain)
	native.Register(QUIT_SIDE_CHAIN, QuitSideChain)
	native.Register(APPROVE_QUIT_SIDE_CHAIN, ApproveQuitSideChain)

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

	err = putSideChain(native, registerSideChain)
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
	if sideChain.ChainId == math.MaxUint64 {
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

	err = putSideChain(native, sideChain)
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
