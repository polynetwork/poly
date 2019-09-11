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
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	wire_bch "github.com/gcash/bchd/wire"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	crosscommon "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/side_chain_manager"
	"github.com/ontio/multi-chain/native/service/utils"
)

const (
	BTC_ADDRESS      = "btc"
	NOTIFY_BTC_PROOF = "notifyBtcProof"
	UTXOS            = "utxos"
)

type BTCHandler struct {
}

func NewBTCHandler() *BTCHandler {
	return &BTCHandler{}
}

func (this *BTCHandler) Vote(service *native.NativeService) (bool, *crosscommon.MakeTxParam, error) {
	params := new(crosscommon.VoteParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return false, nil, fmt.Errorf("btc Vote, contract params deserialize error: %v", err)
	}

	address, err := common.AddressFromBase58(params.Address)
	if err != nil {
		return false, nil, fmt.Errorf("btc Vote, common.AddressFromBase58 error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(service, address)
	if err != nil {
		return false, nil, fmt.Errorf("btc Vote, utils.ValidateOwner error: %v", err)
	}

	vote, err := getBtcVote(service, params.TxHash)
	if err != nil {
		return false, nil, fmt.Errorf("btc Vote, getBtcVote error: %v", err)
	}
	vote.VoteMap[params.Address] = params.Address
	err = putBtcVote(service, params.TxHash, vote)
	if err != nil {
		return false, nil, fmt.Errorf("btc Vote, putBtcVote error: %v", err)
	}

	err = crosscommon.ValidateVote(service, vote)
	if err != nil {
		return false, nil, fmt.Errorf("btc Vote, ValidateVote error: %v", err)
	}

	proofBytes, err := getBtcProof(service, params.TxHash)
	if err != nil {
		return false, nil, fmt.Errorf("btc Vote, getBtcTx error: %v", err)
	}
	proof := new(BtcProof)
	err = proof.Deserialization(common.NewZeroCopySource(proofBytes))
	if err != nil {
		return false, nil, fmt.Errorf("btc Vote, proof.Deserialization error: %v", err)
	}

	mtx := wire.NewMsgTx(wire.TxVersion)
	reader := bytes.NewReader(proof.Tx)
	err = mtx.BtcDecode(reader, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return false, nil, fmt.Errorf("btc Vote, failed to decode the transaction")
	}

	err = addUtxos(service, params.FromChainID, proof.Height, mtx)
	if err != nil {
		return false, nil, fmt.Errorf("btc Vote, updateUtxo error: %s", err)
	}

	var p targetChainParam
	err = p.resolve(mtx.TxOut[0].Value, mtx.TxOut[1])
	if err != nil {
		return false, nil, fmt.Errorf("btc Vote, failed to resolve parameter: %v", err)
	}

	return true, &crosscommon.MakeTxParam{
		FromChainID:         params.FromChainID,
		FromContractAddress: BTC_ADDRESS,
		ToChainID:           p.ChainId,
		Args:                p.AddrAndVal,
	}, nil
}

func (this *BTCHandler) MakeDepositProposal(service *native.NativeService) (*crosscommon.MakeTxParam, error) {
	params := new(crosscommon.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("btc Verify, contract params deserialize error: %v", err)
	}
	if params.Proof == "" || params.TxData == "" {
		return nil, fmt.Errorf("btc Verify, GetInput() data can't be empty")
	}
	tx, err := hex.DecodeString(params.TxData)
	if err != nil {
		return nil, fmt.Errorf("btc Verify, failed to decode transaction from string to bytes: %v", err)
	}
	proof, err := hex.DecodeString(params.Proof)
	if err != nil {
		return nil, fmt.Errorf("btc Verify, failed to decode proof from string to bytes: %v", err)
	}
	err = notifyBtcTx(service, proof, tx, params.Height, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("btc Verify, failed to verify: %v", err)
	}

	return nil, nil
}

func (this *BTCHandler) MakeTransaction(service *native.NativeService, param *crosscommon.MakeTxParam) error {
	amounts := make(map[string]int64)

	//toAddr := hex.EncodeToString(param.Args[:26])
	//amount := binary.BigEndian.Uint64(param.Args[26:])

	//amounts[toAddr] = int64(amount) // ??
	amounts["mjEoyyCPsLzJ23xMX6Mti13zMyN36kzn57"] = int64(1) // ??

	destAsset, err := side_chain_manager.GetDestAsset(service, param.FromChainID,
		param.ToChainID, param.FromContractAddress)
	if err != nil {
		return fmt.Errorf("btc MakeTransaction, side_chain_manager.GetAssetContractAddress error: %v", err)
	}
	if destAsset.ContractAddress != "btc" {
		return fmt.Errorf("btc MakeTransaction, destContractAddr is %s not btc", destAsset.ContractAddress)
	}

	err = makeBtcTx(service, param.ToChainID, amounts)
	if err != nil {
		return fmt.Errorf("btc MakeTransaction, failed to make transaction: %v", err)
	}
	return nil
}

func notifyBtcTx(native *native.NativeService, proof, tx []byte, height uint32, btcChainID uint64) error {
	sideChain, err := side_chain_manager.GetSideChain(native, btcChainID)
	if err != nil {
		return fmt.Errorf("notifyBtcTx, side_chain_manager.GetSideChain error: %v", err)
	}

	mtx := wire.NewMsgTx(wire.TxVersion)
	reader := bytes.NewReader(tx)
	err = mtx.BtcDecode(reader, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return fmt.Errorf("notifyBtcTx, failed to decode the transaction")
	}

	mb := wire_bch.MsgMerkleBlock{}
	err = mb.BchDecode(bytes.NewReader(proof), wire_bch.ProtocolVersion, wire_bch.LatestEncoding)
	if err != nil {
		return fmt.Errorf("notifyBtcTx, failed to decode proof: %v", err)
	}

	txid := mtx.TxHash()
	isExist := false
	for _, hash := range mb.Hashes {
		if bytes.Equal(hash[:], txid[:]) {
			isExist = true
			break
		}
	}
	if !isExist {
		return fmt.Errorf("notifyBtcTx, transaction %s not found in proof", txid.String())
	}

	btcProof := &BtcProof{
		Tx:           tx,
		Proof:        proof,
		Height:       height,
		BlocksToWait: sideChain.BlocksToWait,
	}
	sink := common.NewZeroCopySink(nil)
	btcProof.Serialization(sink)

	putBtcProof(native, txid[:], sink.Bytes())

	notifyBtcProof(native, hex.EncodeToString(sink.Bytes()))
	return nil
}

func makeBtcTx(service *native.NativeService, chainID uint64, amounts map[string]int64) error {
	if len(amounts) == 0 {
		return fmt.Errorf("makeBtcTx, GetInput() no amount")
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
	amountSum = amountSum - FEE

	pubKeys := getPubKeys()
	script, err := buildScript(pubKeys, REQUIRE)
	if err != nil {
		return fmt.Errorf("makeBtcTx, failed to get multiPk-script: %v", err)
	}

	choosed, sum, err := chooseUtxos(service, chainID, amountSum, FEE)
	if err != nil {
		return fmt.Errorf("makeBtcTx, chooseUtxos error: %v", err)
	}
	txIns := make([]btcjson.TransactionInput, 0)
	for _, u := range choosed {
		hash, err := chainhash.NewHash(u.Op.Hash)
		if err != nil {
			return fmt.Errorf("makeBtcTx, chainhash.NewHash error: %v", err)
		}
		txIns = append(txIns, btcjson.TransactionInput{hash.String(), u.Op.Index})
	}

	charge := sum - amountSum - FEE
	if charge < 0 {
		return fmt.Errorf("makeBtcTx, not enough utxos: the charge amount cannot be less than 0, charge "+
			"is %d satoshi", charge)
	}

	mtx, err := getUnsignedTx(txIns, amounts, charge, script, nil)
	if err != nil {
		return fmt.Errorf("makeBtcTx, get rawtransaction fail: %v", err)
	}

	var buf bytes.Buffer
	err = mtx.BtcEncode(&buf, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return fmt.Errorf("makeBtcTx, serialize rawtransaction fail: %v", err)
	}

	// TODO: Define a key
	service.GetCacheDB().Put([]byte(BTC_TX_PREFIX), buf.Bytes())
	service.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{"makeBtcTx", hex.EncodeToString(buf.Bytes())},
		})

	// TODO: charge
	return nil
}
