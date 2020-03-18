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
	"github.com/ontio/multi-chain/native/service/governance/side_chain_manager"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	crosscommon "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/utils"
)

type BTCHandler struct {
}

func NewBTCHandler() *BTCHandler {
	return &BTCHandler{}
}

func (this *BTCHandler) MultiSign(service *native.NativeService) error {
	params := new(crosscommon.MultiSignParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return fmt.Errorf("MultiSign, contract params deserialize error: %v", err)
	}
	multiSignInfo, err := getBtcMultiSignInfo(service, params.TxHash)
	if err != nil {
		return fmt.Errorf("MultiSign, getBtcMultiSignInfo error: %v", err)
	}

	_, ok := multiSignInfo.MultiSignInfo[params.Address]
	if ok {
		return fmt.Errorf("MultiSign, address %s already sign", params.Address)
	}

	redeemScript, err := side_chain_manager.GetBtcRedeemScriptBytes(service, params.RedeemKey)
	if err != nil {
		return fmt.Errorf("MultiSign, get btc redeem script with redeem key %v from db error: %v", params.RedeemKey, err)
	}

	_, addrs, n, err := txscript.ExtractPkScriptAddrs(redeemScript, netParam)
	if err != nil {
		return fmt.Errorf("MultiSign, failed to extract pkscript addrs: %v", err)
	}
	if len(multiSignInfo.MultiSignInfo) == n {
		return fmt.Errorf("MultiSign, already enough signature: %d", n)
	}

	txb, err := service.GetCacheDB().Get(utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(BTC_TX_PREFIX),
		params.TxHash))
	if err != nil {
		return fmt.Errorf("MultiSign, failed to get tx %s from cacheDB: %v", hex.EncodeToString(params.TxHash), err)
	}
	mtx := wire.NewMsgTx(wire.TxVersion)
	err = mtx.BtcDecode(bytes.NewBuffer(txb), wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return fmt.Errorf("MultiSign, failed to decode tx: %v", err)
	}

	pkScripts := make([][]byte, len(mtx.TxIn))
	for i, in := range mtx.TxIn {
		pkScripts[i] = in.SignatureScript
		in.SignatureScript = nil
	}
	amts, stxos, err := getStxoAmts(service, params.ChainID, mtx.TxIn, params.RedeemKey)
	if err != nil {
		return fmt.Errorf("MultiSign, failed to get stxos: %v", err)
	}
	err = verifySigs(params.Signs, params.Address, addrs, redeemScript, mtx, pkScripts, amts)
	if err != nil {
		return fmt.Errorf("MultiSign, failed to verify: %v", err)
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
		err = addSigToTx(multiSignInfo, addrs, redeemScript, mtx, pkScripts)
		if err != nil {
			return fmt.Errorf("MultiSign, failed to add sig to tx: %v", err)
		}
		var buf bytes.Buffer
		err = mtx.BtcEncode(&buf, wire.ProtocolVersion, wire.LatestEncoding)
		if err != nil {
			return fmt.Errorf("MultiSign, failed to encode msgtx to bytes: %v", err)
		}

		witScript, err := getLockScript(redeemScript)
		if err != nil {
			return fmt.Errorf("MultiSign, failed to get lock script: %v", err)
		}
		utxos, err := getUtxos(service, params.ChainID, params.RedeemKey)
		if err != nil {
			return fmt.Errorf("MultiSign, getUtxos error: %v", err)
		}
		txid := mtx.TxHash()
		for i, v := range mtx.TxOut {
			if bytes.Equal(witScript, v.PkScript) {
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
		putUtxos(service, params.ChainID, params.RedeemKey, utxos)
		btcFromTxInfo, err := getBtcFromInfo(service, params.TxHash)
		if err != nil {
			return fmt.Errorf("MultiSign, failed to get from tx hash %s from cacheDB: %v",
				hex.EncodeToString(params.TxHash), err)
		}
		putStxos(service, params.ChainID, params.RedeemKey, stxos)
		service.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: utils.CrossChainManagerContractAddress,
				States: []interface{}{"btcTxToRelay", btcFromTxInfo.FromChainID, side_chain_manager.BTC_CHAIN_ID,
					hex.EncodeToString(buf.Bytes()), hex.EncodeToString(btcFromTxInfo.FromTxHash), params.RedeemKey},
			})
	}
	return nil
}

