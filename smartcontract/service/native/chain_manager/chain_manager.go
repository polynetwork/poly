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

package chain_manager

import (
	"bytes"
	"fmt"

	"encoding/hex"
	"encoding/json"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/header_sync"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	//status
	RegisterSideChainStatus Status = iota
	SideChainStatus
	QuitingStatus
)

const (
	//function name
	INIT_CONFIG             = "initConfig"
	REGISTER_MAIN_CHAIN     = "registerMainChain"
	REGISTER_SIDE_CHAIN     = "registerSideChain"
	SET_GOVERNANCE_EPOCH    = "setGovernanceEpoch"
	APPROVE_SIDE_CHAIN      = "approveSideChain"
	REJECT_SIDE_CHAIN       = "rejectSideChain"
	QUIT_SIDE_CHAIN         = "quitSideChain"
	APPROVE_QUIT_SIDE_CHAIN = "approveQuitSideChain"
	BLACK_SIDE_CHAIN        = "blackSideChain"
	STAKE_SIDE_CHAIN        = "stakeSideChain"
	UNSTAKE_SIDE_CHAIN      = "unStakeSideChain"
	INFLATION               = "inflation"
	APPROVE_INFLATION       = "approveInflation"
	REJECT_INFLATION        = "rejectInflation"
	IF_STAKED               = "ifStaked"
	GET_EPOCH               = "getEpoch"

	//key prefix
	MAIN_CHAIN            = "mainChain"
	SIDE_CHAIN            = "sideChain"
	INFLATION_INFO        = "inflationInfo"
	SIDE_CHAIN_STAKE_INFO = "sideChainStakeInfo"
	GOVERNANCE_EPOCH      = "governanceEpoch"
)

//Init governance contract address
func InitChainManager() {
	native.Contracts[utils.ChainManagerContractAddress] = RegisterChainManagerContract
}

//Register methods of governance contract
func RegisterChainManagerContract(native *native.NativeService) {
	native.Register(INIT_CONFIG, InitConfig)
	native.Register(REGISTER_MAIN_CHAIN, RegisterMainChain)
	native.Register(REGISTER_SIDE_CHAIN, RegisterSideChain)
	native.Register(SET_GOVERNANCE_EPOCH, SetGovernanceEpoch)
	native.Register(APPROVE_SIDE_CHAIN, ApproveSideChain)
	native.Register(REJECT_SIDE_CHAIN, RejectSideChain)
	native.Register(QUIT_SIDE_CHAIN, QuitSideChain)
	native.Register(APPROVE_QUIT_SIDE_CHAIN, ApproveQuitSideChain)
	native.Register(BLACK_SIDE_CHAIN, BlackSideChain)
	native.Register(STAKE_SIDE_CHAIN, StakeSideChain)
	native.Register(UNSTAKE_SIDE_CHAIN, UnStakeSideChain)
	native.Register(INFLATION, Inflation)
	native.Register(APPROVE_INFLATION, ApproveInflation)
	native.Register(REJECT_INFLATION, RejectInflation)

	native.Register(IF_STAKED, IfStaked)
	native.Register(GET_EPOCH, GetEpoch)
}

