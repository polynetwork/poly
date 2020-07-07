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
package ont

import (
	"encoding/hex"
	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/merkle"
	common2 "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStateroot(t *testing.T) {
	crossStatesRoot := "91d969e63a2cf0c0e9bc76ed5aa99d3d024696d249229baf1f6dc688967e240b"
	root, err := common.Uint256FromHexString(crossStatesRoot)
	if err != nil {
		fmt.Println("common.Uint256FromHexString", err)
	}
	proofHex := "80000000000000000107283730613439356165323865646233316432613661613164643439613261643863616335393736643001012a30783030303030303030303030303030303030303030303030303030303030303030303030303030303001022241476a44344d6f32356b7a6353747968317374703774586b55754d6f704434334e54020384"
	proof, err := hex.DecodeString(proofHex)
	if err != nil {
		fmt.Println("hex.DecodeString", err)
	}

	v, err := merkle.MerkleProve(proof, root[:])

	s := common.NewZeroCopySource(v)
	merkleValue := new(common2.ToMerkleValue)
	err = merkleValue.Deserialization(s)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%d, %x\n", merkleValue.FromChainID, merkleValue.TxHash)
	fmt.Printf("%s\n", merkleValue.MakeTxParam.FromContractAddress)
	assert.NotNil(t, v)
}
