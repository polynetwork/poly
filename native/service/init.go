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

package service

import (
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/cross_chain_manager"
	"github.com/polynetwork/poly/native/service/governance/neo3_state_manager"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	"github.com/polynetwork/poly/native/service/governance/relayer_manager"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/header_sync"
	"github.com/polynetwork/poly/native/service/utils"
)

func init() {
	native.Contracts[utils.SideChainManagerContractAddress] = side_chain_manager.RegisterSideChainManagerContract
	native.Contracts[utils.HeaderSyncContractAddress] = header_sync.RegisterHeaderSyncContract
	native.Contracts[utils.CrossChainManagerContractAddress] = cross_chain_manager.RegisterCrossChainManagerContract
	native.Contracts[utils.NodeManagerContractAddress] = node_manager.RegisterNodeManagerContract
	native.Contracts[utils.RelayerManagerContractAddress] = relayer_manager.RegisterRelayerManagerContract
	native.Contracts[utils.Neo3StateManagerContractAddress] = neo3_state_manager.RegisterStateValidatorManagerContract

	config.EXTRA_INFO_HEIGHT_FORK_CHECK = true
}
