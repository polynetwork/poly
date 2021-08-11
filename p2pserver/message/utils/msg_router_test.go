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

package utils

import (
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/p2pserver/message/types"
	"github.com/polynetwork/poly/p2pserver/net/netserver"
	"github.com/polynetwork/poly/p2pserver/net/protocol"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testHandler(data *types.MsgPayload, p2p p2p.P2P, pid *actor.PID, args ...interface{}) {
	log.Info("Test handler")
}

// TestMsgRouter tests a basic function of a message router
func TestMsgRouter(t *testing.T) {
	network := netserver.NewNetServer()
	msgRouter := NewMsgRouter(network)
	assert.NotNil(t, msgRouter)

	msgRouter.RegisterMsgHandler("test", testHandler)
	msgRouter.UnRegisterMsgHandler("test")
	msgRouter.Start()
	msgRouter.Stop()
}
