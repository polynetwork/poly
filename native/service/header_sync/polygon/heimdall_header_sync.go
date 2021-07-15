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
package polygon

import (
	"encoding/hex"
	"fmt"

	"bytes"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/common/log"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"
	polygonTypes "github.com/polynetwork/poly/native/service/header_sync/polygon/types"
	polygonCmn "github.com/polynetwork/poly/native/service/header_sync/polygon/types/common"
	"github.com/polynetwork/poly/native/service/utils"
)

type HeimdallHandler struct {
}

// NewHeimdallHandler ...
func NewHeimdallHandler() *HeimdallHandler {
	return &HeimdallHandler{}
}

type CosmosHeader struct {
	Header  polygonTypes.Header
	Commit  *polygonTypes.Commit
	Valsets []*polygonTypes.Validator
}

// SyncGenesisHeader ...
func (h *HeimdallHandler) SyncGenesisHeader(native *native.NativeService) (err error) {
	param := new(hscommon.SyncGenesisHeaderParam)
	if err := param.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("HeimdallHandler SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return fmt.Errorf("HeimdallHandler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("HeimdallHandler SyncGenesisHeader, checkWitness error: %v", err)
	}
	// get genesis header from input parameters
	cdc := polygonTypes.NewCDC()
	var header CosmosHeader
	err = cdc.UnmarshalBinaryBare(param.GenesisHeader, &header)
	if err != nil {
		return fmt.Errorf("HeimdallHandler SyncGenesisHeader: %s", err)
	}
	// check if has genesis header
	info, err := GetEpochSwitchInfo(native, param.ChainID)
	if err == nil && info != nil {
		return fmt.Errorf("HeimdallHandler SyncGenesisHeader, genesis header had been initialized")
	}
	PutEpochSwitchInfo(native, param.ChainID, &CosmosEpochSwitchInfo{
		Height:             header.Header.Height,
		NextValidatorsHash: header.Header.NextValidatorsHash,
		ChainID:            header.Header.ChainID,
		BlockHash:          header.Header.Hash(),
	})
	return nil
}

func (h *HeimdallHandler) SyncBlockHeader(native *native.NativeService) error {
	params := new(hscommon.SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	cdc := polygonTypes.NewCDC()
	cnt := 0
	info, err := GetEpochSwitchInfo(native, params.ChainID)
	if err != nil {
		return fmt.Errorf("SyncBlockHeader, get epoch switching height failed: %v", err)
	}
	for _, v := range params.Headers {
		var myHeader CosmosHeader
		err := cdc.UnmarshalBinaryBare(v, &myHeader)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader failed to unmarshal header: %v", err)
		}
		if bytes.Equal(myHeader.Header.NextValidatorsHash, myHeader.Header.ValidatorsHash) {
			continue
		}
		if info.Height >= myHeader.Header.Height {
			log.Debugf("SyncBlockHeader, height %d is lower or equal than epoch switching height %d",
				myHeader.Header.Height, info.Height)
			continue
		}
		if err = VerifyCosmosHeader(&myHeader, info); err != nil {
			return fmt.Errorf("SyncBlockHeader, failed to verify header: %v", err)
		}
		info.NextValidatorsHash = myHeader.Header.NextValidatorsHash
		info.Height = myHeader.Header.Height
		info.BlockHash = myHeader.Header.Hash()
		cnt++
	}
	if cnt == 0 {
		return fmt.Errorf("no header you commited is useful")
	}
	PutEpochSwitchInfo(native, params.ChainID, info)
	return nil
}

// SyncCrossChainMsg ...
func (h *HeimdallHandler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}

func GetEpochSwitchInfo(service *native.NativeService, chainId uint64) (*CosmosEpochSwitchInfo, error) {
	val, err := service.GetCacheDB().Get(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(hscommon.EPOCH_SWITCH), utils.GetUint64Bytes(chainId)))
	if err != nil {
		return nil, fmt.Errorf("failed to get epoch switching height: %v", err)
	}
	raw, err := cstates.GetValueFromRawStorageItem(val)
	if err != nil {
		return nil, fmt.Errorf("deserialize bytes from raw storage item err: %v", err)
	}
	info := &CosmosEpochSwitchInfo{}
	if err = info.Deserialization(common.NewZeroCopySource(raw)); err != nil {
		return nil, fmt.Errorf("failed to deserialize CosmosEpochSwitchInfo: %v", err)
	}
	return info, nil
}

func PutEpochSwitchInfo(service *native.NativeService, chainId uint64, info *CosmosEpochSwitchInfo) {
	sink := common.NewZeroCopySink(nil)
	info.Serialization(sink)
	service.GetCacheDB().Put(
		utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(hscommon.EPOCH_SWITCH), utils.GetUint64Bytes(chainId)),
		cstates.GenRawStorageItem(sink.Bytes()))
	notifyEpochSwitchInfo(service, chainId, info)
}

func notifyEpochSwitchInfo(native *native.NativeService, chainID uint64, info *CosmosEpochSwitchInfo) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States: []interface{}{chainID, info.BlockHash.String(), info.Height,
				info.NextValidatorsHash.String(), info.ChainID, native.GetHeight()},
		})
}

type CosmosEpochSwitchInfo struct {
	// The height where validators set changed last time. Poly only accept
	// header and proof signed by new validators. That means the header
	// can not be lower than this height.
	Height int64

	// Hash of the block at `Height`. Poly don't save the whole header.
	// So we can identify the content of this block by `BlockHash`.
	BlockHash polygonCmn.HexBytes

	// The hash of new validators set which used to verify validators set
	// committed with proof.
	NextValidatorsHash polygonCmn.HexBytes

	// The cosmos chain-id of this chain basing Cosmos-sdk.
	ChainID string
}

