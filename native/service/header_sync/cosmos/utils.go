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

package cosmos

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/tendermint/tendermint/types"
)

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
