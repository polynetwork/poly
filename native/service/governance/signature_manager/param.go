/*
 * Copyright (C) 2022 The poly network Authors
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
package signature_manager

import (
	"fmt"

	"github.com/polynetwork/poly/common"
)

type AddSignatureParam struct {
	Address     common.Address
	SideChainID uint64
	Subject     []byte
	Signature   []byte
}

func (this *AddSignatureParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.Address[:])
	sink.WriteUint64(this.SideChainID)
	sink.WriteVarBytes(this.Subject)
	sink.WriteVarBytes(this.Signature)
}

func (this *AddSignatureParam) Deserialization(source *common.ZeroCopySource) error {

	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize address error")
	}
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("common.AddressParseFromBytes, deserialize address error: %s", err)
	}
	sideChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("common.NextUint64, deserialize sideChainID error: %s", err)
	}

	subject, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize subject error")
	}

	signature, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize signature error")
	}

	this.Address = addr
	this.SideChainID = sideChainID
	this.Subject = subject
	this.Signature = signature
	return nil
}