func (this *BTCHandler) MakeDepositProposal(service *native.NativeService) (*crosscommon.MakeTxParam, error) {
	params := new(crosscommon.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return nil, fmt.Errorf("btc MakeDepositProposal, contract params deserialize error: %v", err)
	}
	if len(params.Proof) == 0 || len(params.Extra) == 0 {
		return nil, fmt.Errorf("btc MakeDepositProposal, GetInput() data can't be empty")
	}

	value, err := verifyFromBtcTx(service, params.Proof, params.Extra, params.SourceChainID, params.Height)

	if err != nil {
		return nil, fmt.Errorf("MakeDepositProposal, verifyFromBtcTx error: %s", err)
	}

	if err := crosscommon.CheckDoneTx(service, value.TxHash, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("MakeDepositProposal, check done transaction error:%s", err)
	}

	if err := crosscommon.PutDoneTx(service, value.TxHash, params.SourceChainID); err != nil {
		return nil, fmt.Errorf("MakeDepositProposal, PutDoneTx error:%s", err)
	}

	// decode tx and then update utxos
	mtx := wire.NewMsgTx(wire.TxVersion)
	reader := bytes.NewReader(params.Extra)
	err = mtx.BtcDecode(reader, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromBtcProof, failed to decode the transaction %s: %s", hex.EncodeToString(params.Extra), err)
	}
	err = addUtxos(service, params.SourceChainID, service.GetHeight(), mtx)
	if err != nil {
		return nil, fmt.Errorf("btc Vote, updateUtxo error: %s", err)
	}

	return value, nil
}

func (this *BTCHandler) MakeTransaction(service *native.NativeService, param *crosscommon.MakeTxParam,
	fromChainID uint64) error {
	amounts := make(map[string]int64)
	source := common.NewZeroCopySource(param.Args)
	toAddrBytes, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("btc MakeTransaction, deserialize toAddr error")
	}
	amount, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("btc MakeTransaction, deserialize amount error")
	}
	amounts[string(toAddrBytes)] = int64(amount)
	redeemScriptBytes, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("btc MakeTransaction, deserialize redeem script error")
	}

	redeemKey := btcutil.Hash160(redeemScriptBytes)
	contractBind, err := side_chain_manager.GetContractBind(service, side_chain_manager.BTC_CHAIN_ID, fromChainID, redeemKey)
	if err != nil {
		return fmt.Errorf("btc MakeTransaction, side_chain_manager.GetContractBind error: %v", err)
	}
	if contractBind == nil {
		return fmt.Errorf("btc MakeTransaction, contract for %s of chain-id %d is not registered",
			hex.EncodeToString(redeemKey), fromChainID)
	}
	if !bytes.Equal(contractBind.Contract, param.FromContractAddress) {
		return fmt.Errorf("btc MakeTransaction, your contract %s is not match with %s registered",
			hex.EncodeToString(param.FromContractAddress), hex.EncodeToString(contractBind.Contract))
	}
	err = makeBtcTx(service, param.ToChainID, amounts, param.TxHash, fromChainID, redeemScriptBytes, redeemKey)
	if err != nil {
		return fmt.Errorf("btc MakeTransaction, failed to make transaction: %v", err)
	}
	return nil
}

func makeBtcTx(service *native.NativeService, chainID uint64, amounts map[string]int64, fromTxHash []byte,
	fromChainID uint64, redeemScript, rk []byte) error {
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

	// get tx outs
	outs, err := getTxOuts(amounts)
	if err != nil {
		return fmt.Errorf("makeBtcTx, %v", err)
	}

	out, err := getChangeTxOut(0, redeemScript)
	if err != nil {
		return fmt.Errorf("makeBtcTx, %v", err)
	}

	_, addrs, m, _ := txscript.ExtractPkScriptAddrs(redeemScript, netParam)
	choosed, sum, gasFee, err := chooseUtxos(service, chainID, amountSum, append(outs, out), rk, m, len(addrs))
	if err != nil {
		return fmt.Errorf("makeBtcTx, chooseUtxos error: %v", err)
	}
	amts := make([]uint64, len(choosed))
	txIns := make([]*wire.TxIn, len(choosed))
	for i, u := range choosed {
		hash, err := chainhash.NewHash(u.Op.Hash)
		if err != nil {
			return fmt.Errorf("makeBtcTx, chainhash.NewHash error: %v", err)
		}
		txIns[i] = wire.NewTxIn(wire.NewOutPoint(hash, u.Op.Index), u.ScriptPubkey, nil)
		amts[i] = u.Value
	}
	for i := range outs {
		outs[i].Value = outs[i].Value - int64(float64(gasFee)/float64(amountSum)*float64(outs[i].Value))
	}
	out.Value = sum - amountSum
	mtx, err := getUnsignedTx(txIns, outs, out, nil)
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

	btcFromInfo := &BtcFromInfo{
		FromTxHash:  fromTxHash,
		FromChainID: fromChainID,
	}
	if err = putBtcFromInfo(service, txHash[:], btcFromInfo); err != nil {
		return fmt.Errorf("makeBtcTx, putBtcFromInfo failed: %v", err)
	}
	service.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{"makeBtcTx", hex.EncodeToString(rk), hex.EncodeToString(buf.Bytes()), amts},
		})

	return nil
}
