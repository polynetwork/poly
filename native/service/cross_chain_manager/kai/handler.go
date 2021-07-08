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

package kai

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	kaiclient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/types"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	hskai "github.com/polynetwork/poly/native/service/header_sync/kai"
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
		return nil, fmt.Errorf("KAI MakeDepositProposal, contract params deserialize error: %s", err)
	}
	info, err := hskai.GetEpochSwitchInfo(service, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("KAI MakeDepositProposal, failed to get epoch switching height: %v", err)
	}
	if info.Height > int64(params.Height) {
		return nil, fmt.Errorf("KAI MakeDepositProposal, the height %d of header is lower than epoch "+
			"switching height %d", params.Height, info.Height)
	}

	if len(params.HeaderOrCrossChainMsg) == 0 {
		return nil, fmt.Errorf("you must commit the header used to verify transaction's proof and get none")
	}

	var myHeader kaiclient.FullHeader
	if err := json.Unmarshal(params.HeaderOrCrossChainMsg, &myHeader); err != nil {
		return nil, fmt.Errorf("KAI MakeDepositProposal, unmarshal cosmos header failed: %v", err)
	}
	if myHeader.Header.Height != uint64(params.Height) {
		return nil, fmt.Errorf("KAI MakeDepositProposal, "+
			"height of your header is %d not equal to %d in parameter", myHeader.Header.Height, params.Height)
	}

	if err = hskai.VerifyHeader(&myHeader, info); err != nil {
		return nil, fmt.Errorf("KAI MakeDepositProposal, failed to verify KAI header: %v", err)
	}

	sideChain, err := side_chain_manager.GetSideChain(service, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("KAI MakeDepositProposal, side_chain_manager.GetSideChain error: %v", err)
	}

	if !myHeader.Header.ValidatorsHash.Equal(myHeader.Header.NextValidatorsHash) &&
		int64(myHeader.Header.Height) > info.Height {
		hskai.PutEpochSwitchInfo(service, params.SourceChainID, &hskai.EpochSwitchInfo{
			Height:             int64(myHeader.Header.Height),
			BlockHash:          myHeader.Header.Hash().Bytes(),
			NextValidatorsHash: myHeader.Header.NextValidatorsHash.Bytes(),
			ChainID:            info.ChainID,
		})
	}

	value, err := verifyTx(myHeader.Header, service, params.Proof, params.Extra, params.SourceChainID, params.Height, sideChain)
	if err != nil {
		return nil, fmt.Errorf("KAI MakeDepositProposal, verifyFromEthTx error: %s", err)
	}

	if err := scom.CheckDoneTx(service, value.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("KAI MakeDepositProposal, check done transaction error:%s", err)
	}
	if err := scom.PutDoneTx(service, value.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("KAI MakeDepositProposal, PutDoneTx error:%s", err)
	}

	return nil, nil
}

func verifyTx(header *types.Header, native *native.NativeService, proof, extra []byte, fromChainID uint64, height uint32, sideChain *side_chain_manager.SideChain) (param *scom.MakeTxParam, err error) {
	kaiProof := new(Proof)
	err = json.Unmarshal(proof, kaiProof)
	if err != nil {
		return nil, fmt.Errorf("verifyTx, unmarshal proof error:%s", err)
	}

	if len(kaiProof.StorageProofs) != 1 {
		return nil, fmt.Errorf("verifyTx, incorrect proof format")
	}

	proofResult, err := verifyMerkleProof(kaiProof, header, sideChain.CCMCAddress)
	if err != nil {
		return nil, fmt.Errorf("verifyTx, verifyMerkleProof error:%v", err)
	}

	if proofResult == nil {
		return nil, fmt.Errorf("verifyTx, verifyMerkleProof failed")
	}

	if !checkProofResult(proofResult, extra) {
		return nil, fmt.Errorf("verifyTx, verify proof value hash failed, proof result:%x, extra:%x", proofResult, extra)
	}

	data := common.NewZeroCopySource(extra)
	txParam := new(scom.MakeTxParam)
	if err := txParam.Deserialization(data); err != nil {
		return nil, fmt.Errorf("verifyTx, deserialize merkleValue error:%s", err)
	}
	return txParam, nil
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
	Nounce   *big.Int
	Balance  *big.Int
	Storage  ecommon.Hash
	Codehash ecommon.Hash
}

