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
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	wire_bch "github.com/gcash/bchd/wire"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	crosscommon "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/governance/side_chain_manager"
	"github.com/ontio/multi-chain/native/service/utils"
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

	multiSignInfo, err := getBtcMultiSignInfo(service, params.TxHash)
	if err != nil {
		return fmt.Errorf("MultiSign, getBtcMultiSignInfo error: %v", err)
	}
	_, ok := multiSignInfo.MultiSignInfo[params.Address]
	if ok {
		return fmt.Errorf("MultiSign, address %s already sign", params.Address)
	}

	redeemScript, err := getBtcRedeemScriptBytes(service)
	if err != nil {
		return fmt.Errorf("MultiSign, getBtcRedeemScript error: %v", err)
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
	amts, stxos, err := getStxoAmts(service, params.ChainID, mtx.TxIn)
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
		utxos, err := getUtxos(service, params.ChainID)
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
		putUtxos(service, params.ChainID, utxos)
		btcFromTxInfo, err := getBtcFromInfo(service, params.TxHash)
		if err != nil {
			return fmt.Errorf("MultiSign, failed to get from tx hash %s from cacheDB: %v",
				hex.EncodeToString(params.TxHash), err)
		}
		putStxos(service, params.ChainID, stxos)
		service.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: utils.CrossChainManagerContractAddress,
				States: []interface{}{"btcTxToRelay", btcFromTxInfo.FromChainID, 0, hex.EncodeToString(buf.Bytes()),
					hex.EncodeToString(btcFromTxInfo.FromTxHash)},
			})
	}
	return nil
}

func (this *BTCHandler) Vote(service *native.NativeService) (bool, *crosscommon.MakeTxParam, uint64, error) {
	params := new(crosscommon.VoteParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return false, nil, 0, fmt.Errorf("btc Vote, contract params deserialize error: %v", err)
	}

	address, err := common.AddressFromBase58(params.Address)
	if err != nil {
		return false, nil, 0, fmt.Errorf("btc Vote, common.AddressFromBase58 error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(service, address)
	if err != nil {
		return false, nil, 0, fmt.Errorf("btc Vote, utils.ValidateOwner error: %v", err)
	}

	vote, err := getBtcVote(service, params.TxHash)
	if err != nil {
		return false, nil, 0, fmt.Errorf("btc Vote, getBtcVote error: %v", err)
	}
	_, ok := vote.VoteMap[params.Address]
	if ok {
		return false, nil, 0, fmt.Errorf("btc Vote, address %s already voted", params.Address)
	}
	vote.VoteMap[params.Address] = params.Address
	err = putBtcVote(service, params.TxHash, vote)
	if err != nil {
		return false, nil, 0, fmt.Errorf("btc Vote, putBtcVote error: %v", err)
	}
	service.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{"Vote", vote.VoteMap},
		})

	err = crosscommon.ValidateVote(service, vote)
	if err != nil {
		return false, nil, 0, nil
	}

	proofBytes, err := getBtcProof(service, params.TxHash)
	if err != nil {
		return false, nil, 0, fmt.Errorf("btc Vote, getBtcTx error: %v", err)
	}
	proof := new(BtcProof)
	err = proof.Deserialization(common.NewZeroCopySource(proofBytes))
	if err != nil {
		return false, nil, 0, fmt.Errorf("btc Vote, proof.Deserialization error: %v", err)
	}

	mtx := wire.NewMsgTx(wire.TxVersion)
	reader := bytes.NewReader(proof.Tx)
	err = mtx.BtcDecode(reader, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return false, nil, 0, fmt.Errorf("btc Vote, failed to decode the transaction")
	}

	err = addUtxos(service, params.FromChainID, service.GetHeight(), mtx)
	if err != nil {
		return false, nil, 0, fmt.Errorf("btc Vote, updateUtxo error: %s", err)
	}

	var p targetChainParam
	err = p.resolve(mtx.TxOut[0].Value, mtx.TxOut[1])
	if err != nil {
		return false, nil, 0, fmt.Errorf("btc Vote, failed to resolve parameter: %v", err)
	}

	txHash := mtx.TxHash()
	return true, &crosscommon.MakeTxParam{
		TxHash:              txHash[:],
		FromContractAddress: []byte(BTC_ADDRESS),
		ToChainID:           p.args.ToChainID,
		ToContractAddress:   p.args.ToContractAddress,
		Method:              "unlock",
		Args:                p.AddrAndVal,
	}, params.FromChainID, nil
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

func (this *BTCHandler) MakeTransaction(service *native.NativeService, param *crosscommon.MakeTxParam,
	fromChainID uint64) error {
	if !bytes.Equal(param.ToContractAddress, []byte(BTC_ADDRESS)) {
		return fmt.Errorf("btc MakeTransaction, destContractAddr is %s not btc", string(param.ToContractAddress))
	}
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

	err := makeBtcTx(service, param.ToChainID, amounts, param.TxHash, fromChainID)
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
	if sideChain == nil {
		return fmt.Errorf("notifyBtcTx, side chain is not registered")
	}

	mtx := wire.NewMsgTx(wire.TxVersion)
	reader := bytes.NewReader(tx)
	err = mtx.BtcDecode(reader, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return fmt.Errorf("notifyBtcTx, failed to decode the transaction %s: %s", hex.EncodeToString(tx), err)
	}
	redeem, err := getBtcRedeemScriptBytes(native)
	if err != nil {
		return fmt.Errorf("notifyBtcTx, failed to get redeem: %v", err)
	}
	err = checkTxOuts(mtx, redeem)
	if err != nil {
		return fmt.Errorf("notifyBtcTx, wrong outputs: %v", err)
	}

	err = ifCanResolve(mtx.TxOut[1], mtx.TxOut[0].Value)
	if err != nil {
		return fmt.Errorf("notifyBtcTx, failed to resolve parameter: %v", err)
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

func makeBtcTx(service *native.NativeService, chainID uint64, amounts map[string]int64, fromTxHash []byte,
	fromChainID uint64) error {
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
	redeemScript, err := getBtcRedeemScriptBytes(service)
	if err != nil {
		return fmt.Errorf("makeBtcTx, getBtcRedeemScript error: %v", err)
	}
	out, err := getChangeTxOut(0, redeemScript)
	if err != nil {
		return fmt.Errorf("makeBtcTx, %v", err)
	}

	choosed, sum, err := chooseUtxos(service, chainID, amountSum)
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

	gasFee := int64(float64(estimateSerializedTxSize(txIns, outs, out)*MIN_SATOSHI_TO_RELAY_PER_BYTE) * WEIGHT)
	if amountSum <= gasFee {
		return fmt.Errorf("makeBtcTx, amounts sum(%d) must greater than fee %d", amountSum, gasFee)
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
	err = putBtcFromInfo(service, txHash[:], btcFromInfo)
	if err != nil {
		return fmt.Errorf("makeBtcTx, putBtcFromInfo failed: %v", err)
	}

	service.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{"makeBtcTx", hex.EncodeToString(buf.Bytes()), amts},
		})

	return nil
}
