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
package cosmos

import (
	"bytes"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/header_sync/cosmos"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type CosmosHandler struct{}

func NewCosmosHandler() *CosmosHandler {
	return &CosmosHandler{}
}

type CosmosProofValue struct {
	Kp    string
	Value []byte
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

func (this *CosmosHandler) MakeDepositProposal(service *native.NativeService) (*scom.MakeTxParam, error) {
	params := new(scom.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("Cosmos MakeDepositProposal, contract params deserialize error: %s", err)
	}
	info, err := cosmos.GetEpochSwitchInfo(service, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("Cosmos MakeDepositProposal, failed to get epoch switching height: %v", err)
	}
	if info.Height > int64(params.Height) {
		return nil, fmt.Errorf("Cosmos MakeDepositProposal, the height %d of header is lower than epoch "+
			"switching height %d", params.Height, info.Height)
	}

	if len(params.HeaderOrCrossChainMsg) == 0 {
		return nil, fmt.Errorf("you must commit the header used to verify transaction's proof and get none")
	}
	cdc := newCDC()
	var myHeader cosmos.CosmosHeader
	if err := cdc.UnmarshalBinaryBare(params.HeaderOrCrossChainMsg, &myHeader); err != nil {
		return nil, fmt.Errorf("Cosmos MakeDepositProposal, unmarshal cosmos header failed: %v", err)
	}
	if myHeader.Header.Height != int64(params.Height) {
		return nil, fmt.Errorf("Cosmos MakeDepositProposal, "+
			"height of your header is %d not equal to %d in parameter", myHeader.Header.Height, params.Height)
	}
	if err = cosmos.VerifyCosmosHeader(&myHeader, info); err != nil {
		return nil, fmt.Errorf("Cosmos MakeDepositProposal, failed to verify cosmos header: %v", err)
	}
	if !bytes.Equal(myHeader.Header.ValidatorsHash, myHeader.Header.NextValidatorsHash) &&
		myHeader.Header.Height > info.Height {
		cosmos.PutEpochSwitchInfo(service, params.SourceChainID, &cosmos.CosmosEpochSwitchInfo{
			Height:             myHeader.Header.Height,
			BlockHash:          myHeader.Header.Hash(),
			NextValidatorsHash: myHeader.Header.NextValidatorsHash,
			ChainID:            myHeader.Header.ChainID,
		})
	}

	var proofValue CosmosProofValue
	if err = cdc.UnmarshalBinaryBare(params.Extra, &proofValue); err != nil {
		return nil, fmt.Errorf("Cosmos MakeDepositProposal, unmarshal proof value err: %v", err)
	}
	var proof merkle.Proof
	err = cdc.UnmarshalBinaryBare(params.Proof, &proof)
	if err != nil {
		return nil, fmt.Errorf("Cosmos MakeDepositProposal, unmarshal proof err: %v", err)
	}
	if len(proofValue.Kp) != 0 {
		prt := rootmulti.DefaultProofRuntime()
		err = prt.VerifyValue(&proof, myHeader.Header.AppHash, proofValue.Kp, proofValue.Value)
		if err != nil {
			return nil, fmt.Errorf("Cosmos MakeDepositProposal, proof error: %s", err)
		}
	} else {
		prt := rootmulti.DefaultProofRuntime()
		err = prt.VerifyAbsence(&proof, myHeader.Header.AppHash, string(proofValue.Value))
		if err != nil {
			return nil, fmt.Errorf("Cosmos MakeDepositProposal, proof error: %s", err)
		}
	}
	data := common.NewZeroCopySource(proofValue.Value)
	txParam := new(scom.MakeTxParam)
	if err := txParam.Deserialization(data); err != nil {
		return nil, fmt.Errorf("Cosmos MakeDepositProposal, deserialize merkleValue error:%s", err)
	}
	if err := scom.CheckDoneTx(service, txParam.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("Cosmos MakeDepositProposal, check done transaction error:%s", err)
	}
	if err := scom.PutDoneTx(service, txParam.CrossChainID, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("Cosmos MakeDepositProposal, PutDoneTx error:%s", err)
	}
	return txParam, nil
}
