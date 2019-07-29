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

package btc

import (
	"bytes"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/wire"
	wire_bch "github.com/gcash/bchd/wire"
	"github.com/gcash/bchutil/merkleblock"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/smartcontract/service/native/side_chain_manager"
)

const (
	BTC_TX_PREFIX string = "btctx"
)

type BTCHandler struct {
}

func NewBTCHandler() *BTCHandler {
	return &BTCHandler{}
}

func (this *BTCHandler) Verify(service *native.NativeService) (*inf.EntranceParam, error) {
	//todo add logic
	return nil, nil
}

func (this *BTCHandler) MakeTransaction(service *native.NativeService, param *inf.EntranceParam) error {
	//todo add logic
	return nil
}

// Verify merkle proof in bytes, and return the result in true or false
// Firstly, calculate the merkleRoot from input `proof`; Then get header.MerkleRoot
// by a spv client and check if they are equal.
func VerifyBtcTx(native *native.NativeService, proof []byte, tx []byte, height uint32,
	pubKeys [][]byte, require int) (bool, error) {
	mb := wire_bch.MsgMerkleBlock{}
	err := mb.BchDecode(bytes.NewReader(proof), wire_bch.ProtocolVersion, wire_bch.LatestEncoding)
	if err != nil {
		return false, err
	}

	mtx := wire.NewMsgTx(wire.TxVersion)
	reader := bytes.NewReader(tx)
	err = mtx.BtcDecode(reader, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return false, fmt.Errorf("VerifyBtcTx, failed to decode the transaction")
	}

	// check the number of tx's outputs and their types
	ret, err := checkTxOutputs(mtx, pubKeys, require)
	if ret != true || err != nil {
		return false, fmt.Errorf("VerifyBtcTx, wrong outputs: %v", err)
	}
	var param targetChainParam
	err = param.resolve(mtx.TxOut[0].Value, mtx.TxOut[1])
	if err != nil {
		return false, fmt.Errorf("VerifyBtcTx, failed to resolve parameter: %v", err)
	}

	//TODO: How to deal with param? We need to check this param, including chain_id, address..
	//check if chainid exist
	sideChain, err := side_chain_manager.GetSideChain(native, param.ChainId)
	if err != nil {
		return false, fmt.Errorf("VerifyBtcTx, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain.Chainid != param.ChainId {
		return false, fmt.Errorf("VerifyBtcTx, side chain is not registered")
	}

	txid := mtx.TxHash()
	if !bytes.Equal(mb.Hashes[0][:], txid[:]) && !bytes.Equal(mb.Hashes[1][:], txid[:]) {
		return false, fmt.Errorf("VerifyBtcTx, wrong transaction hash: %s in proof are not equal with %s",
			mb.Hashes[0].String(), txid.String())
	}

	mBlock := merkleblock.NewMerkleBlockFromMsg(mb)
	merkleRootCalc := mBlock.ExtractMatches()
	if merkleRootCalc == nil || mBlock.BadTree() || len(mBlock.GetMatches()) == 0 {
		return false, fmt.Errorf("VerifyBtcTx, bad merkle tree")
	}

	header, err := NewRestClient().GetHeaderFromSpv(height)
	if err != nil {
		return false, fmt.Errorf("VerifyBtcTx, failed to get header from spv client: %v", err)
	}

	if !bytes.Equal(merkleRootCalc[:], header.MerkleRoot[:]) {
		return false, fmt.Errorf("VerifyBtcTx, merkle root not equal")
	}

	return true, nil
}

// Create a raw transaction that returns the BTC that once locked the multi-sign account
// to the original account and this transacion is not signed. In the end of this function,
// serialized raw transaction would be put into native.CacheDB.
// Parameter `prevTxids` is the txid of the previous output of the transaction input reference,
// `prevIndexes` contain the indexes of the output in the transaction, `amounts` is the mapping
// of accounts and amounts in transaction's output. Return true if building transacion success.
func MakeBtcTx(native *native.NativeService, prevTxids []string, prevIndexes []uint32, amounts map[string]float64) (bool, error) {
	if len(prevIndexes) != len(prevTxids) || len(prevTxids) == 0 {
		return false, fmt.Errorf("MakeBtcTx, wrong num of transaction's inputs")
	}
	var txIns []btcjson.TransactionInput
	for i := 0; i < len(prevTxids); i++ {
		txIns = append(txIns, btcjson.TransactionInput{
			Txid: prevTxids[i],
			Vout: prevIndexes[i],
		})
	}

	mtx, err := getUnsignedTx(txIns, amounts, nil)
	if err != nil {
		return false, fmt.Errorf("MakeBtcTx, get rawtransaction fail: %v", err)
	}
	var buf bytes.Buffer
	err = mtx.BtcEncode(&buf, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return false, fmt.Errorf("MakeBtcTx, serialize rawtransaction fail: %v", err)
	}

	// TODO: Define a key
	native.CacheDB.Put(append([]byte(BTC_TX_PREFIX)), buf.Bytes())
	return true, nil
}
