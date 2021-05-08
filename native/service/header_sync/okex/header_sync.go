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
package okex

import (
	"encoding/hex"
	"fmt"

	"bytes"

	tbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/common/log"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/header_sync/okex/ethsecp256k1"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/tendermint/tendermint/types"
)

// Handler ...
type Handler struct {
}

// NewHandler ...
func NewHandler() *Handler {
	return &Handler{}
}

// NewCDC ...
func NewCDC() *codec.Codec {
	cdc := codec.New()

	ethsecp256k1.RegisterCodec(cdc)
	return cdc
}

type CosmosHeader struct {
	Header  types.Header
	Commit  *types.Commit
	Valsets []*types.Validator
}

// SyncGenesisHeader ...
func (h *Handler) SyncGenesisHeader(native *native.NativeService) (err error) {
	param := new(hscommon.SyncGenesisHeaderParam)
	if err := param.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("CosmosHandler SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(native)
	if err != nil {
		return fmt.Errorf("CosmosHandler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("CosmosHandler SyncGenesisHeader, checkWitness error: %v", err)
	}
	// get genesis header from input parameters
	cdc := NewCDC()
	var header CosmosHeader
	err = cdc.UnmarshalBinaryBare(param.GenesisHeader, &header)
	if err != nil {
		return fmt.Errorf("CosmosHandler SyncGenesisHeader: %s", err)
	}
	// check if has genesis header
	info, err := GetEpochSwitchInfo(native, param.ChainID)
	if err == nil && info != nil {
		return fmt.Errorf("CosmosHandler SyncGenesisHeader, genesis header had been initialized")
	}
	PutEpochSwitchInfo(native, param.ChainID, &CosmosEpochSwitchInfo{
		Height:             header.Header.Height,
		NextValidatorsHash: header.Header.NextValidatorsHash,
		ChainID:            header.Header.ChainID,
		BlockHash:          header.Header.Hash(),
	})
	return nil
}

func (h *Handler) SyncBlockHeader(native *native.NativeService) error {
	params := new(hscommon.SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	cdc := NewCDC()
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
func (h *Handler) SyncCrossChainMsg(native *native.NativeService) error {
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
	BlockHash tbytes.HexBytes

	// The hash of new validators set which used to verify validators set
	// committed with proof.
	NextValidatorsHash tbytes.HexBytes

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

func VerifyCosmosHeader(myHeader *CosmosHeader, info *CosmosEpochSwitchInfo) error {
	// now verify this header
	valset := types.NewValidatorSet(myHeader.Valsets)
	if !bytes.Equal(info.NextValidatorsHash, valset.Hash()) {
		return fmt.Errorf("VerifyCosmosHeader, block validator is not right, next validator hash: %s, "+
			"validator set hash: %s", info.NextValidatorsHash.String(), hex.EncodeToString(valset.Hash()))
	}
	if !bytes.Equal(myHeader.Header.ValidatorsHash, valset.Hash()) {
		return fmt.Errorf("VerifyCosmosHeader, block validator is not right!, header validator hash: %s, "+
			"validator set hash: %s", myHeader.Header.ValidatorsHash.String(), hex.EncodeToString(valset.Hash()))
	}
	if myHeader.Commit.GetHeight() != myHeader.Header.Height {
		return fmt.Errorf("VerifyCosmosHeader, commit height is not right! commit height: %d, "+
			"header height: %d", myHeader.Commit.GetHeight(), myHeader.Header.Height)
	}
	if !bytes.Equal(myHeader.Commit.BlockID.Hash, myHeader.Header.Hash()) {
		return fmt.Errorf("VerifyCosmosHeader, commit hash is not right!, commit block hash: %s,"+
			" header hash: %s", myHeader.Commit.BlockID.Hash.String(), hex.EncodeToString(valset.Hash()))
	}
	if err := myHeader.Commit.ValidateBasic(); err != nil {
		return fmt.Errorf("VerifyCosmosHeader, commit is not right! err: %s", err.Error())
	}
	if valset.Size() != len(myHeader.Commit.Signatures) {
		return fmt.Errorf("VerifyCosmosHeader, the size of precommits is not right!")
	}
	talliedVotingPower := int64(0)
	for idx, commitSig := range myHeader.Commit.Signatures {
		if commitSig.Absent() {
			continue // OK, some precommits can be missing.
		}
		_, val := valset.GetByIndex(idx)
		// Validate signature.
		precommitSignBytes := myHeader.Commit.VoteSignBytes(info.ChainID, idx)
		if !val.PubKey.VerifyBytes(precommitSignBytes, commitSig.Signature) {
			return fmt.Errorf("VerifyCosmosHeader, Invalid commit -- invalid signature: %v", commitSig)
		}
		// Good precommit!
		if myHeader.Commit.BlockID.Equals(commitSig.BlockID(myHeader.Commit.BlockID)) {
			talliedVotingPower += val.VotingPower
		}
	}
	if talliedVotingPower <= valset.TotalVotingPower()*2/3 {
		return fmt.Errorf("VerifyCosmosHeader, voteing power is not enough!")
	}

	return nil
}
