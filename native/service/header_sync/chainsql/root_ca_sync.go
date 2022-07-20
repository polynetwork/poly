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
	"encoding/pem"
	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/governance/node_manager"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/tjfoc/gmsm/sm2"
)

type ChainsqlHandler struct{}

func NewChainsqlHandler() *ChainsqlHandler {
	return &ChainsqlHandler{}
}

func (this *ChainsqlHandler) SyncGenesisHeader(ns *native.NativeService) error {
	params := new(scom.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(ns.GetInput())); err != nil {
		return fmt.Errorf("ChainsqlHandler SyncGenesisHeader, contract params deserialize error: %v", err)
	}

	// Get current epoch operator
	operatorAddress, err := node_manager.GetCurConOperator(ns)
	if err != nil {
		return fmt.Errorf("ChainsqlHandler SyncGenesisHeader, get current consensus operator address error: %v", err)
	}
	// check witness
	err = utils.ValidateOwner(ns, operatorAddress)
	if err != nil {
		return fmt.Errorf("ChainsqlHandler SyncGenesisHeader, checkWitness error: %v", err)
	}

	blk, _ := pem.Decode(params.GenesisHeader)
	if blk == nil {
		return fmt.Errorf("ChainsqlHandler SyncGenesisHeader, failed to decode PEM formatted block")
	}
	if blk.Type != "CERTIFICATE" {
		return fmt.Errorf("ChainsqlHandler SyncGenesisHeader, wrong block type: %s", blk.Type)
	}
	cert, err := sm2.ParseCertificate(blk.Bytes)
	if err != nil {
		return fmt.Errorf("ChainsqlHandler SyncGenesisHeader, failed to parse certificate: %v", err)
	}

	now := ns.GetBlockTime()
	if now.After(cert.NotAfter) || now.Before(cert.NotBefore) {
		return fmt.Errorf("ChainsqlHandler SyncGenesisHeader, wrong time for new CA: "+
			"(start: %d, end: %d, block_time: %d)",
			cert.NotBefore.Unix(), cert.NotAfter.Unix(), now.Unix())
	}

	root := &ChainsqlRoot{
		RootCA: cert,
	}

	if err = PutChainsqlRoot(ns, root, params.ChainID); err != nil {
		return fmt.Errorf("ChainsqlHandler SyncGenesisHeader, failed to put new Chainsql root CA into storage: %v", err)
	}

	return nil
}

func (this *ChainsqlHandler) SyncBlockHeader(ns *native.NativeService) error {
	return nil
}

func (this *ChainsqlHandler) SyncCrossChainMsg(ns *native.NativeService) error {
	return nil
}
