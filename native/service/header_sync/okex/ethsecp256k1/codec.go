/*
 * Copyright (C) 2021 The poly network Authors
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
package ethsecp256k1

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// CryptoCodec is the default amino codec used by ethermint
var CryptoCodec = codec.New()

func init() {
	RegisterCodec(CryptoCodec)
}

// RegisterCodec registers all the necessary types with amino for the given
// codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(PubKey{}, PubKeyName, nil)
	cdc.RegisterConcrete(PrivKey{}, PrivKeyName, nil)
}
