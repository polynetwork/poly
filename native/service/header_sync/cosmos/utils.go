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

	tm34crypto "github.com/switcheo/tendermint/crypto"
	tm34ed25519 "github.com/switcheo/tendermint/crypto/ed25519"
	tm34secp256k1 "github.com/switcheo/tendermint/crypto/secp256k1"
	tm34sr25519 "github.com/switcheo/tendermint/crypto/sr25519"
	tm34bytes "github.com/switcheo/tendermint/libs/bytes"
	tm34version "github.com/switcheo/tendermint/proto/tendermint/version"
	tm34types "github.com/switcheo/tendermint/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"
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
	valSetHash := HashCosmosValSet(valset, myHeader.Header.Version.Block.Uint64())
	if !bytes.Equal(info.NextValidatorsHash, valSetHash) {
		// recheck legacy hash to allow for upgrade block to pass (info has old hash format, but header has new block version)
		if !bytes.Equal(info.NextValidatorsHash, valset.Hash()) {
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
		_, val := valset.GetByIndex(idx)
		// Validate signature.
		precommitSignBytes := VoteSignBytes(myHeader, idx)
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

// HashCosmosHeader supports hashing both pre and post stargate tendermint block headers
func HashCosmosHeader(header types.Header) []byte {
	// hash encoding changed from amino to protobuf on block version 11, tm v0.34:
	// https://github.com/tendermint/tendermint/pull/5173
	if header.Version.Block < 11 {
		return header.Hash() // legacy hash
	}
	// convert header to new type
	h := tm34types.Header{
		Version: tm34version.Consensus{
			Block: uint64(header.Version.Block),
			App:   uint64(header.Version.App),
		},
		ChainID: header.ChainID,
		Height:  header.Height,
		Time:    header.Time,
		LastBlockID: tm34types.BlockID{
			Hash: tm34bytes.HexBytes(header.LastBlockID.Hash),
			PartSetHeader: tm34types.PartSetHeader{
				Total: uint32(header.LastBlockID.PartsHeader.Total),
				Hash:  tm34bytes.HexBytes(header.LastBlockID.PartsHeader.Hash),
			},
		},
		LastCommitHash:     tm34bytes.HexBytes(header.LastCommitHash),
		DataHash:           tm34bytes.HexBytes(header.DataHash),
		ValidatorsHash:     tm34bytes.HexBytes(header.ValidatorsHash),
		NextValidatorsHash: tm34bytes.HexBytes(header.NextValidatorsHash),
		ConsensusHash:      tm34bytes.HexBytes(header.ConsensusHash),
		AppHash:            tm34bytes.HexBytes(header.AppHash),
		LastResultsHash:    tm34bytes.HexBytes(header.LastResultsHash),
		EvidenceHash:       tm34bytes.HexBytes(header.EvidenceHash),
		ProposerAddress:    tm34bytes.HexBytes(header.ProposerAddress),
	}
	// use new type's hash that hashes protobuf bytes instead of amino bytes
	return h.Hash()
}

// HashCosmosHeader supports hashing both pre and post stargate validator sets
func HashCosmosValSet(valSet *types.ValidatorSet, blockVersion uint64) []byte {
	// hash encoding changed from amino to protobuf on block version 11, tm v0.34:
	// https://github.com/tendermint/tendermint/pull/5173
	if blockVersion < 11 {
		return valSet.Hash() // legacy hash
	}
	// convert vals and valset to new type
	vals := make([]*tm34types.Validator, valSet.Size())
	for i, v := range valSet.Validators {
		var pubKey tm34crypto.PubKey
		switch pk := v.PubKey.(type) {
		case sr25519.PubKeySr25519:
			pubKey = tm34sr25519.PubKey(pk[:])
		case ed25519.PubKeyEd25519:
			pubKey = tm34ed25519.PubKey(pk[:])
		case secp256k1.PubKeySecp256k1:
			pubKey = tm34secp256k1.PubKey(pk[:])
		default:
			panic(fmt.Sprintf("Unknown pubkey type: %x", v.PubKey))
		}
		vals[i] = tm34types.NewValidator(pubKey, v.VotingPower)
		vals[i].ProposerPriority = v.ProposerPriority
	}
	vs := tm34types.NewValidatorSet(vals)
	// use new type's hash that hashes protobuf bytes instead of amino bytes
	return vs.Hash()
}

func VoteSignBytes(header *CosmosHeader, valIdx int) []byte {
	// hash encoding changed from amino to protobuf on block version 11, tm v0.34:
	// https://github.com/tendermint/tendermint/pull/5173
	if header.Header.Version.Block < 11 {
		return header.Commit.VoteSignBytes(header.Header.ChainID, valIdx) // legacy hash
	}
	// convert commit to new type
	sigs := make([]tm34types.CommitSig, len(header.Commit.Signatures))
	for i, v := range header.Commit.Signatures {
		sigs[i] = tm34types.CommitSig{
			BlockIDFlag:      tm34types.BlockIDFlag(v.BlockIDFlag),
			ValidatorAddress: tm34bytes.HexBytes(v.ValidatorAddress),
			Timestamp:        v.Timestamp,
			Signature:        v.Signature,
		}
	}
	commit := tm34types.NewCommit(
		header.Commit.Height,
		int32(header.Commit.Round),
		tm34types.BlockID{
			Hash: tm34bytes.HexBytes(header.Commit.BlockID.Hash),
			PartSetHeader: tm34types.PartSetHeader{
				Total: uint32(header.Commit.BlockID.PartsHeader.Total),
				Hash:  tm34bytes.HexBytes(header.Commit.BlockID.PartsHeader.Hash),
			},
		},
		sigs,
	)
	return commit.VoteSignBytes(header.Header.ChainID, int32(valIdx))
}
