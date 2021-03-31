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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/polynetwork/poly/native/service/header_sync/zilliqa"
	"strings"

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
		return nil, fmt.Errorf("zil MakeDepositProposal, verifyFromZILTx error: %s", err)
	}

	if err := scom.CheckDoneTx(service, value.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("zil MakeDepositProposal, check done transaction error:%s", err)
	}
	if err := scom.PutDoneTx(service, value.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("zil MakeDepositProposal, PutDoneTx error:%s", err)
	}
	return value, nil
}

// should be the same as relayer side
type ZILProof struct {
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

	var pf [][]byte
	for _, p := range zilProof.AccountProof {
		bytes := util.DecodeHex(p)
		pf = append(pf, bytes)
	}

	db := mpt.NewFromProof(pf)
	root := blockData.BlockHeader.HashSet.StateRootHash[:]
	key := strings.TrimPrefix(util.EncodeHex(sideChain.CCMCAddress), "0x")
	accountBaseBytes, err := mpt.Verify([]byte(key), db, root)
	if err != nil {
		return nil, fmt.Errorf("verifyMerkleProof, verify account proof error:%s, key is %s proof is: %+v, root is %s", err, key, zilProof.AccountProof, util.EncodeHex(root))
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
	storageKey := util.DecodeHex(string(zilProof.StorageProofs[0].Key))
	hashedStorageKey := util.Sha256(storageKey)
	proofResult, err := mpt.Verify([]byte((util.EncodeHex(hashedStorageKey))), db2, accountBase.StorageRoot)
	if err != nil {
		return nil, fmt.Errorf("verifyMerkleProof, verify state proof error:%s, key is %s account proof is: %+v, state proof is: %+v, account bytes is: %s, root is %s", err,
			util.EncodeHex(storageKey), zilProof.AccountProof, zilProof.StorageProofs[0].Proof, util.EncodeHex(accountBaseBytes), util.EncodeHex(accountBase.StorageRoot))
	}

	if proofResult == nil {
		return nil, fmt.Errorf("verifyMerkleProof, verify state proof error:%s, key is %s account proof is: %+v, state proof is: %+v, account bytes is: %s, root is %s", "result is nil",
			util.EncodeHex(storageKey), zilProof.AccountProof, zilProof.StorageProofs[0].Proof, util.EncodeHex(accountBaseBytes), util.EncodeHex(accountBase.StorageRoot))
	}

	if !checkProofResult(proofResult, extra) {
		return nil, fmt.Errorf("verifyMerkleProof, check state proof result failed proof result: %s, extra: %s", util.EncodeHex(proofResult), util.EncodeHex(extra))
	}

	data := common.NewZeroCopySource(extra)
	txParam := new(scom.MakeTxParam)
	if err := txParam.Deserialization(data); err != nil {
		return nil, fmt.Errorf("VerifyFromZilProof, deserialize merkleValue error:%s", err)
	}
	return txParam, nil
}

func checkProofResult(result, value []byte) bool {
	origin := strings.ToLower(string(result))
	origin = strings.TrimPrefix(strings.ReplaceAll(origin, "\"", ""), "0x")

	hash := crypto.Keccak256(value)
	target := util.EncodeHex(hash)
	target = strings.ToLower(target)
	target = strings.TrimPrefix(target, "0x")

	return origin == target
}
