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
package chainsql

import (
	"fmt"
	pcom "github.com/polynetwork/poly/common"
	"github.com/tjfoc/gmsm/sm2"
)

type ChainsqlRoot struct {
	RootCA *sm2.Certificate
}

func (root *ChainsqlRoot) Serialization(sink *pcom.ZeroCopySink) {
	sink.WriteVarBytes(root.RootCA.Raw)
}

func (root *ChainsqlRoot) Deserialization(source *pcom.ZeroCopySource) error {
	var (
		err error
	)
	raw, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("failed to deserialize RootCA")
	}
	root.RootCA, err = sm2.ParseCertificate(raw)
	if err != nil {
		return fmt.Errorf("failed to parse cert: %v", err)
	}

	return nil
}