func InitConfig(native *native.NativeService) ([]byte, error) {
	configuration := new(config.VBFTConfig)
	buf, err := serialization.ReadVarBytes(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitConfig, serialization.ReadVarBytes error: %v", err)
	}
	if err := configuration.Deserialize(bytes.NewBuffer(buf)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitConfig, configuration.Deserialize error: %v", err)
	}

	//init admin OntID
	err = appCallInitContractAdmin(native, []byte(configuration.AdminOntID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitConfig, appCallInitContractAdmin error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func RegisterMainChain(native *native.NativeService) ([]byte, error) {
	params := new(RegisterMainChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterMainChain, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterMainChain, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterMainChain, checkWitness error: %v", err)
	}

	header, err := types.HeaderFromRawBytes(params.GenesisHeader)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterMainChain, deserialize header err: %v", err)
	}
	//block header storage
	err = header_sync.PutBlockHeader(native, header)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterMainChain, put blockHeader error: %v", err)
	}

	//consensus node pk storage
	err = header_sync.UpdateConsensusPeer(native, header, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterMainChain, update ConsensusPeer error: %v", err)
	}

	//update main chain
	err = putMainChain(native, header.ShardID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterMainChain, put MainChain error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

func RegisterSideChain(native *native.NativeService) ([]byte, error) {
	params := new(RegisterSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, contract params deserialize error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check auth of OntID
	err := appCallVerifyToken(native, contract, params.Caller, REGISTER_SIDE_CHAIN, uint64(params.KeyNo))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, verifyToken failed: %v", err)
	}

	header, err := types.HeaderFromRawBytes(params.GenesisBlockHeader)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, deserialize header err: %v", err)
	}

	//check if side chain exist
	chainIDBytes, err := utils.GetUint64Bytes(header.ShardID)
	if err != nil {
		return nil, fmt.Errorf("RegisterSideChain, getUint64Bytes error: %v", err)
	}
	sideChainBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIDE_CHAIN), chainIDBytes))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, get sideChainBytes error: %v", err)
	}
	if sideChainBytes != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, side chain is already registered")
	}

	//side chain storage
	sideChain := &SideChain{
		ChainID:            header.ShardID,
		Ratio:              uint64(params.Ratio),
		Deposit:            uint64(params.Deposit),
		OngNum:             0,
		OngPool:            uint64(params.OngPool),
		Status:             RegisterSideChainStatus,
		GenesisBlockHeader: params.GenesisBlockHeader,
	}
	err = putSideChain(native, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, put sideChain error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func SetGovernanceEpoch(native *native.NativeService) ([]byte, error) {
	params := new(GovernanceEpoch)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetGovernanceEpoch, contract params deserialize error: %v", err)
	}

	//get consensus multi sign address
	address, err := getConsensusMultiAddress(native, params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetGovernanceEpoch, getConsensusMultiAddress error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetGovernanceEpoch, checkWitness error: %v", err)
	}

	err = putGovernanceEpoch(native, params)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SetGovernanceEpoch, putGovernanceEpoch error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func ApproveSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainIDParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, contract params deserialize error: %v", err)
	}

	// get admin from database
	operatorAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, checkWitness error: %v", err)
	}

	//change side chain status
	sideChain, err := GetSideChain(native, params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, get sideChain error: %v", err)
	}
	if sideChain.Status != RegisterSideChainStatus {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, side chain is not register side chain status")
	}
	sideChain.Status = SideChainStatus

	err = putSideChain(native, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, put sideChain error: %v", err)
	}

	header, err := types.HeaderFromRawBytes(sideChain.GenesisBlockHeader)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, deserialize header err: %v", err)
	}
	//block header storage
	err = header_sync.PutBlockHeader(native, header)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, put blockHeader error: %v", err)
	}

	//consensus node pk storage
	err = header_sync.UpdateConsensusPeer(native, header, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, update ConsensusPeer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func RejectSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainIDParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectSideChain, contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectSideChain, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectSideChain, checkWitness error: %v", err)
	}

	//change side chain status
	sideChain, err := GetSideChain(native, params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectSideChain, get sideChain error: %v", err)
	}
	if sideChain.Status != RegisterSideChainStatus {
		return utils.BYTE_FALSE, fmt.Errorf("RejectSideChain, side chain is not register side chain status")
	}
	err = deleteSideChain(native, params.ChainID)
	if err != nil {
		return nil, fmt.Errorf("RejectSideChain, deleteSideChain error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func QuitSideChain(native *native.NativeService) ([]byte, error) {
	params := new(QuitSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, contract params deserialize error: %v", err)
	}

	//get consensus multi sign address
	address, err := getConsensusMultiAddress(native, params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, getConsensusMultiAddress error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, checkWitness error: %v", err)
	}

	//get side chain
	sideChain, err := GetSideChain(native, params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, get sideChain error: %v", err)
	}
	if sideChain.ChainID != params.ChainID {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, side chain is not registered")
	}
	if sideChain.Status != SideChainStatus {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, side chain is not side chain status")
	}
	sideChain.Status = QuitingStatus

	err = putSideChain(native, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, put sideChain error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func ApproveQuitSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainIDParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, checkWitness error: %v", err)
	}

	//get side chain
	sideChain, err := GetSideChain(native, params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, get sideChain error: %v", err)
	}
	if sideChain.ChainID != params.ChainID {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, side chain is not registered")
	}
	if sideChain.Status != QuitingStatus {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, side chain is not quiting status")
	}
	err = deleteSideChain(native, params.ChainID)
	if err != nil {
		return nil, fmt.Errorf("ApproveQuitSideChain, deleteSideChain error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func BlackSideChain(native *native.NativeService) ([]byte, error) {
	params := new(BlackSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackSideChain, contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackSideChain, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackSideChain, checkWitness error: %v", err)
	}

	//get side chain
	sideChain, err := GetSideChain(native, params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackSideChain, get sideChain error: %v", err)
	}

	amount := sideChain.OngNum + sideChain.Deposit
	//ong transfer
	err = appCallTransferOng(native, utils.ChainManagerContractAddress, params.Address, amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackSideChain, ong transfer error: %v", err)
	}

	err = deleteSideChain(native, params.ChainID)
	if err != nil {
		return nil, fmt.Errorf("BlackSideChain, deleteSideChain error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func StakeSideChain(native *native.NativeService) ([]byte, error) {
	params := new(StakeSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("StakeSideChain, contract params deserialize error: %v", err)
	}
	pk, err := hex.DecodeString(params.Pubkey)
	pubkey, err := keypair.DeserializePublicKey(pk)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("StakeSideChain, keypair.DeserializePublicKey error: %v", err)
	}
	address := types.AddressFromPubKey(pubkey)

	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("StakeSideChain, validateOwner error: %v", err)
	}

	//transfer ong
	err = appCallTransferOng(native, address, utils.ChainManagerContractAddress, params.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("StakeSideChain, ong transfer error: %v", err)
	}

	//put stake info into storage
	stakeInfo, err := GetStakeInfo(native, params.ChainID, params.Pubkey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("StakeSideChain, GetStakeInfo error: %v", err)
	}
	stakeInfo.ChainID = params.ChainID
	stakeInfo.Pubkey = params.Pubkey
	stakeInfo.Amount = stakeInfo.Amount + params.Amount

	err = putStakeInfo(native, stakeInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("StakeSideChain, putStakeInfo error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func UnStakeSideChain(native *native.NativeService) ([]byte, error) {
	params := new(StakeSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnStakeSideChain, contract params deserialize error: %v", err)
	}
	pk, err := hex.DecodeString(params.Pubkey)
	pubkey, err := keypair.DeserializePublicKey(pk)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("StakeSideChain, keypair.DeserializePublicKey error: %v", err)
	}
	address := types.AddressFromPubKey(pubkey)

	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnStakeSideChain, validateOwner error: %v", err)
	}

	//check if consensus peer
	consensusPeer1, consensusPeer2, err := header_sync.GetRecent2ConsensusPeers(native, params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnStakeSideChain, header_sync.GetConsensusPeers error: %v", err)
	}
	_, ok := consensusPeer1.PeerMap[params.Pubkey]
	if ok {
		return utils.BYTE_FALSE, fmt.Errorf("UnStakeSideChain, peer is consensus peer now, can not unstake")
	}
	_, ok = consensusPeer2.PeerMap[params.Pubkey]
	if ok {
		return utils.BYTE_FALSE, fmt.Errorf("UnStakeSideChain, peer is consensus peer previous epoch, can not unstake")
	}

	//update stake info
	stakeInfo, err := GetStakeInfo(native, params.ChainID, params.Pubkey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnStakeSideChain, GetStakeInfo error: %v", err)
	}
	stakeInfo.ChainID = params.ChainID
	stakeInfo.Pubkey = params.Pubkey
	if stakeInfo.Amount < params.Amount {
		return utils.BYTE_FALSE, fmt.Errorf("UnStakeSideChain, stake is not enough to withdraw")
	}
	stakeInfo.Amount = stakeInfo.Amount - params.Amount
	if stakeInfo.Amount == 0 {
		chainIDBytes, err := utils.GetUint64Bytes(stakeInfo.ChainID)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("UnStakeSideChain, getUint64Bytes error: %v", err)
		}
		pubkeyPrefix, err := hex.DecodeString(stakeInfo.Pubkey)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("UnStakeSideChain, hex.DecodeString pubkey format error: %v", err)
		}
		native.CacheDB.Delete(utils.ConcatKey(utils.ChainManagerContractAddress, []byte(SIDE_CHAIN_STAKE_INFO), chainIDBytes, pubkeyPrefix))
	} else {
		err = putStakeInfo(native, stakeInfo)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("UnStakeSideChain, putStakeInfo error: %v", err)
		}
	}

	//transfer ong
	err = appCallTransferOng(native, utils.ChainManagerContractAddress, address, params.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UnStakeSideChain, ong transfer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func Inflation(native *native.NativeService) ([]byte, error) {
	params := new(InflationParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, contract params deserialize error: %v", err)
	}

	//get consensus multi sign address
	address, err := getConsensusMultiAddress(native, params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, getConsensusMultiAddress error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, validateOwner error: %v", err)
	}

	//get side chain
	sideChain, err := GetSideChain(native, params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, get sideChain error: %v", err)
	}
	if sideChain.Status != SideChainStatus {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, side chain status is not normal status")
	}

	err = putInflationInfo(native, params)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, put inflationInfo error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func ApproveInflation(native *native.NativeService) ([]byte, error) {
	params := new(ChainIDParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveInflation, contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveInflation, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveInflation, checkWitness error: %v", err)
	}

	//get inflation info
	inflationInfo, err := getInflationInfo(native, params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveInflation, get inflationInfo error: %v", err)
	}

	//get side chain
	sideChain, err := GetSideChain(native, params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveInflation, get sideChain error: %v", err)
	}

	sideChain.Deposit = sideChain.Deposit + inflationInfo.DepositAdd
	sideChain.OngPool = sideChain.OngPool + inflationInfo.OngPoolAdd
	err = putSideChain(native, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveInflation, put sideChain error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func RejectInflation(native *native.NativeService) ([]byte, error) {
	params := new(ChainIDParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectInflation, contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectInflation, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectInflation, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	chainIDBytes, err := utils.GetUint64Bytes(params.ChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectInflation, getUint64Bytes error: %v", err)
	}
	native.CacheDB.Delete(utils.ConcatKey(contract, []byte(INFLATION_INFO), chainIDBytes))
	return utils.BYTE_TRUE, nil
}

func IfStaked(native *native.NativeService) ([]byte, error) {
	header, err := types.HeaderFromRawBytes(native.Input)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("IfStaked, types.HeaderFromRawBytes error: %s", err)
	}
	blkInfo := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(header.ConsensusPayload, blkInfo); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("IfStaked, unmarshal blockInfo error: %s", err)
	}
	if blkInfo.NewChainConfig == nil {
		return utils.BYTE_FALSE, fmt.Errorf("IfStaked, blkInfo.NewChainConfig is nil")
	}
	var stakeSum uint64
	for _, p := range blkInfo.NewChainConfig.Peers {
		stakeInfo, err := GetStakeInfo(native, header.ShardID, p.ID)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("IfStaked, GetStakeInfo error: %s", err)
		}
		stakeSum = stakeSum + stakeInfo.Amount
	}

	//get side chain info
	sideChain, err := GetSideChain(native, header.ShardID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("IfStaked, GetSideChain error: %s", err)
	}
	if stakeSum < sideChain.Deposit {
		return utils.BYTE_FALSE, fmt.Errorf("IfStaked, stake of consensus peer is not enough")
	}
	return utils.BYTE_TRUE, nil
}

func GetEpoch(native *native.NativeService) ([]byte, error) {
	chainID, err := utils.GetBytesUint64(native.Input)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetEpoch, utils.GetBytesUint64 error: %s", err)
	}
	governanceEpoch, err := GetGovernanceEpoch(native, chainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetEpoch, GetGovernanceEpoch error: %s", err)
	}
	r, err := utils.GetUint32Bytes(governanceEpoch.Epoch)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetEpoch, utils.GetUint64Bytes error: %s", err)
	}
	return r, nil
}
