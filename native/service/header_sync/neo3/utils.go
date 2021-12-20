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

package neo3

import (
	"fmt"
	"github.com/joeqian10/neo3-gogogo/crypto"
	"github.com/joeqian10/neo3-gogogo/sc"
	"github.com/joeqian10/neo3-gogogo/tx"
	"github.com/polynetwork/poly/common"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/neo3_state_manager"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

//verify header of any height
//find key height and get neoconsensus first, then check the witness
func verifyHeader(native *native.NativeService, chainID uint64, header *NeoBlockHeader, magic uint32) error {
	neoConsensus, err := getConsensusValByChainId(native, chainID)
	if err != nil {
		return fmt.Errorf("verifyHeader, get Consensus error: %s", err)
	}
	if !neoConsensus.NextConsensus.Equals(header.Witness.GetScriptHash()) {
		return fmt.Errorf("verifyHeader, invalid script hash in header error, expected: %s, got: %s", neoConsensus.NextConsensus.String(), header.Witness.GetScriptHash().String())
	}
	msg, err := header.GetMessage(magic)
	if err != nil {
		return fmt.Errorf("verifyHeader, unable to get hash data of header")
	}
	// verify witness
	if verified := tx.VerifyMultiSignatureWitness(msg, header.Witness); !verified {
		return fmt.Errorf("verifyHeader, VerifyMultiSignatureWitness error: %s, height: %d", err, header.GetIndex())
	}
	return nil
}

func VerifyCrossChainMsgSig(native *native.NativeService, magic uint32, crossChainMsg *NeoCrossChainMsg) error {
	// get neo3 state validator from native contract
	svListBytes, err := neo3_state_manager.GetCurrentStateValidator(native)
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, neo3_state_manager.GetCurrentStateValidator error: %v", err)
	}
	svStrings, err := neo3_state_manager.DeserializeStringArray(svListBytes)
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, neo3_state_manager.DeserializeStringArray error: %v", err)
	}
	pubKeys := make([]crypto.ECPoint, len(svStrings), len(svStrings))
	for i, v := range svStrings {
		pubKey, err := crypto.NewECPointFromString(v)
		if err != nil {
			return fmt.Errorf("verifyCrossChainMsg, crypto.NewECPointFromString error: %v", err)
		}
		pubKeys[i] = *pubKey
	}
	n := len(pubKeys)
	m := n - (n-1)/3
	msc, err := sc.CreateMultiSigContract(m, pubKeys) // sort public keys inside
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, sc.CreateMultiSigContract error: %v", err)
	}
	expected := msc.GetScriptHash()
	got, err := crossChainMsg.GetScriptHash()
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, getScripthash error: %v", err)
	}
	// compare state validator
	if !expected.Equals(got) {
		return fmt.Errorf("verifyCrossChainMsg, invalid script hash in NeoCrossChainMsg error, expected: %s, got: %s", expected.String(), got.String())
	}
	msg, err := crossChainMsg.GetMessage(magic)
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, unable to get unsigned message of neo crossChainMsg")
	}
	// verify witness
	if len(crossChainMsg.Witnesses) == 0 {
		return fmt.Errorf("verifyCrossChainMsg, incorrect witness length")
	}
	invScript, err := crypto.Base64Decode(crossChainMsg.Witnesses[0].Invocation)
	if err != nil {
		return fmt.Errorf("crypto.Base64Decode, decode invocation script error: %v", err)
	}
	verScript, err := crypto.Base64Decode(crossChainMsg.Witnesses[0].Verification)
	if err != nil {
		return fmt.Errorf("crypto.Base64Decode, decode verification script error: %v", err)
	}
	witness := &tx.Witness{
		InvocationScript:   invScript,
		VerificationScript: verScript,
	}
	v1 := tx.VerifyMultiSignatureWitness(msg, witness)
	if !v1 {
		return fmt.Errorf("verifyCrossChainMsg, verify witness failed, height: %d", crossChainMsg.Index)
	}
	return nil
}

func getConsensusValByChainId(native *native.NativeService, chainID uint64) (*NeoConsensus, error) {
	contract := utils.HeaderSyncContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	neoConsensusStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(hscommon.CONSENSUS_PEER), chainIDBytes))
	if err != nil {
		return nil, fmt.Errorf("getNextConsensusByHeight, get nextConsensusStore error: %v", err)
	}
	if neoConsensusStore == nil {
		return nil, fmt.Errorf("getNextConsensusByHeight, can not find any record")
	}
	neoConsensusBytes, err := cstates.GetValueFromRawStorageItem(neoConsensusStore)
	if err != nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, deserialize from raw storage item err: %v", err)
	}
	neoConsensus := new(NeoConsensus)
	if err := neoConsensus.Deserialization(common.NewZeroCopySource(neoConsensusBytes)); err != nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, deserialize consensusPeer error: %v", err)
	}
	return neoConsensus, nil
}

func putConsensusValByChainId(native *native.NativeService, neoConsensus *NeoConsensus) error {
	contract := utils.HeaderSyncContractAddress
	sink := common.NewZeroCopySink(nil)
	neoConsensus.Serialization(sink)
	chainIDBytes := utils.GetUint64Bytes(neoConsensus.ChainID)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(hscommon.CONSENSUS_PEER), chainIDBytes), cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}
