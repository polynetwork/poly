/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package cosmos

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/log"
	"github.com/ontio/multi-chain/native"
	hscommon "github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/types"
	mctypes "github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/native/service/utils"
)

type CosmosHandler struct {
}

type CosmosHeader struct {
	Header     types.Header
	Commit     *types.Commit
	Valsets    []*types.Validator
}

func NewCosmosHandler() *CosmosHandler {
	return &CosmosHandler{}
}

func newCDC() *codec.Codec {
	cdc := codec.New()

	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	cdc.RegisterConcrete(ed25519.PubKeyEd25519{}, ed25519.PubKeyAminoName, nil)
	cdc.RegisterConcrete(secp256k1.PubKeySecp256k1{}, secp256k1.PubKeyAminoName, nil)
	cdc.RegisterConcrete(multisig.PubKeyMultisigThreshold{}, multisig.PubKeyMultisigThresholdAminoRoute, nil)

	cdc.RegisterInterface((*crypto.PrivKey)(nil), nil)
	cdc.RegisterConcrete(ed25519.PrivKeyEd25519{}, ed25519.PrivKeyAminoName, nil)
	cdc.RegisterConcrete(secp256k1.PrivKeySecp256k1{}, secp256k1.PrivKeyAminoName, nil)
	return cdc
}

func (this *CosmosHandler) SyncGenesisHeader(native *native.NativeService) error {
	params := new(hscommon.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	// get operator from database
	operatorAddress, err := mctypes.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return fmt.Errorf("CosmosHandler SyncGenesisHeader, checkWitness error: %v", err)
	}
	// get genesis header from input parameters
	cdc := newCDC()
	var header CosmosHeader
	err = cdc.UnmarshalBinaryBare(params.GenesisHeader, &header)
	if err != nil {
		return fmt.Errorf("CosmosHandler SyncGenesisHeader: %s", err)
	}
	// check if has genesis header
	has, err := hasGenesis(native, params.ChainID)
	if err != nil {
		return fmt.Errorf("CosmosHandler SyncGenesisHeader, hasGenesis error: %v", err)
	}
	if has {
		return fmt.Errorf("CosmosHandler SyncGenesisHeader, genesis header had been initialized")
	}
	// now we can init genesis
	err = putGenesisHeader(native, cdc, &header, params.ChainID)
	if err != nil {
		return fmt.Errorf("CosmosHandler SyncGenesisHeader, put blockHeader error: %v", err)
	}
	keyHeights := new(KeyHeights)
	keyHeights.AddNewHeight(header.Header.Height)
	PutKeyHeights(native, params.ChainID, keyHeights)
	return nil
}

func (this *CosmosHandler) SyncBlockHeader(native *native.NativeService) error {
	params := new(hscommon.SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	cdc := newCDC()
	for _, v := range params.Headers {
		var myHeader CosmosHeader
		err := cdc.UnmarshalBinaryBare(v, &myHeader)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader: %s", err)
		}
		// Check if this header is exist
		//
		height := myHeader.Header.Height
		_, err = GetHeaderByHeight(native, cdc, height, params.ChainID)
		if err == nil {
			log.Warnf("SyncBlockHeader, this header has synced.")
			continue
		}
		keyHeights, err := GetKeyHeights(native, params.ChainID)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, GetKeyHeights error:%v", err)
		}
		keyHeight, err := keyHeights.FindKeyHeight(height)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, FindKeyHeight error:%v", err)
		}
		prevHeader, err := GetHeaderByHeight(native, cdc, keyHeight, params.ChainID)
		if err != nil {
			return fmt.Errorf("SyncBlockHeader, get prev header error: %v", err)
		}
		// now verify this header
		valHash := prevHeader.Header.NextValidatorsHash
		valset := types.NewValidatorSet(myHeader.Valsets)
		header := myHeader.Header
		commit := myHeader.Commit
		if bytes.Equal(valHash, valset.Hash()) != true {
			return fmt.Errorf("block validator is not right!, next validator hash: %s, validator set hash: %s", valHash.String(), hex.EncodeToString(valset.Hash()))
		}
		if bytes.Equal(header.ValidatorsHash, valset.Hash()) != true {
			return fmt.Errorf("block validator is not right!, header validator hash: %s, validator set hash: %s", header.ValidatorsHash.String(), hex.EncodeToString(valset.Hash()))
		}
		if commit.GetHeight() != header.Height {
			return fmt.Errorf("commit height is not right! commit height: %d, header height: %d", commit.GetHeight(), header.Height)
		}
		if bytes.Equal(commit.BlockID.Hash, header.Hash()) != true {
			return fmt.Errorf("commit hash is not right!, commit block hash: %s, header hash: %s", commit.BlockID.Hash.String(), hex.EncodeToString(valset.Hash()))
		}
		if err := commit.ValidateBasic(); err != nil {
			return fmt.Errorf("commit is not right! err: %s", err.Error())
		}
		if valset.Size() != len(commit.Signatures) {
			return fmt.Errorf("the size of precommits is not right!")
		}
		talliedVotingPower := int64(0)
		for idx, commitSig := range commit.Signatures {
			if commitSig.Absent() {
				continue // OK, some precommits can be missing.
			}
			_, val := valset.GetByIndex(idx)
			// Validate signature.
			precommitSignBytes := commit.VoteSignBytes(prevHeader.Header.ChainID, idx)
			if !val.PubKey.VerifyBytes(precommitSignBytes, commitSig.Signature) {
				return fmt.Errorf("Invalid commit -- invalid signature: %v", commitSig)
			}
			// Good precommit!
			if commit.BlockID.Equals(commitSig.BlockID(commit.BlockID)) {
				talliedVotingPower += val.VotingPower
			}
		}

		if talliedVotingPower <= valset.TotalVotingPower()*2/3 {
			return fmt.Errorf("voteing power is not enough!")
		}
		// update
		putHeader(native, cdc, params.ChainID, &myHeader)
		keyHeights.AddNewHeight(myHeader.Header.Height)
		PutKeyHeights(native, params.ChainID, keyHeights)
	}
	return nil
}

func (this *CosmosHandler) SyncCrossChainMsg(native *native.NativeService) error {
	return nil
}
