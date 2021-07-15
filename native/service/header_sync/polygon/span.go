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
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	polygonTypes "github.com/polynetwork/poly/native/service/header_sync/polygon/types"
)

type Span struct {
	ID                uint64       `protobuf:"varint,1,opt,name=id,proto3" json:"id" yaml:"id"`
	StartBlock        uint64       `protobuf:"varint,2,opt,name=start_block,json=startBlock,proto3" json:"start_block" yaml:"start_block"`
	EndBlock          uint64       `protobuf:"varint,3,opt,name=end_block,json=endBlock,proto3" json:"end_block" yaml:"end_block"`
	ValidatorSet      ValidatorSet `protobuf:"bytes,4,opt,name=validator_set,json=validatorSet,proto3" json:"validator_set" yaml:"validator_set"`
	SelectedProducers []Validator  `protobuf:"bytes,5,rep,name=selected_producers,json=selectedProducers,proto3" json:"selected_producers" yaml:"selected_producers"`
	BorChainId        string       `protobuf:"bytes,6,opt,name=bor_chain_id,json=borChainId,proto3" json:"bor_chain_id" yaml:"bor_chain_id"`
}

func SpanFromHeimdall(hs *polygonTypes.HeimdallSpan) (span *Span, err error) {
	span = &Span{
		ID:         hs.ID,
		StartBlock: hs.StartBlock,
		EndBlock:   hs.EndBlock,
		BorChainId: hs.BorChainId,
	}

	var bp Validator
	for _, hp := range hs.SelectedProducers {
		bp, err = ValidatorFromHeimdall(&hp)
		if err != nil {
			return
		}
		span.SelectedProducers = append(span.SelectedProducers, bp)
	}

	span.ValidatorSet, err = ValidatorSetFromHeimdall(&hs.ValidatorSet)
	return
}

func ValidatorSetFromHeimdall(hvs *polygonTypes.HeimdallValidatorSet) (bvs ValidatorSet, err error) {
	proposer, err := ValidatorFromHeimdall(hvs.Proposer)
	if err != nil {
		return
	}
	bvs.Proposer = &proposer

	for _, hv := range hvs.Validators {
		var bv Validator
		bv, err = ValidatorFromHeimdall(hv)
		if err != nil {
			return
		}
		bvs.Validators = append(bvs.Validators, &bv)
	}
	return
}

func ValidatorFromHeimdall(val *polygonTypes.HeimdallValidator) (v Validator, err error) {
	if len(val.PubKey) != 65 {
		err = fmt.Errorf("invalid pubkey from heimdall")
		return
	}
	if len(val.Signer) != 20 {
		err = fmt.Errorf("invalid signer from heimdall")
		return
	}
	v = Validator{
		ID:               uint64(val.ID),
		VotingPower:      val.VotingPower,
		ProposerPriority: val.ProposerPriority,
		Address:          common.BytesToAddress([]byte(val.Signer)),
	}
	return
}
