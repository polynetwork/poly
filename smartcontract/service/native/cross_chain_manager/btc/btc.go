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
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	wire_bch "github.com/gcash/bchd/wire"
	"github.com/gcash/bchutil/merkleblock"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/smartcontract/service/native/side_chain_manager"
)

type BTCHandler struct {
}

func NewBTCHandler() *BTCHandler {
	return &BTCHandler{}
}

func (this *BTCHandler) Verify(service *native.NativeService) (*inf.MakeTxParam, error) {
	params := new(inf.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.Input)); err != nil {
		return nil, fmt.Errorf("btc Verify, contract params deserialize error: %v", err)
	}
	if params.Proof == "" || params.TxData == "" {
		return nil, fmt.Errorf("btc Verify, input data can't be empty")
	}
	tx, err := hex.DecodeString(params.TxData)
	if err != nil {
		return nil, fmt.Errorf("btc Verify, failed to decode transaction from string to bytes: %v", err)
	}
	proof, err := hex.DecodeString(params.Proof)
	if err != nil {
		return nil, fmt.Errorf("btc Verify, failed to decode proof from string to bytes: %v", err)
	}
	ok, p, err := verifyBtcTx(service, proof, tx, params.Height)
	if err != nil {
		return nil, fmt.Errorf("btc Verify, failed to verify: %v", err)
	} else if ok != true {
		return nil, fmt.Errorf("btc Verify, verify not passed")
	}
	
	return &inf.MakeTxParam{
		FromChainID: params.SourceChainID,
		ToChainID:   p.ChainId, 
		Address:     hex.EncodeToString(p.Addr),
		Amount:      uint64(p.Value),
	}, nil
}

func (this *BTCHandler) MakeTransaction(service *native.NativeService, param *inf.MakeTxParam) error {
	amounts := make(map[string]int64)
	amounts[param.Address] = int64(param.Amount) // ??
	err := makeBtcTx(service, amounts)
	if err != nil {
		return fmt.Errorf("btc MakeTransaction, failed to make transaction: %v", err)
	}
	return nil
}

