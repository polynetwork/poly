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

package common

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/tjfoc/gmsm/sm2"
)

const (
	//key prefix
	CROSS_CHAIN_MSG             = "crossChainMsg"
	CURRENT_MSG_HEIGHT          = "currentMsgHeight"
	BLOCK_HEADER                = "blockHeader"
	CURRENT_HEADER_HEIGHT       = "currentHeaderHeight"
	HEADER_INDEX                = "headerIndex"
	CONSENSUS_PEER              = "consensusPeer"
	CONSENSUS_PEER_BLOCK_HEIGHT = "consensusPeerBlockHeight"
	KEY_HEIGHTS                 = "keyHeights"
	ETH_CACHE                   = "ethCaches"
	GENESIS_HEADER              = "genesisHeader"
	ROOT_CERT                   = "rootCert"
	MULTI_ROOT_CERT             = "rootCerts"
	MAIN_CHAIN                  = "mainChain"
	EPOCH_SWITCH                = "epochSwitch"
	SYNC_HEADER_NAME            = "syncHeader"
	SYNC_CROSSCHAIN_MSG         = "syncCrossChainMsg"
	SYNC_CERT                   = "syncCertificate"
	LATEST_HEIGHT_IN_PROCESSING = "latestHeightInProcessing"
)

type HeaderSyncHandler interface {
	SyncGenesisHeader(service *native.NativeService) error
	SyncBlockHeader(service *native.NativeService) error
	SyncCrossChainMsg(service *native.NativeService) error
}

type SyncGenesisHeaderParam struct {
	ChainID       uint64
	GenesisHeader []byte
}

func (this *SyncGenesisHeaderParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ChainID)
	sink.WriteVarBytes(this.GenesisHeader)
}

func (this *SyncGenesisHeaderParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("SyncGenesisHeaderParam deserialize chainID error")
	}
	genesisHeader, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize genesisHeader count error")
	}
	this.ChainID = chainID
	this.GenesisHeader = genesisHeader
	return nil
}

type SyncBlockHeaderParam struct {
	ChainID uint64
	Address common.Address
	Headers [][]byte
}

func (this *SyncBlockHeaderParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ChainID)
	sink.WriteAddress(this.Address)
	sink.WriteUint64(uint64(len(this.Headers)))
	for _, v := range this.Headers {
		sink.WriteVarBytes(v)
	}
}

func (this *SyncBlockHeaderParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("SyncGenesisHeaderParam deserialize chainID error")
	}
	address, eof := source.NextAddress()
	if eof {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error")
	}
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize header count error")
	}
	var headers [][]byte
	for i := 0; uint64(i) < n; i++ {
		header, eof := source.NextVarBytes()
		if eof {

			return fmt.Errorf("utils.DecodeVarBytes, deserialize header error")
		}
		headers = append(headers, header)
	}
	this.ChainID = chainID
	this.Address = address
	this.Headers = headers
	return nil
}

type SyncCrossChainMsgParam struct {
	ChainID        uint64
	Address        common.Address
	CrossChainMsgs [][]byte
}

func (this *SyncCrossChainMsgParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ChainID)
	sink.WriteAddress(this.Address)
	sink.WriteUint64(uint64(len(this.CrossChainMsgs)))
	for _, v := range this.CrossChainMsgs {
		sink.WriteVarBytes(v)
	}
}

func (this *SyncCrossChainMsgParam) Deserialization(source *common.ZeroCopySource) error {
	chainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("SyncGenesisHeaderParam deserialize chainID error")
	}
	address, eof := source.NextAddress()
	if eof {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error")
	}
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize header count error")
	}
	var crossChainMsgs [][]byte
	for i := 0; uint64(i) < n; i++ {
		crossChainMsg, eof := source.NextVarBytes()
		if eof {

			return fmt.Errorf("utils.DecodeVarBytes, deserialize crossChainMsg error")
		}
		crossChainMsgs = append(crossChainMsgs, crossChainMsg)
	}
	this.ChainID = chainID
	this.Address = address
	this.CrossChainMsgs = crossChainMsgs
	return nil
}

func NotifyPutHeader(native *native.NativeService, chainID uint64, height uint64, blockHash string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States:          []interface{}{SYNC_HEADER_NAME, chainID, height, blockHash, native.GetHeight()},
		})
}

func NotifyPutCrossChainMsg(native *native.NativeService, chainID uint64, height uint32) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States:          []interface{}{SYNC_CROSSCHAIN_MSG, chainID, height, native.GetHeight()},
		})
}

