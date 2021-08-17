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

package neo

import (
	"encoding/hex"
	"fmt"
	"github.com/joeqian10/neo-gogogo/tx"
	"github.com/polynetwork/poly/common"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

//verify header of any height
//find key height and get neoconsensus first, then check the witness
func verifyHeader(native *native.NativeService, chainID uint64, header *NeoBlockHeader) error {
	neoConsensus, err := getConsensusValByChainId(native, chainID)
	if err != nil {
		return fmt.Errorf("verifyHeader, get Consensus error: %s", err)
	}
	if neoConsensus.NextConsensus != header.Witness.GetScriptHash() {
		return fmt.Errorf("verifyHeader, invalid script hash in header error, expected: %s, got: %s", neoConsensus.NextConsensus.String(), header.Witness.GetScriptHash().String())
	}

	msg, err := header.GetMessage()
	if err != nil {
		return fmt.Errorf("verifyHeader, unable to get hash data of header")
	}
	if verified := tx.VerifyMultiSignatureWitness(msg, header.Witness); !verified {
		return fmt.Errorf("verifyHeader, VerifyMultiSignatureWitness error: %s, height: %d", err, header.Index)
	}
	return nil
}

func VerifyCrossChainMsgSig(native *native.NativeService, chainID uint64, crossChainMsg *NeoCrossChainMsg) error {
	neoConsensus, err := getConsensusValByChainId(native, chainID)
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, get ConsensusPeer error: %v", err)
	}
	crossChainMsgConsensus, err := crossChainMsg.GetScriptHash()
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, getScripthash error: %v", err)
	}
	if neoConsensus.NextConsensus != crossChainMsgConsensus {
		return fmt.Errorf("verifyCrossChainMsg, invalid script hash in NeoCrossChainMsg error, expected: %s, got: %s", neoConsensus.NextConsensus.String(), crossChainMsgConsensus.String())
	}
	msg, err := crossChainMsg.GetMessage()
	if err != nil {
		return fmt.Errorf("verifyCrossChainMsg, unable to get unsigned message of neo crossChainMsg")
	}
	invScript, _ := hex.DecodeString(crossChainMsg.Witness.InvocationScript)
	verScript, _ := hex.DecodeString(crossChainMsg.Witness.VerificationScript)
	witness := &tx.Witness{
		InvocationScript:   invScript,
		VerificationScript: verScript,
	}
	if verified := tx.VerifyMultiSignatureWitness(msg, witness); !verified {
		return fmt.Errorf("verifyCrossChainMsg, VerifyMultiSignatureWitness error: %s, height: %d", "verification failed", crossChainMsg.Index)
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
