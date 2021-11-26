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

package starcoin

import (
	"encoding/json"
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	cmanager "github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/header_sync/starcoin"
	"github.com/starcoinorg/starcoin-go/types"
)

type STCProof struct {
	AccountProof []string `json:"accountProof"`
}

func verifyFromEthTx(native *native.NativeService, proof, extra []byte, fromChainID uint64, height uint32, sideChain *cmanager.SideChain) (*scom.MakeTxParam, error) {
	bestHeader, err := starcoin.GetCurrentHeader(native, fromChainID)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, get current header fail, error:%s", err)
	}
	bestHeight := uint32(bestHeader.Number)
	if bestHeight < height || bestHeight-height < uint32(sideChain.BlocksToWait-1) {
		return nil, fmt.Errorf("VerifyFromEthProof, transaction is not confirmed, current height: %d, input height: %d", bestHeight, height)
	}

	blockData, err := starcoin.GetHeaderByHeight(native, uint64(height), fromChainID)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, get header by height, height:%d, error:%s", height, err)
	}

	stcProof := new(STCProof)
	err = json.Unmarshal(proof, stcProof)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromSTCProof, unmarshal proof error:%s", err)
	}

	_, err = VerifyEventProof(stcProof, blockData, sideChain.CCMCAddress)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, verifyMerkleProof error:%v", err)
	}

	data := common.NewZeroCopySource(extra)
	txParam := new(scom.MakeTxParam)
	if err := txParam.Deserialization(data); err != nil {
		return nil, fmt.Errorf("VerifyFromEthProof, deserialize merkleValue error:%s", err)
	}
	return txParam, nil
}

func VerifyEventProof(proof *STCProof, data *types.BlockHeader, address []byte) (bool, error) {
	//TODO add event proof verify
	return true, nil
}
