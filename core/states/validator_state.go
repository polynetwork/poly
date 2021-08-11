/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */

package states

import (
	"fmt"
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/common/serialization"
)

type ValidatorState struct {
	StateBase
	PublicKey keypair.PublicKey
}

func (this *ValidatorState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	buf := keypair.SerializePublicKey(this.PublicKey)
	if err := serialization.WriteVarBytes(w, buf); err != nil {
		return err
	}
	return nil
}

func (this *ValidatorState) Deserialize(r io.Reader) error {
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return fmt.Errorf("[ValidatorState], StateBase Deserialize failed, error:%s", err)
	}
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("[ValidatorState], PublicKey Deserialize failed, error:%s", err)
	}
	pk, err := keypair.DeserializePublicKey(buf)
	if err != nil {
		return fmt.Errorf("[ValidatorState], PublicKey Deserialize failed, error:%s", err)
	}
	this.PublicKey = pk
	return nil
}
