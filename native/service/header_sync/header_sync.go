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

package header_sync

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/merkle"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
)

const (
	//function name
	SYNC_GENESIS_HEADER  = "syncGenesisHeader"
	SYNC_BLOCK_HEADER    = "syncBlockHeader"
	SYNC_CONSENSUS_PEERS = "syncConsensusPeers"

	//key prefix
	BLOCK_HEADER                = "blockHeader"
	CURRENT_HEIGHT              = "currentHeight"
	HEADER_INDEX                = "headerIndex"
	CONSENSUS_PEER              = "consensusPeer"
	CONSENSUS_PEER_BLOCK_HEIGHT = "consensusPeerBlockHeight"
	KEY_HEIGHTS                 = "keyHeights"
	SYNC_ADDRESS                = "syncAddress"
)

//Init governance contract address
func InitHeaderSync() {
	native.Contracts[utils.HeaderSyncContractAddress] = RegisterHeaderSyncContract
}

//Register methods of governance contract
func RegisterHeaderSyncContract(native *native.NativeService) {
	native.Register(SYNC_GENESIS_HEADER, SyncGenesisHeader)
	native.Register(SYNC_BLOCK_HEADER, SyncBlockHeader)
	native.Register(SYNC_CONSENSUS_PEERS, SyncConsensusPeers)
}

func SyncGenesisHeader(native *native.NativeService) ([]byte, error) {
	params := new(SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, checkWitness error: %v", err)
	}

	header, err := types.HeaderFromRawBytes(params.GenesisHeader)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, deserialize header err: %v", err)
	}
	//block header storage
	err = PutBlockHeader(native, header)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, put blockHeader error: %v", err)
	}

	//consensus node pk storage
	err = UpdateConsensusPeer(native, header, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, update ConsensusPeer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func SyncBlockHeader(native *native.NativeService) ([]byte, error) {
	params := new(SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	for _, v := range params.Headers {
		header, err := types.HeaderFromRawBytes(v)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, new_types.HeaderFromRawBytes error: %v", err)
		}
		_, err = GetHeaderByHeight(native, header.ChainID, header.Height)
		if err == nil {
			return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, %d, %d", header.ChainID, header.Height)
		}
		err = verifyHeader(native, header)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, verifyHeader error: %v", err)
		}
		err = PutBlockHeader(native, header)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, put BlockHeader error: %v", err)
		}
		err = UpdateConsensusPeer(native, header, params.Address)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, update ConsensusPeer error: %v", err)
		}
	}
	return utils.BYTE_TRUE, nil
}

func SyncConsensusPeers(native *native.NativeService) ([]byte, error) {
	params := new(SyncConsensusPeerParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncConsensusPeers, contract params deserialize error: %v", err)
	}

	header, err := types.HeaderFromRawBytes(params.Header)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncConsensusPeers, types.HeaderFromRawBytes error: %v", err)
	}
	err = verifyHeader(native, header)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncConsensusPeers, verifyHeader error: %v", err)
	}

	path, err := hex.DecodeString(params.Proof)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncConsensusPeers, proof hex.DecodeString error: %v", err)
	}
	v := merkle.MerkleProve(path, header.CrossStatesRoot)
	if v == nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncConsensusPeers, merkle.MerkleProve verify merkle proof error")
	}
	consensusPeers := &ConsensusPeers{
		PeerMap: make(map[string]*Peer),
	}
	s := common.NewZeroCopySource(v)
	if err := consensusPeers.Deserialization(s); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncConsensusPeers, deserialize consensusPeers error:%s", err)
	}
	ok, err := checkIfConsensusPeersSynced(native, consensusPeers.ChainID, header.Height)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncConsensusPeers, checkIfConsensusPeersSynced error: %v", err)
	}
	if ok {
		return utils.BYTE_FALSE, fmt.Errorf("SyncConsensusPeers, consensusPeers are already synced")
	}
	err = putConsensusPeers(native, consensusPeers)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncConsensusPeers, put ConsensusPeers error: %s", err)
	}
	err = putSyncAddress(native, consensusPeers.ChainID, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("updateConsensusPeer, put SyncAddress eerror: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

