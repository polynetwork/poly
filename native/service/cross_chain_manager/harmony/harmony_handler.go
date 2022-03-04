/*
 * Copyright (C) 2022 The poly network Authors
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

package harmony

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	ecom "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/header_sync/harmony"
)

type Handler struct {}

func NewHandler() *Handler {
	return new(Handler)
}

// MakeDepositProposal ...
func (h *Handler) MakeDepositProposal(service *native.NativeService) (txParam *scom.MakeTxParam ,err error) {
	params := new(scom.EntranceParam)
	err = params.Deserialization(common.NewZeroCopySource(service.GetInput()))
	if err != nil {
		err = fmt.Errorf("HarmonyHandler failed to derserialize contract input, err: %v", err)
		return
	}

	txParam, err = h.VerifyDepositProposal(service, params)
	if err != nil {
		err = fmt.Errorf("HarmonyHandler failed to verify deposit proposal, err: %v", err)
		return
	}

	err = scom.CheckDoneTx(service, txParam.CrossChainID, params.SourceChainID)
	if err != nil {
		err = fmt.Errorf("HarmonyHandler check done transaction err: %v", err)
		return
	}

	err = scom.PutDoneTx(service, txParam.CrossChainID, params.SourceChainID)
	if err != nil {
		err = fmt.Errorf("HarmonyHandler failed to mark tx as done, err: %v", err)
		return
	}
	return
}

// Verify harmony deposit proposal
func (h *Handler) VerifyDepositProposal(
	service *native.NativeService, params *scom.EntranceParam) (txParam *scom.MakeTxParam, err error) {
	header, err := harmony.DecodeHeaderWithSig(params.HeaderOrCrossChainMsg)
	if err != nil {
		return
	}

	height := header.Header.Number().Uint64()
	if height != uint64(params.Height) {
		err = fmt.Errorf("header height(%v) does not match with proof height(%v)", height, params.Height)
		return
	}

	// Fetch current consensus epoch
	epoch, err := harmony.GetEpoch(service, params.SourceChainID)
	if err != nil || epoch == nil {
		err = fmt.Errorf("failed to get current epoch, err: %v", err)
		return
	}

	sideChain, err := side_chain_manager.GetSideChain(service, params.SourceChainID)
	if err != nil || sideChain == nil {
		err = fmt.Errorf("failed to get side chain instance, err: %v", err)
		return
	}
	ctx, err := harmony.DecodeHarmonyContext(sideChain.ExtraInfo)
	if err != nil {
		err = fmt.Errorf("failed to decode context, err: %v", err)
		return
	}
	err = ctx.Init()
	if err != nil {
		err = fmt.Errorf("failed to get network config and shard schedule: %v", err)
		return
	}
	// Check if current epoch is valid
	err = ctx.VerifyEpoch(epoch)
	if err != nil {
		err = fmt.Errorf("verify epoch failed for err: %v", err)
		return
	}

	// Verify proof header
	err = epoch.VerifyHeader(ctx, header.Header)
	if err !=  nil {
		err = fmt.Errorf("failed to verify header with current epoch info, err %v", err)
		return
	}

	// Verify header with signature
	err = epoch.VerifyHeaderSig(ctx, header)
	if err != nil {
		err = fmt.Errorf("verify header with signature failed, err: %v", err)
		return
	}

	// Verify eth proof
	proof := new(Proof)
	err = json.Unmarshal(params.Proof, proof)
	if err != nil {
		err = fmt.Errorf("decode eth proof failed, err: %v", err)
		return
	}
	err = VerifyCrossChainProof(crypto.Keccak256(params.Extra), proof, header.Header.Root(), sideChain.CCMCAddress)
	if err != nil {
		err = fmt.Errorf("VerifyCrossChainProof failed, err: %v", err)
		return
	}

	txParam, err = DecodeMakeTxParam(params.Extra)
	return
}

// Deserialize MakeTxParam
func DecodeMakeTxParam(data []byte) (txParam *scom.MakeTxParam, err error) {
	txParam = new(scom.MakeTxParam)
	err = txParam.Deserialization(common.NewZeroCopySource(data))
	if err != nil {
		txParam = nil
		err = fmt.Errorf("failed to deserialize MakeTxParam, err: %v", err)
	}
	return
}

// Proof ...
type Proof struct {
	Address       string         `json:"address"`
	Balance       string         `json:"balance"`
	CodeHash      string         `json:"codeHash"`
	Nonce         string         `json:"nonce"`
	StorageHash   string         `json:"storageHash"`
	AccountProof  []string       `json:"accountProof"`
	StorageProofs []StorageProof `json:"storageProof"`
}

// StorageProof ...
type StorageProof struct {
	Key   string   `json:"key"`
	Value string   `json:"value"`
	Proof []string `json:"proof"`
}

// ProofAccount ...
type ProofAccount struct {
	Nonce   *big.Int
	Balance  *big.Int
	Storage  ecom.Hash
	Codehash ecom.Hash
}

// Verify account proof and contract storage proof
func VerifyCrossChainProof(value []byte, proof *Proof, root ecom.Hash, address []byte) (err error) {
	nodeList := new(light.NodeList)
	for _, s := range proof.AccountProof {
		nodeList.Put(nil, ecom.Hex2Bytes(scom.Replace0x(s)))
	}
	ns := nodeList.NodeSet()
	addr := ecom.Hex2Bytes(scom.Replace0x(proof.Address))
	if !bytes.Equal(addr, address) {
		err = fmt.Errorf("contract address(%s) does not match with proof account(%s)", address, proof.Address)
		return
	}
	accountKey := crypto.Keccak256(addr)

	// Verify account proof
	accountValue, err := trie.VerifyProof(root, accountKey, ns)
	if err != nil {
		err = fmt.Errorf("account VerifyProof failure, err: %v", err)
		return
	}

	nonce, ok := new(big.Int).SetString(scom.Replace0x(proof.Nonce), 16)
	if !ok {
		err = fmt.Errorf("invalid account nonce: %s", proof.Nonce)
		return
	}
	balance, ok := new(big.Int).SetString(scom.Replace0x(proof.Balance), 16)
	if !ok {
		err = fmt.Errorf("invalid account balance: %s", proof.Balance)
		return
	}
	storageHash := ecom.HexToHash(proof.StorageHash)
	accountBytes, err := rlp.EncodeToBytes(&ProofAccount{
		Nonce: nonce,
		Balance: balance,
		Storage: storageHash,
		Codehash: ecom.HexToHash(proof.CodeHash),
	})
	if err != nil {
		err = fmt.Errorf("rlp encode account value failed, err: %v", err)
		return
	}
	if !bytes.Equal(accountBytes, accountValue) {
		err = fmt.Errorf("account value does not match, wanted: %x, got: %x", accountBytes, accountValue)
		return
	}

	// Verify storage proof
	if len(proof.StorageProofs) != 1 {
		err = fmt.Errorf("invalid storage proof size, %v", proof.StorageProofs)
		return
	}
	sp := proof.StorageProofs[0]
	nodeList = new(light.NodeList)
	storageKey := crypto.Keccak256(ecom.HexToHash(sp.Key).Bytes())
	for _, p := range sp.Proof {
		nodeList.Put(nil, ecom.Hex2Bytes(scom.Replace0x(p)))
	}
	storageValue, err := trie.VerifyProof(storageHash, storageKey, nodeList.NodeSet())
	if err != nil {
		err = fmt.Errorf("account storage VerifyProof failure, err: %v", err)
		return
	}
	err = CheckProofResult(storageValue, value)
	if err != nil {
		err = fmt.Errorf("CheckProofResult failed, err: %v", err)
		return
	}
	return
}

// Check proof storage value hash
func CheckProofResult(result, value []byte) (err error) {
	var temp []byte
	err = rlp.DecodeBytes(result, &temp)
	if err != nil {
		err = fmt.Errorf("rlp decode proof result failed, err: %v", err)
		return
	}
	var hash []byte
	for i := len(temp); i< 32; i++ {
		hash = append(hash, 0)
	}
	hash = append(hash, temp...)
	if !bytes.Equal(hash, value) {
		err = fmt.Errorf("storage value does not match with proof result, wanted %x, got %x", result, value)
		return
	}
	return
}