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

package zilliqa

import (
	"encoding/json"
	"fmt"
	"github.com/Zilliqa/gozilliqa-sdk/core"
	"github.com/Zilliqa/gozilliqa-sdk/mpt"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	"github.com/polynetwork/poly/native/service/header_sync/zilliqa"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
)

// Handler ...
type Handler struct {
}

// NewHandler ...
func NewHandler() *Handler {
	return &Handler{}
}

// MakeDepositProposal ...
func (h *Handler) MakeDepositProposal(service *native.NativeService) (*scom.MakeTxParam, error) {
	params := new(scom.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("zilliqa MakeDepositProposal, contract params deserialize error: %s", err)
	}

	sideChain, err := side_chain_manager.GetSideChain(service, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("zilliqa MakeDepositProposal, side_chain_manager.GetSideChain error: %v", err)
	}

	value, err := verifyFromTx(service, params.Proof, params.Extra, params.SourceChainID, params.Height, sideChain)
	if err != nil {
		return nil, fmt.Errorf("eth MakeDepositProposal, verifyFromEthTx error: %s", err)
	}

	if err := scom.CheckDoneTx(service, value.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("eth MakeDepositProposal, check done transaction error:%s", err)
	}
	if err := scom.PutDoneTx(service, value.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("eth MakeDepositProposal, PutDoneTx error:%s", err)
	}
	return value, nil
}

// should be the same as relayer side
type ZILProof struct {
	Address       string         `json:"address"`
	Balance       string         `json:"balance"`
	CodeHash      string         `json:"codeHash"`
	Nonce         string         `json:"nonce"`
	StorageHash   string         `json:"storageHash"`
	AccountProof  []string       `json:"accountProof"`
	StorageProofs []StorageProof `json:"storageProof"`
}

// key should be storage key (in zilliqa)
type StorageProof struct {
	Key   []byte   `json:"key"`
	Value []byte   `json:"value"`
	Proof []string `json:"proof"`
}

func verifyFromTx(native *native.NativeService, proof, extra []byte, fromChainID uint64, height uint32, sideChain *side_chain_manager.SideChain) (param *scom.MakeTxParam, err error) {
	bestHeader, err := zilliqa.GetCurrentTxHeader(native, fromChainID)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromZilProof, get current header fail, error:%s", err)
	}

	bestHeight := uint32(bestHeader.BlockHeader.BlockNum)
	if bestHeight < height {
		return nil, fmt.Errorf("VerifyFromZilProof, transaction is not confirmed, current height: %d, input height: %d", bestHeight, height)
	}
	blockData, err := zilliqa.GetTxHeaderByHeight(native, uint64(height), fromChainID)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromZilProof, get header by height, height:%d, error:%s", height, err)
	}

	var zilProof ZILProof
	err = json.Unmarshal(proof, &zilProof)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromZilProof, unmarshal proof error:%s", err)
	}

	if len(zilProof.StorageProofs) != 1 {
		return nil, fmt.Errorf("VerifyFromZilProof, incorrect proof format")
	}

	// key is contract
	key := sideChain.CCMCAddress

	var pf [][]byte
	for _, p := range zilProof.AccountProof {
		bytes := util.DecodeHex(p)
		pf = append(pf, bytes)
	}
	db := mpt.NewFromProof(pf)
	root := blockData.BlockHeader.HashSet.StateRootHash[:]
	accountBaseBytes, err := mpt.Verify(key, db, root)
	if err != nil {
		return nil, fmt.Errorf("verifyMerkleProof, verify account proof error:%s\n", err)
	}

	accountBase, err := core.AccountBaseFromBytes(accountBaseBytes)
	if err != nil {
		return nil, fmt.Errorf("verifyMerkleProof, get account info error:%s\n", err)
	}

	var proof2 [][]byte
	for _, p := range zilProof.StorageProofs[0].Proof {
		bytes := util.DecodeHex(p)
		proof2 = append(proof2, bytes)
	}

	db2 := mpt.NewFromProof(proof2)
	value, err := mpt.Verify(zilProof.StorageProofs[0].Key, db2, accountBase.StorageRoot)
	if err != nil {
		return nil, fmt.Errorf("verifyMerkleProof, varify state info error:%s\n", err)
	}

	if value == nil {
		return nil, fmt.Errorf("VerifyFromZilProof, verifyMerkleProof failed!")
	}

	data := common.NewZeroCopySource(value)
	txParam := new(scom.MakeTxParam)
	if err := txParam.Deserialization(data); err != nil {
		return nil, fmt.Errorf("VerifyFromZilProof, deserialize merkleValue error:%s", err)
	}
	return txParam, nil
}
