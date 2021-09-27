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

package consensus_vote

import (
	"crypto/sha256"
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/utils"
)

type VoteHandler struct {
}

func NewVoteHandler() *VoteHandler {
	return &VoteHandler{}
}

func (this *VoteHandler) MakeDepositProposal(service *native.NativeService) (*scom.MakeTxParam, error) {
	params := new(scom.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("vote MakeDepositProposal, contract params deserialize error: %v", err)
	}

	//check witness
	address, err := common.AddressParseFromBytes(params.RelayerAddress)
	if err != nil {
		return nil, fmt.Errorf("vote MakeDepositProposal, common.AddressParseFromBytes error: %v", err)
	}
	err = utils.ValidateOwner(service, address)
	if err != nil {
		return nil, fmt.Errorf("vote MakeDepositProposal, checkWitness error: %v", err)
	}

	//use sourcechainid, height, extra as unique id
	unique := &scom.EntranceParam{
		SourceChainID: params.SourceChainID,
		Height:        params.Height,
		Extra:         params.Extra,
	}
	sink := common.NewZeroCopySink(nil)
	unique.Serialization(sink)
	temp := sha256.Sum256(sink.Bytes())
	id := temp[:]

	ok, err := CheckVotes(service, id, address)
	if err != nil {
		return nil, fmt.Errorf("vote MakeDepositProposal, CheckVotes error: %v", err)
	}
	if ok {
		data := common.NewZeroCopySource(params.Extra)
		txParam := new(scom.MakeTxParam)
		if err := txParam.Deserialization(data); err != nil {
			return nil, fmt.Errorf("vote MakeDepositProposal, deserialize MakeTxParam error:%s", err)
		}
		return txParam, nil
	}
	return nil, nil
}
