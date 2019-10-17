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

package neo

import (
	"fmt"
	"github.com/joeqian10/neo-utils/neoutils/neorpc"
	"strconv"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	hscommon "github.com/ontio/multi-chain/native/service/header_sync/common"
)

type NEOHandler struct {
}

func NewNEOHandler() *NEOHandler {
	return &NEOHandler{}
}

func (this *NEOHandler) SyncGenesisHeader(native *native.NativeService) error {
	//params := new(hscommon.SyncGenesisHeaderParam)
	//if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
	//	return fmt.Errorf("SyncGenesisHeader, contract params deserialize error: %v", err)
	//}
	//
	//// get operator from database
	//operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	//if err != nil {
	//	return err
	//}
	//
	//////check witness
	////err = utils.ValidateOwner(native, operatorAddress)
	////if err != nil {
	////	return fmt.Errorf("SyncGenesisHeader, checkWitness error: %v", err)
	////}
	//
	//header, err := ntypes.HeaderFromRawBytes(params.GenesisHeader)
	//if err != nil {
	//	return fmt.Errorf("SyncGenesisHeader, deserialize header err: %v", err)
	//}
	////block header storage
	//err = PutBlockHeader(native, header)
	//if err != nil {
	//	return fmt.Errorf("SyncGenesisHeader, put blockHeader error: %v", err)
	//}
	//
	////consensus node pk storage
	//err = UpdateConsensusPeer(native, header, operatorAddress)
	//if err != nil {
	//	return fmt.Errorf("SyncGenesisHeader, update ConsensusPeer error: %v", err)
	//}
	return nil
}

func (this *NEOHandler) SyncBlockHeader(native *native.NativeService) error {
	params := new(hscommon.SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	for _, v := range params.Headers {
		header := &neorpc.BlockHeader{}
		header.ToBlockHeader(v)
		chainID, err := strconv.Atoi(header.ChainID)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, strconv.Atoi shardID error: %v", err)
		}
		_, err = GetHeaderByHeight(native,  header.Index)
		if err == nil {
			return fmt.Errorf("SyncBlockHeader, %d, %d", chainID, header.Index)
		}
		//err = verifyHeader(native, header)
		//if err != nil {
		//	return fmt.Errorf("SyncBlockHeader, verifyHeader error: %v", err)
		//}
		err = PutBlockHeader(native, header)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, put BlockHeader error: %v", err)
		}
		//err = UpdateConsensusPeer(native, header, params.Address)
		//if err != nil {
		//	return fmt.Errorf("SyncBlockHeader, update ConsensusPeer error: %v", err)
		//}
	}
	return nil
}
