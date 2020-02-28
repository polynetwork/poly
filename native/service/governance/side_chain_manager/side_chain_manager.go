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
	REMOVE_SIDE_CHAIN           = "removeSideChain"
	REGISTER_REDEEM             = "registerRedeem"

	//key prefix
	SIDE_CHAIN_APPLY          = "sideChainApply"
	UPDATE_SIDE_CHAIN_REQUEST = "updateSideChainRequest"
	SIDE_CHAIN                = "sideChain"
	REDEEM_BIND               = "redeemBind"
	BIND_SIGN_INFO            = "bindSignInfo"
)

//Register methods of node_manager contract
func RegisterSideChainManagerContract(native *native.NativeService) {
	native.Register(REGISTER_SIDE_CHAIN, RegisterSideChain)
	native.Register(APPROVE_REGISTER_SIDE_CHAIN, ApproveRegisterSideChain)
	native.Register(UPDATE_SIDE_CHAIN, UpdateSideChain)
	native.Register(APPROVE_UPDATE_SIDE_CHAIN, ApproveUpdateSideChain)
	native.Register(REMOVE_SIDE_CHAIN, RemoveSideChain)

	native.Register(REGISTER_REDEEM, RegisterRedeem)
}

func RegisterSideChain(native *native.NativeService) ([]byte, error) {
	params := new(RegisterSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, contract params deserialize error: %v", err)
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
		ChainId:      params.ChainId,
		Router:       params.Router,
		Name:         params.Name,
		BlocksToWait: params.BlocksToWait,
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
	//check consensus signs
	ok, err := node_manager.CheckConsensusSigns(native, APPROVE_REGISTER_SIDE_CHAIN, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}

	registerSideChain, err := getSideChainApply(native, params.Chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, getRegisterSideChain error: %v", err)
	}
	if registerSideChain == nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, chainid is not requested")
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
	sideChain, err := GetSideChain(native, params.ChainId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, getSideChain error: %v", err)
	}
	if sideChain == nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, side chain is not registered")
	}
	updateSideChain := &SideChain{
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
	//check consensus signs
	ok, err := node_manager.CheckConsensusSigns(native, APPROVE_UPDATE_SIDE_CHAIN, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}

	sideChain, err := getUpdateSideChain(native, params.Chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, getUpdateSideChain error: %v", err)
	}
	if sideChain.ChainId == math.MaxUint64 {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, chainid is not requested update")
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

func RemoveSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainidParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveSideChain, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveSideChain, checkWitness error: %v", err)
	}
	//check consensus signs
	ok, err := node_manager.CheckConsensusSigns(native, REMOVE_SIDE_CHAIN, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveSideChain, CheckConsensusSigns error: %v", err)
	}
	if !ok {
		return utils.BYTE_TRUE, nil
	}

	sideChain, err := GetSideChain(native, params.Chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveSideChain, getUpdateSideChain error: %v", err)
	}
	if sideChain == nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveSideChain, side chain is not registered")
	}
	chainidByte := utils.GetUint64Bytes(params.Chainid)
	native.GetCacheDB().Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(SIDE_CHAIN), chainidByte))
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.NodeManagerContractAddress,
			States:          []interface{}{"RemoveSideChain", params.Chainid},
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
	contract, err := GetContractBind(native, params.RedeemChainID, params.ContractChainID, hex.EncodeToString(rk))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRedeem, failed to get contract and version: %v", err)
	}
	if contract != nil && contract.Ver >= params.CVersion {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterRedeem, contract version %d is less than last version %d",
			params.CVersion, contract.Ver)
	}
	verified, err := verifyBtcSigs(params.Signs, addrs, params.ContractAddress, params.Redeem, params.CVersion)
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
