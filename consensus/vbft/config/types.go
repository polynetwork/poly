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

package vconfig

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/poly/core/types"
)

// PubkeyID returns a marshaled representation of the given public key.
func PubkeyID(pub keypair.PublicKey) string {
	nodeid := hex.EncodeToString(keypair.SerializePublicKey(pub))
	return nodeid
}

func Pubkey(nodeid string) (keypair.PublicKey, error) {
	pubKey, err := hex.DecodeString(nodeid)
	if err != nil {
		return nil, err
	}
	pk, err := keypair.DeserializePublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("deserialize failed: %s", err)
	}
	return pk, err
}

func VbftBlock(header *types.Header) (*VbftBlockInfo, error) {
	blkInfo := &VbftBlockInfo{}
	if err := json.Unmarshal(header.ConsensusPayload, blkInfo); err != nil {
		return nil, fmt.Errorf("unmarshal blockInfo: %s", err)
	}
	return blkInfo, nil
}
