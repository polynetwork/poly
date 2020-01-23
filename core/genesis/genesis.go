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

package genesis

import (
	"fmt"
	"github.com/ontio/multi-chain/native/service/utils"
	"time"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/common/constants"
	"github.com/ontio/multi-chain/consensus/vbft/config"
	"github.com/ontio/multi-chain/core/payload"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native/states"
	"github.com/ontio/ontology-crypto/keypair"
)

const (
	BlockVersion uint32 = 0
	GenesisNonce uint64 = 2083236893

	INIT_CONFIG = "initConfig"
	INIT_ONT    = "init" // should be the same name as Ont contract INIT_NAME
)

var GenBlockTime = (config.DEFAULT_GEN_BLOCK_TIME * time.Second)

var INIT_PARAM = map[string]string{
	"gasPrice": "0",
}

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
	ontInitTx := newOntInit()
	consensusPayload, err := vconfig.GenesisConsensusPayload(0)
	if err != nil {
		return nil, fmt.Errorf("consensus genesis init failed: %s", err)
	}

	//blockdata
	genesisHeader := &types.Header{
		Version:          types.CURR_HEADER_VERSION,
		ChainID:          types.MAIN_CHAIN_ID,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        constants.GENESIS_BLOCK_TIMESTAMP,
		Height:           uint32(0),
		ConsensusData:    GenesisNonce,
		NextBookkeeper:   nextBookkeeper,
		ConsensusPayload: consensusPayload,

		Bookkeepers: nil,
		SigData:     nil,
	}

	if config.DefConfig.P2PNode.NetworkId == config.NETWORK_ID_MAIN_NET {
		genesisHeader.ChainID = types.MAIN_CHAIN_ID
	} else if config.DefConfig.P2PNode.NetworkId == config.NETWORK_ID_TEST_NET {
		genesisHeader.ChainID = types.TESTNET_CHAIN_ID
	}

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			nodeManagerConfig,
			ontInitTx,
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

func newOntInit() *types.Transaction {
	contractInvokeParam := &states.ContractInvokeParam{Address: utils.OntContractAddress,
		Method: INIT_ONT, Args: []byte{}}
	invokeCode := new(common.ZeroCopySink)
	contractInvokeParam.Serialization(invokeCode)

	return NewInvokeTransaction(invokeCode.Bytes(), 0)
}

//NewInvokeTransaction return smart contract invoke transaction
func NewInvokeTransaction(invokeCode []byte, nonce uint32) *types.Transaction {
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	tx := &types.Transaction{
		Version: types.CURR_TX_VERSION,
		ChainID: types.MAIN_CHAIN_ID,
		TxType:  types.Invoke,
		Payload: invokePayload,
		Nonce:   nonce,
	}

	if config.DefConfig.P2PNode.NetworkId == config.NETWORK_ID_MAIN_NET {
		tx.ChainID = types.MAIN_CHAIN_ID
	} else if config.DefConfig.P2PNode.NetworkId == config.NETWORK_ID_TEST_NET {
		tx.ChainID = types.TESTNET_CHAIN_ID
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
