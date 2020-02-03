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

package service

import (
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/cross_chain_manager"
	"github.com/ontio/multi-chain/native/service/governance/node_manager"
	"github.com/ontio/multi-chain/native/service/governance/relayer_manager"
	"github.com/ontio/multi-chain/native/service/governance/side_chain_manager"
	"github.com/ontio/multi-chain/native/service/header_sync"
	"github.com/ontio/multi-chain/native/service/ont"
	"github.com/ontio/multi-chain/native/service/ont_lock_proxy"
	"github.com/ontio/multi-chain/native/service/utils"
)

func init() {
	native.Contracts[utils.SideChainManagerContractAddress] = side_chain_manager.RegisterSideChainManagerContract
	native.Contracts[utils.HeaderSyncContractAddress] = header_sync.RegisterHeaderSyncContract
	native.Contracts[utils.CrossChainManagerContractAddress] = cross_chain_manager.RegisterCrossChainManagerContract
	native.Contracts[utils.NodeManagerContractAddress] = node_manager.RegisterNodeManagerContract
	native.Contracts[utils.RelayerManagerContractAddress] = relayer_manager.RegisterRelayerManagerContract
	native.Contracts[utils.OntContractAddress] = ont.RegisterOntContract
	native.Contracts[utils.OntLockProxyContractAddress] = ont_lock_proxy.RegisterOntLockContract
}
