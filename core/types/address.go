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

package types

import (
	"github.com/polynetwork/poly/common"
	"github.com/ontio/ontology-crypto/keypair"
)

func AddressFromPubKey(pubkey keypair.PublicKey) common.Address {
	buf := keypair.SerializePublicKey(pubkey)
	return common.AddressFromVmCode(buf)
}

func AddressFromMultiPubKeys(pubkeys []keypair.PublicKey, m int) (common.Address, error) {
	sink := common.NewZeroCopySink(nil)
	if err := EncodeMultiPubKeyProgramInto(sink, pubkeys, uint16(m)); err != nil {
		return common.ADDRESS_EMPTY, nil
	}
	return common.AddressFromVmCode(sink.Bytes()), nil
}

func AddressFromBookkeepers(bookkeepers []keypair.PublicKey) (common.Address, error) {
	if len(bookkeepers) == 1 {
		return AddressFromPubKey(bookkeepers[0]), nil
	}
	return AddressFromMultiPubKeys(bookkeepers, len(bookkeepers)-(len(bookkeepers)-1)/3)
}
