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

package init

import (
	"bytes"
	"github.com/ontio/multi-chain/smartcontract/service/native/side_chain_manager"
	"math/big"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native/auth"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager"
	cont "github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/ont"
	params "github.com/ontio/multi-chain/smartcontract/service/native/global_params"
	"github.com/ontio/multi-chain/smartcontract/service/native/governance"
	"github.com/ontio/multi-chain/smartcontract/service/native/header_sync"
	"github.com/ontio/multi-chain/smartcontract/service/native/ong"
	"github.com/ontio/multi-chain/smartcontract/service/native/ont"
	"github.com/ontio/multi-chain/smartcontract/service/native/ontid"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
	"github.com/ontio/multi-chain/smartcontract/service/neovm"
	vm "github.com/ontio/multi-chain/vm/neovm"
)

var (
	COMMIT_DPOS_BYTES = InitBytes(utils.GovernanceContractAddress, governance.COMMIT_DPOS)
)

func init() {
	ong.InitOng()
	ont.InitOnt()
	params.InitGlobalParams()
	ontid.Init()
	auth.Init()
	governance.InitGovernance()
	cont.InitCrossChain()
	header_sync.InitHeaderSync()
	side_chain_manager.InitSideChainManager()
	cross_chain_manager.InitEntrance()
}

func InitBytes(addr common.Address, method string) []byte {
	bf := new(bytes.Buffer)
	builder := vm.NewParamsBuilder(bf)
	builder.EmitPushByteArray([]byte{})
	builder.EmitPushByteArray([]byte(method))
	builder.EmitPushByteArray(addr[:])
	builder.EmitPushInteger(big.NewInt(0))
	builder.Emit(vm.SYSCALL)
	builder.EmitPushByteArray([]byte(neovm.NATIVE_INVOKE_NAME))

	return builder.ToArray()
}
