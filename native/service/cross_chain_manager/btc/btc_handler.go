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

	"github.com/btcsuite/btcd/txscript"
	sneovm "github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/vm/neovm"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	wire_bch "github.com/gcash/bchd/wire"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	crosscommon "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/side_chain_manager"
	"github.com/ontio/multi-chain/native/service/utils"
	vtypes "github.com/ontio/ontology/vm/neovm/types"
)

const (
	BTC_ADDRESS      = "btc"
	NOTIFY_BTC_PROOF = "notifyBtcProof"
	UTXOS            = "utxos"
	REDEEM_SCRIPT    = "redeemScript"
	MULTI_SIGN_INFO  = "multiSignInfo"
)

type BTCHandler struct {
}

func NewBTCHandler() *BTCHandler {
	return &BTCHandler{}
}

func (this *BTCHandler) InitRedeemScript(service *native.NativeService) error {
	params := new(crosscommon.InitRedeemScriptParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return fmt.Errorf("InitRedeemScript, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return err
	}

	//check witness
	err = utils.ValidateOwner(service, operatorAddress)
	if err != nil {
		return fmt.Errorf("InitRedeemScript, checkWitness error: %v", err)
	}

	err = putBtcRedeemScript(service, params.RedeemScript)
	if err != nil {
		return fmt.Errorf("InitRedeemScript, putBtcRedeemScript error: %v", err)
	}

	return nil
}

func (this *BTCHandler) MultiSign(service *native.NativeService) error {
	params := new(crosscommon.MultiSignParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return fmt.Errorf("MultiSign, contract params deserialize error: %v", err)
	}
	txHash, err := chainhash.NewHash(params.TxHash)
	if err != nil {
		return fmt.Errorf("MultiSign, failed to get tx hash from param: %v", err)
	}

	txb, err := service.GetCacheDB().Get(utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(BTC_TX_PREFIX), txHash[:]))
	if err != nil {
		return fmt.Errorf("MultiSign, failed to get tx %s from cacheDB: %v", txHash.String(), err)
	}
	mtx := wire.NewMsgTx(wire.TxVersion)
	err = mtx.BtcDecode(bytes.NewBuffer(txb), wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return fmt.Errorf("MultiSign, failed to decode tx: %v", err)
	}

	redeemScript, err := getBtcRedeemScriptBytes(service)
	if err != nil {
		return fmt.Errorf("MultiSign, getBtcRedeemScript error: %v", err)
	}
	cls, addrs, n, err := txscript.ExtractPkScriptAddrs(redeemScript, netParam)
	if err != nil {
		return fmt.Errorf("MultiSign, failed to extract pkscript addrs: %v", err)
	}
	if cls.String() != "multisig" {
		return fmt.Errorf("MultiSign, wrong class of redeem: %s", cls.String())
	}

	err = verifySigs(params.Signs, params.Address, addrs, redeemScript, mtx)
	if err != nil {
		return fmt.Errorf("MultiSign, failed to verify: %v", err)
	}

	multiSignInfo, err := getBtcMultiSignInfo(service, params.TxHash)
	if err != nil {
		return fmt.Errorf("MultiSign, getBtcMultiSignInfo error: %v", err)
	}
	multiSignInfo.MultiSignInfo[params.Address] = params.Signs
	err = putBtcMultiSignInfo(service, params.TxHash, multiSignInfo)
	if err != nil {
		return fmt.Errorf("MultiSign, putBtcMultiSignInfo error: %v", err)
	}

	if len(multiSignInfo.MultiSignInfo) != n {
		service.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: utils.CrossChainManagerContractAddress,
				States:          []interface{}{"btcTxMultiSign", params.TxHash, multiSignInfo.MultiSignInfo},
			})
	} else {
		mtx, err = addSigToTx(multiSignInfo, addrs, redeemScript, mtx, len(params.Signs))
		if err != nil {
			return fmt.Errorf("MultiSign, failed to add sig to tx: %v", err)
		}
		var buf bytes.Buffer
		err = mtx.BtcEncode(&buf, wire.ProtocolVersion, wire.LatestEncoding)
		if err != nil {
			return fmt.Errorf("MultiSign, failed to encode msgtx to bytes: %v", err)
		}

		p2shAddr, err := btcutil.NewAddressScriptHash(redeemScript, netParam)
		if err != nil {
			return fmt.Errorf("MultiSign, btcutil.NewAddressScriptHash, failed to get p2sh: %v", err)
		}
		p2shScript, err := txscript.PayToAddrScript(p2shAddr)
		if err != nil {
			return fmt.Errorf("MultiSign, txscript.PayToAddrScript, failed to get p2sh script: %v", err)
		}

		utxos, err := getUtxos(service, utils.BTC_CHAIN_ID)
		if err != nil {
			return fmt.Errorf("MultiSign, getUtxos error: %v", err)
		}
		txid := mtx.TxHash()
		for i, v := range mtx.TxOut {
			if bytes.Compare(p2shScript, v.PkScript) == 0 {
				newUtxo := &Utxo{
					Op: &OutPoint{
						Hash:  txid[:],
						Index: uint32(i),
					},
					Value:        uint64(v.Value),
					ScriptPubkey: v.PkScript,
				}
				utxos.Utxos = append(utxos.Utxos, newUtxo)
			}
		}
		err = putUtxos(service, 0, utxos)
		if err != nil {
			return fmt.Errorf("MultiSign, putUtxos error: %v", err)
		}

		service.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: utils.CrossChainManagerContractAddress,
				States:          []interface{}{"btcTxToRelay", hex.EncodeToString(buf.Bytes())},
			})
	}
	return nil
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
		return false, nil, nil
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

	err = addUtxos(service, utils.BTC_CHAIN_ID, proof.Height, mtx)
	if err != nil {
		return false, nil, fmt.Errorf("btc Vote, updateUtxo error: %s", err)
	}

	var p targetChainParam
	toContract, err := p.resolve(mtx.TxOut[0].Value, mtx.TxOut[1])
	if err != nil {
		return false, nil, fmt.Errorf("btc Vote, failed to resolve parameter: %v", err)
	}

	txHash := mtx.TxHash()
	return true, &crosscommon.MakeTxParam{
		TxHash:              txHash[:],
		FromContractAddress: []byte(BTC_ADDRESS),
		ToChainID:           p.ChainId,
		ToContractAddress:   toContract,
		Method:              "unlock",
		Args:                p.AddrAndVal,
	}, nil
}