func NotifyPutCertificate(native *native.NativeService, chainID uint64, rawCert []byte) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States:          []interface{}{SYNC_CERT, chainID, hex.EncodeToString(rawCert)},
		})
}

type MultiCertTrustChain []*CertTrustChain

func (multi MultiCertTrustChain) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint16(uint16(len(multi)))
	for _, v := range multi {
		v.Serialization(sink)
	}
}

func (multi MultiCertTrustChain) Deserialization(source *common.ZeroCopySource) (MultiCertTrustChain, error) {
	l, eof := source.NextUint16()
	if eof {
		return nil, fmt.Errorf("failed to deserialize length")
	}

	for cnt := uint16(0); cnt < l; cnt++ {
		set := &CertTrustChain{}
		if err := set.Deserialization(source); err != nil {
			return nil, fmt.Errorf("failed to deserialize No.%d cert trust chain: %v", cnt, err)
		}
		multi = append(multi, set)
	}

	return multi, nil
}

func (multi MultiCertTrustChain) ValidateAll(ns *native.NativeService) error {
	for i, v := range multi {
		if err := v.Validate(ns); err != nil {
			return fmt.Errorf("No.%d cert trust chain validate failed: %v", i, err)
		}
	}
	return nil
}

type CertTrustChain struct {
	Certs []*sm2.Certificate
}

func (set *CertTrustChain) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint16(uint16(len(set.Certs)))
	for _, v := range set.Certs {
		sink.WriteVarBytes(v.Raw)
	}
}

func (set *CertTrustChain) Deserialization(source *common.ZeroCopySource) (err error) {
	l, eof := source.NextUint16()
	if eof {
		return fmt.Errorf("failed to deserialize length")
	}
	set.Certs = make([]*sm2.Certificate, l)
	for i := uint16(0); i < l; i++ {
		raw, eof := source.NextVarBytes()
		if eof {
			return fmt.Errorf("failed to get raw bytes for No.%d cert", i)
		}
		set.Certs[i], err = sm2.ParseCertificate(raw)
		if err != nil {
			return fmt.Errorf("failed to parse cert for No.%d: %v", i, err)
		}
	}
	return nil
}

func (set *CertTrustChain) Validate(ns *native.NativeService) error {
	now := ns.GetBlockTime()
	for i, c := range set.Certs {
		//if c.BasicConstraintsValid {
		//	if !c.IsCA {
		//		return fmt.Errorf("No.%d cert is not CA", i)
		//	}
		//}
		if now.Before(c.NotBefore) || now.After(c.NotAfter) {
			return fmt.Errorf("wrong time for no.%d CA: "+
				"(start: %d, end: %d, block_time: %d)",
				i, c.NotBefore.Unix(), c.NotAfter.Unix(), now.Unix())
		}
	}

	return nil
}

func (set *CertTrustChain) ValidCAs(ns *native.NativeService) *CertTrustChain {
	newSet := &CertTrustChain{
		Certs: make([]*sm2.Certificate, 0),
	}
	now := ns.GetBlockTime()
	for _, c := range set.Certs {
		//if c.BasicConstraintsValid {
		//	if !c.IsCA {
		//		continue
		//	}
		//}
		if now.Before(c.NotBefore) || now.After(c.NotAfter) {
			continue
		}
		newSet.Certs = append(newSet.Certs, c)
	}

	return newSet
}

func (set *CertTrustChain) CheckSigWithRootCert(root *sm2.Certificate, signed, sig []byte) error {
	for i, c := range set.Certs {
		if err := c.CheckSignatureFrom(root); err != nil {
			return fmt.Errorf("failed to check sig for No.%d cert from parent: %v", i, err)
		}
		root = c
	}
	if err := root.CheckSignature(root.SignatureAlgorithm, signed, sig); err != nil {
		return fmt.Errorf("failed to check the signature: %v", err)
	}
	return nil
}

func (set *CertTrustChain) CheckSig(signed, sig []byte) error {
	if len(set.Certs) < 1 {
		return errors.New("no cert in chain")
	}
	root := set.Certs[0]
	for i, c := range set.Certs[1:] {
		if err := c.CheckSignatureFrom(root); err != nil {
			return fmt.Errorf("failed to check sig for No.%d cert from parent: %v", i, err)
		}
		root = c
	}
	if err := root.CheckSignature(root.SignatureAlgorithm, signed, sig); err != nil {
		return fmt.Errorf("failed to check the signature: %v", err)
	}
	return nil
}