// Verify merkle proof in bytes, and return the result in true or false
// Firstly, calculate the merkleRoot from input `proof`; Then get header.MerkleRoot
// by a spv client and check if they are equal.
func verifyBtcTx(native *native.NativeService, proof []byte, tx []byte, height uint32) (bool, *targetChainParam, error) {
	cli := NewRestClient()
	besth, err := cli.GetCurrentHeightFromSpv()
	if err != nil {
		return false, nil, fmt.Errorf("verifyBtcTx, failed to get current height from spv: %v", err)
	}
	if besth - height < CONFRIMATION {
		return false, nil, fmt.Errorf("verifyBtcTx, transaction is not confirmed, current height: %d, " +
			"input height: %d", besth, height)
	}

	mtx := wire.NewMsgTx(wire.TxVersion)
	reader := bytes.NewReader(tx)
	err = mtx.BtcDecode(reader, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return false, nil, fmt.Errorf("verifyBtcTx, failed to decode the transaction")
	}

	mb := wire_bch.MsgMerkleBlock{}
	err = mb.BchDecode(bytes.NewReader(proof), wire_bch.ProtocolVersion, wire_bch.LatestEncoding)
	if err != nil {
		return false, nil, fmt.Errorf("")
	}

	txid := mtx.TxHash()
	if !bytes.Equal(mb.Hashes[0][:], txid[:]) && !bytes.Equal(mb.Hashes[1][:], txid[:]) {
		return false, nil, fmt.Errorf("verifyBtcTx, wrong transaction hash: %s in proof are not equal with %s",
			mb.Hashes[0].String(), txid.String())
	}
	val, err := native.CacheDB.Get(append([]byte(VERIFIED_TX), txid[:]...))
	if err != nil {
		return false, nil, fmt.Errorf("verifyBtcTx, failed to get verified transaction: %v", err)
	} else if bytes.Equal(val, []byte{1}) {
		return false, nil, fmt.Errorf("verifyBtcTx, transaction already exist")
	}

	// check the number of tx's outputs and their types
	pubKeys := getPubKeys()
	ok, err := checkTxOutputs(mtx, pubKeys, REQUIRE)
	if ok != true || err != nil {
		return false, nil, fmt.Errorf("verifyBtcTx, wrong outputs: %v", err)
	}
	var param targetChainParam
	err = param.resolve(mtx.TxOut[0].Value, mtx.TxOut[1])
	if err != nil {
		return false, nil, fmt.Errorf("verifyBtcTx, failed to resolve parameter: %v", err)
	}

	//TODO: How to deal with param? We need to check this param, including chain_id, address..
	//check if chainid exist
	sideChain, err := side_chain_manager.GetSideChain(native, param.ChainId)
	if err != nil {
		return false, nil, fmt.Errorf("verifyBtcTx, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain.Chainid != param.ChainId {
		return false, nil, fmt.Errorf("verifyBtcTx, side chain is not registered")
	}

	mBlock := merkleblock.NewMerkleBlockFromMsg(mb)
	merkleRootCalc := mBlock.ExtractMatches()
	if merkleRootCalc == nil || mBlock.BadTree() || len(mBlock.GetMatches()) == 0 {
		return false, nil, fmt.Errorf("verifyBtcTx, bad merkle tree")
	}

	header, err := cli.GetHeaderFromSpv(height)
	if err != nil {
		return false, nil, fmt.Errorf("verifyBtcTx, failed to get header from spv client: %v", err)
	}
	if !bytes.Equal(merkleRootCalc[:], header.MerkleRoot[:]) {
		return false, nil, fmt.Errorf("verifyBtcTx, merkle root not equal")
	}

	native.CacheDB.Put(append([]byte(VERIFIED_TX), txid[:]...), []byte{1})
	return true, &param, nil
}

// Create a raw transaction that returns the BTC that once locked the multi-sign account
// to the original account and this transacion is not signed. In the end of this function,
// serialized raw transaction would be put into native.CacheDB.
// Parameter `prevTxids` is the txid of the previous output of the transaction input reference,
// `prevIndexes` contain the indexes of the output in the transaction, `amounts` is the mapping
// of accounts and amounts in transaction's output. Return true if building transacion success.
func makeBtcTx(native *native.NativeService, amounts map[string]int64) error {
	if len(amounts) == 0 {
		return fmt.Errorf("makeBtcTx, input no amount")
	}
	var amountSum int64
	for i, a := range amounts {
		if a <= 0 || a > btcutil.MaxSatoshi {
			return fmt.Errorf("makeBtcTx, wrong amount: amounts[%s]=%d", i, a)
		}
		amountSum += int64(a)
	}
	if amountSum > btcutil.MaxSatoshi {
		return fmt.Errorf("makeBtcTx, sum(%d) of amounts exceeds the MaxSatoshi", amountSum)
	}

	pubKeys := getPubKeys()
	script, err := buildScript(pubKeys, REQUIRE)
	if err != nil {
		return fmt.Errorf("makeBtcTx, failed to get multiPk-script: %v", err)
	}

	txIns, sum, err := NewRestClient().GetUtxosFromSpv(script, amountSum, FEE)
	if err != nil {
		return fmt.Errorf("makeBtcTx, failed to get utxo from spv: %v", err)
	} else if sum <= 0 || len(txIns) == 0 {
		return fmt.Errorf("makeBtcTx, utxo not found")
	}

	change := sum - amountSum
	if change < 0 {
		return fmt.Errorf("makeBtcTx, not enough utxos: the change amount cannot be less than 0, change " +
			"is %d satoshi", change)
	}

	mtx, err := getUnsignedTx(txIns, amounts, change, script,nil)
	if err != nil {
		return fmt.Errorf("makeBtcTx, get rawtransaction fail: %v", err)
	}

	var buf bytes.Buffer
	err = mtx.BtcEncode(&buf, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return fmt.Errorf("makeBtcTx, serialize rawtransaction fail: %v", err)
	}

	// TODO: Define a key
	native.CacheDB.Put(append([]byte(BTC_TX_PREFIX)), buf.Bytes())
	return nil
}
