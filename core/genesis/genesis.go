/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */

package genesis

import (
	"fmt"
	"github.com/polynetwork/poly/native/service/utils"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/common/constants"
	"github.com/polynetwork/poly/consensus/vbft/config"
	"github.com/polynetwork/poly/core/payload"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native/states"
)

const (
	BlockVersion uint32 = 0
	GenesisNonce uint64 = 2083236893

	INIT_CONFIG = "initConfig"
)

var GenBlockTime = (config.DEFAULT_GEN_BLOCK_TIME * time.Second)

var GenesisBookkeepers []keypair.PublicKey

// BuildGenesisBlock returns the genesis block with default consensus bookkeeper list
func BuildGenesisBlock(defaultBookkeeper []keypair.PublicKey, genesisConfig *config.GenesisConfig) (*types.Block, error) {
	//getBookkeeper
	GenesisBookkeepers = defaultBookkeeper
	nextBookkeeper, err := types.AddressFromBookkeepers(defaultBookkeeper)
	if err != nil {
		return nil, fmt.Errorf("[Block],BuildGenesisBlock err with GetBookkeeperAddress: %s", err)
	}
	conf := common.NewZeroCopySink(nil)
	if genesisConfig.VBFT != nil {
		genesisConfig.VBFT.Serialization(conf)
	}
	nodeManagerConfig := newNodeManagerInit(conf.Bytes())
	consensusPayload, err := vconfig.GenesisConsensusPayload(0)
	if err != nil {
		return nil, fmt.Errorf("consensus genesis init failed: %s", err)
	}

	//blockdata
	genesisHeader := &types.Header{
		Version:          types.CURR_HEADER_VERSION,
		ChainID:          config.GetChainIdByNetId(config.DefConfig.P2PNode.NetworkId),
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        constants.GENESIS_BLOCK_TIMESTAMP,
		Height:           uint32(0),
		ConsensusData:    GenesisNonce,
		NextBookkeeper:   nextBookkeeper,
		ConsensusPayload: consensusPayload,
		BlockRoot:        common.UINT256_EMPTY,
		Bookkeepers:      nil,
		SigData:          nil,
	}

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			nodeManagerConfig,
		},
	}
	genesisBlock.RebuildMerkleRoot()
	return genesisBlock, nil
}

func newNodeManagerInit(config []byte) *types.Transaction {
	tx, err := NewInitNodeManagerTransaction(config)
	if err != nil {
		panic("construct genesis node manager transaction error ")
	}
	return tx
}

//NewInvokeTransaction return smart contract invoke transaction
func NewInvokeTransaction(invokeCode []byte, nonce uint32) *types.Transaction {
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	tx := &types.Transaction{
		Version: types.CURR_TX_VERSION,
		TxType:  types.Invoke,
		Payload: invokePayload,
		Nonce:   nonce,
		ChainID: config.GetChainIdByNetId(config.DefConfig.P2PNode.NetworkId),
	}

	sink := common.NewZeroCopySink(nil)
	err := tx.Serialization(sink)
	if err != nil {
		return &types.Transaction{}
	}
	tx, err = types.TransactionFromRawBytes(sink.Bytes())
	return tx
}

func NewInitNodeManagerTransaction(
	paramBytes []byte,
) (*types.Transaction, error) {
	contractInvokeParam := &states.ContractInvokeParam{Address: utils.NodeManagerContractAddress,
		Method: INIT_CONFIG, Args: paramBytes}
	invokeCode := new(common.ZeroCopySink)
	contractInvokeParam.Serialization(invokeCode)

	return NewInvokeTransaction(invokeCode.Bytes(), 0), nil
}
