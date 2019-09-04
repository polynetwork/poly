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
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager"
	cont "github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/ont"
	"github.com/ontio/multi-chain/smartcontract/service/native/header_sync"
	"github.com/ontio/multi-chain/smartcontract/service/native/side_chain_manager"
)


func init() {
	cont.InitCrossChain()
	header_sync.InitHeaderSync()
	side_chain_manager.InitSideChainManager()
	cross_chain_manager.InitEntrance()
}