func verifyMerkleProof(kaiProof *Proof, blockData *types.Header, contractAddr []byte) ([]byte, error) {
	//1. prepare verify account
	nodeList := new(light.NodeList)

	for _, s := range kaiProof.AccountProof {
		p := scom.Replace0x(s)
		nodeList.Put(nil, ecommon.Hex2Bytes(p))
	}
	ns := nodeList.NodeSet()

	addr := ecommon.Hex2Bytes(scom.Replace0x(kaiProof.Address))
	if !bytes.Equal(addr, contractAddr) {
		return nil, fmt.Errorf("verifyMerkleProof, contract address is error, proof address: %s, side chain address: %s", kaiProof.Address, hex.EncodeToString(contractAddr))
	}
	acctKey := crypto.Keccak256(addr)

	//2. verify account proof
	acctVal, err := trie.VerifyProof(ecommon.Hash(blockData.AppHash), acctKey, ns)
	if err != nil {
		return nil, fmt.Errorf("verifyMerkleProof, verify account proof error:%s", err)
	}

	nounce := new(big.Int)
	_, ok := nounce.SetString(scom.Replace0x(kaiProof.Nonce), 16)
	if !ok {
		return nil, fmt.Errorf("verifyMerkleProof, invalid format of nounce:%s", kaiProof.Nonce)
	}

	balance := new(big.Int)
	_, ok = balance.SetString(scom.Replace0x(kaiProof.Balance), 16)
	if !ok {
		return nil, fmt.Errorf("verifyMerkleProof, invalid format of balance:%s", kaiProof.Balance)
	}

	storageHash := ecommon.HexToHash(scom.Replace0x(kaiProof.StorageHash))
	codeHash := ecommon.HexToHash(scom.Replace0x(kaiProof.CodeHash))

	acct := &ProofAccount{
		Nounce:   nounce,
		Balance:  balance,
		Storage:  storageHash,
		Codehash: codeHash,
	}

	acctrlp, err := rlp.EncodeToBytes(acct)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(acctrlp, acctVal) {
		return nil, fmt.Errorf("verifyMerkleProof, verify account proof failed, wanted:%v, get:%v", acctrlp, acctVal)
	}

	//3.verify storage proof
	nodeList = new(light.NodeList)
	if len(kaiProof.StorageProofs) != 1 {
		return nil, fmt.Errorf("verifyMerkleProof, invalid storage proof format")
	}

	sp := kaiProof.StorageProofs[0]
	storageKey := crypto.Keccak256(ecommon.HexToHash(scom.Replace0x(sp.Key)).Bytes())

	for _, prf := range sp.Proof {
		nodeList.Put(nil, ecommon.Hex2Bytes(scom.Replace0x(prf)))
	}

	ns = nodeList.NodeSet()
	val, err := trie.VerifyProof(storageHash, storageKey, ns)
	if err != nil {
		return nil, fmt.Errorf("verifyMerkleProof, verify storage proof error:%s", err)
	}

	return val, nil
}

func checkProofResult(result, value []byte) bool {
	var tempBytes []byte
	err := rlp.DecodeBytes(result, &tempBytes)
	if err != nil {
		log.Errorf("checkProofResult, rlp.DecodeBytes error:%s\n", err)
		return false
	}
	//
	var s []byte
	for i := len(tempBytes); i < 32; i++ {
		s = append(s, 0)
	}
	s = append(s, tempBytes...)
	hash := crypto.Keccak256(value)
	return bytes.Equal(s, hash)
}
