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
package quorum

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	pcom "github.com/polynetwork/poly/common"
	"math"
)

type QuorumValSet []common.Address

func (vs QuorumValSet) Serialize(sink *pcom.ZeroCopySink) {
	for _, v := range vs {
		sink.WriteBytes(v.Bytes())
	}
}

func (vs *QuorumValSet) Deserialize(source *pcom.ZeroCopySource) error {
	if source.Size()%common.AddressLength != 0 {
		return fmt.Errorf("wrong size of raw addresses: %d", source.Size())
	}
	l := source.Size() / common.AddressLength
	nvs := QuorumValSet(make([]common.Address, l))

	for i := uint64(0); i < l; i++ {
		raw, eof := source.NextBytes(common.AddressLength)
		if eof {
			return fmt.Errorf("failed to get next %d bytes for the No.%d address", common.AddressLength, i)
		}
		nvs[i] = common.BytesToAddress(raw)
	}
	*vs = nvs

	return nil
}

func (vs QuorumValSet) String() (res string) {
	res = "{\n"
	for i, v := range vs {
		res += fmt.Sprintf("[ %d, %s ],\n", i, v.String())
	}
	res += "}"
	return
}

func VerifyQuorumHeader(vs QuorumValSet, hdr *types.Header, isEpoch bool) (*IstanbulExtra, error) {
	extra, err := ExtractIstanbulExtra(hdr)
	if err != nil {
		return nil, fmt.Errorf("extract istanbul extra from header %s error: %v", GetQuorumHeaderHash(hdr).String(), err)
	}

	checker := vs
	if isEpoch {
		if !vs.IfChanged(extra.Validators) {
			return nil, fmt.Errorf("header %s is not epoch header supposed to contains new validators", GetQuorumHeaderHash(hdr).String())
		}
		if err := vs.JustOneChanged(extra.Validators); err != nil {
			return nil, err
		}
		checker = extra.Validators
	}

	if err := checker.VerifySigner(hdr, extra.Seal); err != nil {
		return nil, fmt.Errorf("failed to verify signer: %v", err)
	}
	if err := checker.VerifyCommittedSeals(extra, GetQuorumHeaderHash(hdr)); err != nil {
		return nil, fmt.Errorf("verify committed seals failed for header %s: %v", GetQuorumHeaderHash(hdr).String(), err)
	}
	return extra, nil
}

func (vs QuorumValSet) VerifySigner(hdr *types.Header, seal []byte) error {
	addr, err := GetSignatureAddress(sigHash(hdr).Bytes(), seal)
	if err != nil {
		return err
	}
	idx, _ := vs.GetByAddress(addr)
	if idx == -1 {
		return fmt.Errorf("signer %s is not in validators", addr.Hex())
	}

	return nil
}

func (vs QuorumValSet) VerifyCommittedSeals(extra *IstanbulExtra, hash common.Hash) error {
	addrs, err := GetSigners(hash, extra.CommittedSeal)
	if err != nil {
		return fmt.Errorf("failed to VerifyCommittedSeals: %v", err)
	}
	validSeal := 1
	for _, v := range addrs {
		if vs.Exist(v) {
			validSeal++
			continue
		}
		return fmt.Errorf("addess %s is not in validators", v.String())
	}
	if validSeal <= vs.F() {
		return fmt.Errorf("valid seal not enough: (%d found, %d required)", validSeal, vs.F()+1)
	}
	return nil
}

func (vs QuorumValSet) Exist(address common.Address) bool {
	for _, v := range vs {
		if bytes.Equal(address.Bytes(), v.Bytes()) {
			return true
		}
	}
	return false
}

func (vs QuorumValSet) GetByAddress(addr common.Address) (int, common.Address) {
	for i, val := range vs {
		if bytes.Equal(addr.Bytes(), val.Bytes()) {
			return i, val
		}
	}
	return -1, common.Address{}
}

func (vs QuorumValSet) IfChanged(another QuorumValSet) bool {
	if len(vs) != len(another) {
		return true
	}
	for i := 0; i < len(vs); i++ {
		if bytes.Equal(vs[i].Bytes(), another[i].Bytes()) {
			continue
		}
		return true
	}
	return false
}

func (vs QuorumValSet) JustOneChanged(another QuorumValSet) error {
	var more, less QuorumValSet

	switch len(vs) - len(another) {
	case 1:
		more, less = vs, another
	case -1:
		more, less = another, vs
	default:
		return fmt.Errorf("length of new validitors is %d but original one is %d", len(another), len(vs))
	}

	for i, j := 0, 0; i < len(less); i, j = i+1, j+1 {
		if !bytes.Equal(less[i].Bytes(), more[j].Bytes()) {
			if j++; j-i > 1 {
				return errors.New("more than one validator changed")
			}
		}
	}

	return nil
}

func (vs QuorumValSet) F() int { return int(math.Ceil(float64(len(vs))/3)) - 1 }
