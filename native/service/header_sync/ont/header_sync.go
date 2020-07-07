/*
 * Copyright (C) 2020 The poly network Authors
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

package ont

import (
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/genesis"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/ontio/ontology-crypto/keypair"
	ocommon "github.com/ontio/ontology/common"
	otypes "github.com/ontio/ontology/core/types"
)

type ONTHandler struct {
}

func NewONTHandler() *ONTHandler {
	return &ONTHandler{}
}

func (this *ONTHandler) SyncGenesisHeader(native *native.NativeService) error {
	params := new(hscommon.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncGenesisHeader, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("SyncGenesisHeader, checkWitness error: %v", err)
	}

	header, err := otypes.HeaderFromRawBytes(params.GenesisHeader)
	if err != nil {
		return fmt.Errorf("SyncGenesisHeader, deserialize header err: %v", err)
	}
	//block header storage
	err = PutBlockHeader(native, params.ChainID, header)
	if err != nil {
		return fmt.Errorf("SyncGenesisHeader, put blockHeader error: %v", err)
	}

	//consensus node pk storage
	err = UpdateConsensusPeer(native, params.ChainID, header)
	if err != nil {
		return fmt.Errorf("SyncGenesisHeader, update ConsensusPeer error: %v", err)
	}
	return nil
}

func (this *ONTHandler) SyncBlockHeader(native *native.NativeService) error {
	params := new(hscommon.SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	for _, v := range params.Headers {
		header, err := otypes.HeaderFromRawBytes(v)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, otypes.HeaderFromRawBytes error: %v", err)
		}
		_, err = GetHeaderByHeight(native, params.ChainID, header.Height)
		if err == nil {
			continue
		}
		err = verifyHeader(native, params.ChainID, header)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, verifyHeader error: %v", err)
		}
		err = PutBlockHeader(native, params.ChainID, header)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, put BlockHeader error: %v", err)
		}
		err = UpdateConsensusPeer(native, params.ChainID, header)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, update ConsensusPeer error: %v", err)
		}
	}
	return nil
}

func (this *ONTHandler) SyncCrossChainMsg(native *native.NativeService) error {
	params := new(hscommon.SyncCrossChainMsgParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncCrossChainMsg, contract params deserialize error: %v", err)
	}
	for _, v := range params.CrossChainMsgs {
		source := ocommon.NewZeroCopySource(v)
		crossChainMsg := new(otypes.CrossChainMsg)
		err := crossChainMsg.Deserialization(source)
		if err != nil {
			return fmt.Errorf("SyncCrossChainMsg, deserialize crossChainMsg error: %v", err)
		}
		n, _, irr, eof := source.NextVarUint()
		if irr || eof {
			return fmt.Errorf("SyncCrossChainMsg, deserialization bookkeeper length error")
		}
		var bookkeepers []keypair.PublicKey
		for i := 0; uint64(i) < n; i++ {
			v, _, irr, eof := source.NextVarBytes()
			if irr || eof {
				return fmt.Errorf("SyncCrossChainMsg, deserialization bookkeeper error")
			}
			bookkeeper, err := keypair.DeserializePublicKey(v)
			if err != nil {
				return fmt.Errorf("SyncCrossChainMsg, keypair.DeserializePublicKey error: %v", err)
			}
			bookkeepers = append(bookkeepers, bookkeeper)
		}
		_, err = GetCrossChainMsg(native, params.ChainID, crossChainMsg.Height)
		if err == nil {
			continue
		}
		err = VerifyCrossChainMsg(native, params.ChainID, crossChainMsg, bookkeepers)
		if err != nil {
			return fmt.Errorf("SyncCrossChainMsg, VerifyCrossChainMsg error: %v", err)
		}
		err = PutCrossChainMsg(native, params.ChainID, crossChainMsg)
		if err != nil {
			return fmt.Errorf("SyncCrossChainMsg, put PutCrossChainMsg error: %v", err)
		}
	}
	return nil
}