func (this *BTCHandler) MakeDepositProposal(service *native.NativeService) (*crosscommon.MakeTxParam, error) {
	params := new(crosscommon.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("btc MakeDepositProposal, contract params deserialize error: %v", err)
	}
	if len(params.Proof) == 0 || len(params.Extra) == 0 {
		return nil, fmt.Errorf("btc MakeDepositProposal, GetInput() data can't be empty")
	}
	err := notifyBtcTx(service, params.Proof, params.Extra, params.Height, params.SourceChainID)
	if err != nil {
		return nil, fmt.Errorf("btc MakeDepositProposal, failed to verify: %v", err)
	}

	return nil, nil
}

func (this *BTCHandler) MakeTransaction(service *native.NativeService, param *crosscommon.MakeTxParam) error {
	if bytes.Equal(param.ToContractAddress, []byte(BTC_ADDRESS)) {
		return fmt.Errorf("btc MakeTransaction, destContractAddr is %s not btc", string(param.ToContractAddress))
	}
	amounts := make(map[string]int64)

	bf := bytes.NewBuffer(param.Args)
	items, err := sneovm.DeserializeStackItem(bf)
	if err != nil {
		return fmt.Errorf("neovm.DeserializeStackItem error:%s", err)
	}
	toAddr, err := items.(*vtypes.Map).TryGetValue(neovm.NewStackItem([]byte("address"))).GetByteArray()
	if err != nil {
		return fmt.Errorf("deserialize toAddr error:%s", err)
	}
	amount, err := items.(*vtypes.Map).TryGetValue(neovm.NewStackItem([]byte("amount"))).GetBigInteger()
	if err != nil {
		return fmt.Errorf("deserialize amount error:%s", err)
	}
	amounts[string(toAddr)] = amount.Int64()

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
		return fmt.Errorf("notifyBtcTx, failed to decode the transaction %s: %s", hex.EncodeToString(tx), err)
	}

	mb := wire_bch.MsgMerkleBlock{}
	err = mb.BchDecode(bytes.NewReader(proof), wire_bch.ProtocolVersion, wire_bch.LatestEncoding)
	if err != nil {
		return fmt.Errorf("notifyBtcTx, failed to decode proof: %v", err)
	}

	txid := mtx.TxHash()

	ok, err := checkBtcProof(native, txid[:])
	if err != nil {
		return fmt.Errorf("notifyBtcTx, checkBtcProof error: %v", err)
	}
	if !ok {
		return fmt.Errorf("notifyBtcTx, btc proof already exist")
	}

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

	notifyBtcProof(native, txid.String(), hex.EncodeToString(sink.Bytes()))
	return nil
}

func makeBtcTx(service *native.NativeService, chainID uint64, amounts map[string]int64) error {
	if len(amounts) == 0 {
		return fmt.Errorf("makeBtcTx, GetInput() no amount")
	}
	var amountSum int64
	for k, v := range amounts {
		if v <= 0 || v > btcutil.MaxSatoshi {
			return fmt.Errorf("makeBtcTx, wrong amount: amounts[%s]=%d", k, v)
		}
		amountSum += int64(v)
	}
	if amountSum > btcutil.MaxSatoshi {
		return fmt.Errorf("makeBtcTx, sum(%d) of amounts exceeds the MaxSatoshi", amountSum)
	}
	if amountSum <= FEE {
		return fmt.Errorf("makeBtcTx, amounts sum(%d) must greater than fee %d", amountSum, FEE)
	}

	for k, v := range amounts {
		amounts[k] = v - int64(float64(FEE*v)/float64(amountSum))
	}
	redeemScript, err := getBtcRedeemScriptBytes(service)
	if err != nil {
		return fmt.Errorf("makeBtcTx, getBtcRedeemScript error: %v", err)
	}

	choosed, sum, err := chooseUtxos(service, chainID, amountSum, 0)
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

	charge := sum - amountSum
	mtx, err := getUnsignedTx(txIns, amounts, charge, redeemScript, nil)
	if err != nil {
		return fmt.Errorf("makeBtcTx, get rawtransaction fail: %v", err)
	}

	var buf bytes.Buffer
	err = mtx.BtcEncode(&buf, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return fmt.Errorf("makeBtcTx, serialize rawtransaction fail: %v", err)
	}
	txHash := mtx.TxHash()
	service.GetCacheDB().Put(utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(BTC_TX_PREFIX),
		txHash[:]), buf.Bytes())

	service.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{"makeBtcTx", hex.EncodeToString(buf.Bytes()), hex.EncodeToString(redeemScript)},
		})

	return nil
}
