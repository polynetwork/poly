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
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// Snapshot is the state of the authorization voting at a given point in time.
type Snapshot struct {
	Hash         common.Hash   `json:"hash"`         // Block hash where the snapshot was created
	ValidatorSet *ValidatorSet `json:"validatorSet"` // Validator set at this moment
}

func (s *Snapshot) GetSignerSuccessionNumber(signer common.Address) (int, error) {
	vs := s.ValidatorSet
	proposer := vs.GetProposer().Address
	proposerIndex, _ := vs.GetByAddress(proposer)
	if proposerIndex == -1 {
		return -1, fmt.Errorf("UnauthorizedProposerError:%s", proposer.Hex())
	}
	signerIndex, _ := vs.GetByAddress(signer)
	if signerIndex == -1 {
		return -1, fmt.Errorf("UnauthorizedProposerError:%s", proposer.Hex())
	}

	tempIndex := signerIndex
	if proposerIndex != tempIndex {
		if tempIndex < proposerIndex {
			tempIndex = tempIndex + len(vs.Validators)
		}
	}
	return tempIndex - proposerIndex, nil
}

func (s *Snapshot) Difficulty(signer common.Address) uint64 {
	// if signer is empty
	if bytes.Compare(signer.Bytes(), common.Address{}.Bytes()) == 0 {
		return 1
	}

	validators := s.ValidatorSet.Validators
	proposer := s.ValidatorSet.GetProposer().Address
	totalValidators := len(validators)

	proposerIndex, _ := s.ValidatorSet.GetByAddress(proposer)
	signerIndex, _ := s.ValidatorSet.GetByAddress(signer)

	// temp index
	tempIndex := signerIndex
	if tempIndex < proposerIndex {
		tempIndex = tempIndex + totalValidators
	}

	return uint64(totalValidators - (tempIndex - proposerIndex))
}

// only used in test
func (s *Snapshot) equal(s2 *Snapshot) bool {

	return s.ValidatorSet.String() == s2.ValidatorSet.String()
}
