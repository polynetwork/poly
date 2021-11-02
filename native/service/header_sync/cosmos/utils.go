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
	"reflect"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	hscommon "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"

	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"
	"github.com/tendermint/tendermint/types"
)

var Cdc *amino.Codec

func init() {
	Cdc := amino.NewCodec()

	Cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	Cdc.RegisterConcrete(sr25519.PubKey{},
		sr25519.PubKeyName, nil)
	Cdc.RegisterConcrete(&ed25519.PubKey{},
		ed25519.PubKeyName, nil)
	Cdc.RegisterConcrete(&secp256k1.PubKey{},
		secp256k1.PubKeyName, nil)

	Cdc.RegisterInterface((*crypto.PrivKey)(nil), nil)
	Cdc.RegisterConcrete(sr25519.PrivKey{},
		sr25519.PrivKeyName, nil)
	Cdc.RegisterConcrete(&ed25519.PrivKey{},
		ed25519.PrivKeyName, nil)
	Cdc.RegisterConcrete(&secp256k1.PrivKey{},
		secp256k1.PrivKeyName, nil)

	Cdc.Seal()
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
	valSetHash := HashCosmosValSet(valset, myHeader.Header.Version.Block)
	if !bytes.Equal(info.NextValidatorsHash, valSetHash) {
		if !bytes.Equal(info.NextValidatorsHash, aminoValSetHash(valset)) {
			return fmt.Errorf("VerifyCosmosHeader, block validator is not right, next validator hash: %s, "+
				"validator set hash: %s", info.NextValidatorsHash.String(), hex.EncodeToString(valSetHash))
		}
	}
	if !bytes.Equal(myHeader.Header.ValidatorsHash, valSetHash) {
		return fmt.Errorf("VerifyCosmosHeader, block validator is not right!, header validator hash: %s, "+
			"validator set hash: %s", myHeader.Header.ValidatorsHash.String(), hex.EncodeToString(valSetHash))
	}
	if myHeader.Commit.GetHeight() != myHeader.Header.Height {
		return fmt.Errorf("VerifyCosmosHeader, commit height is not right! commit height: %d, "+
			"header height: %d", myHeader.Commit.GetHeight(), myHeader.Header.Height)
	}
	if !bytes.Equal(myHeader.Commit.BlockID.Hash, HashCosmosHeader(myHeader.Header)) {
		return fmt.Errorf("VerifyCosmosHeader, commit hash is not right!, commit block hash: %s,"+
			" header hash: %s", myHeader.Commit.BlockID.Hash.String(), hex.EncodeToString(valSetHash))
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
		_, val := valset.GetByIndex(int32(idx))
		// Validate signature.
		precommitSignBytes := myHeader.Commit.VoteSignBytes(info.ChainID, int32(idx))
		if !val.PubKey.VerifySignature(precommitSignBytes, commitSig.Signature) {
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

// HashCosmosHeader supports hashing both pre and post stargate tendermint block headers
func HashCosmosHeader(header types.Header) []byte {
	// hash encoding changed from amino to protobuf on block version 11, tm v0.34:
	// https://github.com/tendermint/tendermint/pull/5173
	if header.Version.Block <= 10 {
		return aminoHeaderHash(&header)
	}
	return header.Hash()
}

// HashCosmosHeader supports hashing both pre and post stargate validator sets
func HashCosmosValSet(valSet *types.ValidatorSet, blockVersion uint64) []byte {
	// hash encoding changed from amino to protobuf on block version 11, tm v0.34:
	// https://github.com/tendermint/tendermint/pull/5173
	if blockVersion <= 10 {
		return aminoValSetHash(valSet)
	}
	return valSet.Hash()
}

// See: https://github.com/tendermint/tendermint/blob/v0.33.7/types/block.go#L395
func aminoHeaderHash(h *types.Header) []byte {
	if h == nil || len(h.ValidatorsHash) == 0 {
		return nil
	}
	return merkle.HashFromByteSlices([][]byte{
		cdcEncode(h.Version),
		cdcEncode(h.ChainID),
		cdcEncode(h.Height),
		cdcEncode(h.Time),
		cdcEncode(h.LastBlockID),
		cdcEncode(h.LastCommitHash),
		cdcEncode(h.DataHash),
		cdcEncode(h.ValidatorsHash),
		cdcEncode(h.NextValidatorsHash),
		cdcEncode(h.ConsensusHash),
		cdcEncode(h.AppHash),
		cdcEncode(h.LastResultsHash),
		cdcEncode(h.EvidenceHash),
		cdcEncode(h.ProposerAddress),
	})
}

// See: https://github.com/tendermint/tendermint/blob/v0.33.7/types/validator_set.go#L316
// and: https://github.com/tendermint/tendermint/blob/v0.33.7/types/validator.go#L90
func aminoValSetHash(valSet *types.ValidatorSet) []byte {
	bzs := make([][]byte, len(valSet.Validators))
	for i, val := range valSet.Validators {
		bzs[i] = cdcEncode(struct {
			PubKey      crypto.PubKey
			VotingPower int64
		}{
			val.PubKey,
			val.VotingPower,
		})
	}
	return merkle.HashFromByteSlices(bzs)
}

// See: https://github.com/tendermint/tendermint/blob/v0.33.7/types/encoding_helper.go#L5
func cdcEncode(item interface{}) []byte {
	if item != nil && !isTypedNil(item) && !isEmpty(item) {
		return Cdc.MustMarshalBinaryBare(item)
	}
	return nil
}

// See: https://github.com/tendermint/tendermint/blob/v0.33.7/types/utils.go#L10
func isTypedNil(o interface{}) bool {
	rv := reflect.ValueOf(o)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// See: https://github.com/tendermint/tendermint/blob/v0.33.7/types/utils.go#L20
// Returns true if it has zero length.
func isEmpty(o interface{}) bool {
	rv := reflect.ValueOf(o)
	switch rv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	default:
		return false
	}
}