func (info *CosmosEpochSwitchInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteInt64(info.Height)
	sink.WriteVarBytes(info.BlockHash)
	sink.WriteVarBytes(info.NextValidatorsHash)
	sink.WriteString(info.ChainID)
}

func (info *CosmosEpochSwitchInfo) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	info.Height, eof = source.NextInt64()
	if eof {
		return fmt.Errorf("deserialize height of CosmosEpochSwitchInfo failed")
	}
	info.BlockHash, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("deserialize BlockHash of CosmosEpochSwitchInfo failed")
	}
	info.NextValidatorsHash, eof = source.NextVarBytes()
	if eof {
		return fmt.Errorf("deserialize NextValidatorsHash of CosmosEpochSwitchInfo failed")
	}
	info.ChainID, eof = source.NextString()
	if eof {
		return fmt.Errorf("deserialize ChainID of CosmosEpochSwitchInfo failed")
	}
	return nil
}

func VerifySpan(native *native.NativeService, heimdallPolyChainID uint64, proof *CosmosProof) (span *Span, err error) {
	info, err := GetEpochSwitchInfo(native, heimdallPolyChainID)
	if err != nil {
		err = fmt.Errorf("HeimdallHandler failed to get epoch switching height: %v", err)
		return
	}

	if err = VerifyCosmosHeader(&proof.Header, info); err != nil {
		return nil, fmt.Errorf("HeimdallHandler failed to verify cosmos header: %v", err)
	}

	if len(proof.Proof.Ops) != 2 {
		err = fmt.Errorf("proof size wrong")
		return
	}
	if !bytes.Equal(proof.Proof.Ops[1].Key, []byte("bor")) {
		err = fmt.Errorf("wrong module for proof")
		return
	}

	prt := rootmulti.DefaultProofRuntime()

	err = prt.VerifyValue(&proof.Proof, proof.Header.Header.AppHash, proof.Value.Kp, proof.Value.Value)
	if err != nil {
		err = fmt.Errorf("validateHeaderExtraField VerifyValue error: %s", err)
		return
	}

	heimdallSpan := &polygonTypes.HeimdallSpan{}
	err = polygonTypes.NewCDC().UnmarshalBinaryBare(proof.Value.Value, heimdallSpan)
	if err != nil {
		err = fmt.Errorf("validateHeaderExtraField heimdallSpan UnmarshalBinaryBare error: %s", err)
		return
	}

	span, err = SpanFromHeimdall(heimdallSpan)
	return
}

func VerifyCosmosHeader(myHeader *CosmosHeader, info *CosmosEpochSwitchInfo) error {
	// now verify this header
	valset := polygonTypes.NewValidatorSet(myHeader.Valsets)
	if !bytes.Equal(info.NextValidatorsHash, valset.Hash()) {
		return fmt.Errorf("VerifyCosmosHeader, block validator is not right, next validator hash: %s, "+
			"validator set hash: %s", info.NextValidatorsHash.String(), hex.EncodeToString(valset.Hash()))
	}
	if !bytes.Equal(myHeader.Header.ValidatorsHash, valset.Hash()) {
		return fmt.Errorf("VerifyCosmosHeader, block validator is not right!, header validator hash: %s, "+
			"validator set hash: %s", myHeader.Header.ValidatorsHash.String(), hex.EncodeToString(valset.Hash()))
	}
	if myHeader.Commit.Height() != myHeader.Header.Height {
		return fmt.Errorf("VerifyCosmosHeader, commit height is not right! commit height: %d, "+
			"header height: %d", myHeader.Commit.Height(), myHeader.Header.Height)
	}
	if !bytes.Equal(myHeader.Commit.BlockID.Hash, myHeader.Header.Hash()) {
		return fmt.Errorf("VerifyCosmosHeader, commit hash is not right!, commit block hash: %s,"+
			" header hash: %s", myHeader.Commit.BlockID.Hash.String(), hex.EncodeToString(valset.Hash()))
	}
	if err := myHeader.Commit.ValidateBasic(); err != nil {
		return fmt.Errorf("VerifyCosmosHeader, commit is not right! err: %s", err.Error())
	}
	if valset.Size() != myHeader.Commit.Size() {
		return fmt.Errorf("VerifyCosmosHeader, the size of precommits is not right!")
	}
	talliedVotingPower := int64(0)
	for _, commitSig := range myHeader.Commit.Precommits {
		if commitSig == nil {
			continue
		}
		idx := commitSig.ValidatorIndex
		_, val := valset.GetByIndex(idx)
		if val == nil {
			return fmt.Errorf("VerifyCosmosHeader, validator %d doesn't exist!", idx)
		}
		if commitSig.Type != polygonTypes.PrecommitType {
			return fmt.Errorf("VerifyCosmosHeader, commitSig.Type(%d) wrong", commitSig.Type)
		}
		// Validate signature.
		precommitSignBytes := myHeader.Commit.VoteSignBytes(info.ChainID, idx)
		if !val.PubKey.VerifyBytes(precommitSignBytes, commitSig.Signature) {
			return fmt.Errorf("VerifyCosmosHeader, Invalid commit -- invalid signature: %v", commitSig)
		}
		// Good precommit!
		if myHeader.Commit.BlockID.Equals(commitSig.BlockID) {
			talliedVotingPower += val.VotingPower
		}
	}
	if talliedVotingPower <= valset.TotalVotingPower()*2/3 {
		return fmt.Errorf("VerifyCosmosHeader, voteing power is not enough!")
	}

	return nil
}
