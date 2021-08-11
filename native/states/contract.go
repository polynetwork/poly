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
	"errors"
	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native/event"
)

const MAX_NATIVE_VERSION = 0

// Invoke smart contract struct
// Param Version: invoke smart contract version, default 0
// Param Address: invoke on blockchain smart contract by address
// Param Method: invoke smart contract method, default ""
// Param Args: invoke smart contract arguments
type ContractInvokeParam struct {
	Version byte
	Address common.Address
	Method  string
	Args    []byte
}

func (this *ContractInvokeParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteByte(this.Version)
	sink.WriteAddress(this.Address)
	sink.WriteVarBytes([]byte(this.Method))
	sink.WriteVarBytes([]byte(this.Args))
}

// `ContractInvokeParam.Args` has reference of `source`
func (this *ContractInvokeParam) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.Version, eof = source.NextByte()
	if eof {
		return errors.New("[ContractInvokeParam] deserialize version error")
	}
	if this.Version > MAX_NATIVE_VERSION {
		return fmt.Errorf("[ContractInvokeParam] current version %d over max native contract version %d", this.Version, MAX_NATIVE_VERSION)
	}

	this.Address, eof = source.NextAddress()
	if eof {
		return errors.New("[ContractInvokeParam] deserialize address error")
	}
	var method []byte
	method, eof = source.NextVarBytes()
	if eof {
		return errors.New("[ContractInvokeParam] deserialize method error")
	}
	this.Method = string(method)

	this.Args, eof = source.NextVarBytes()
	if eof {
		return errors.New("[ContractInvokeParam] deserialize args error")
	}
	return nil
}

type PreExecResult struct {
	State  byte
	Result interface{}
	Notify []*event.NotifyEventInfo
}
