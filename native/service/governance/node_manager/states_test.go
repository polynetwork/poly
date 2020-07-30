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

package node_manager

import (
	"github.com/polynetwork/poly/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Deserialize_GovernanceView(t *testing.T) {
	govView := &GovernanceView{
		View:   0,
		Height: 10,
		TxHash: common.UINT256_EMPTY,
	}
	sink := common.NewZeroCopySink(nil)
	govView.Serialization(sink)

	source := common.NewZeroCopySource(sink.Bytes())
	govView1 := new(GovernanceView)
	err := govView1.Deserialization(source)
	assert.Nil(t, err)
	assert.Equal(t, *govView, *govView1)
}
