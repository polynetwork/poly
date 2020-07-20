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
package btc

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	wire_bch "github.com/gcash/bchd/wire"
	"github.com/gcash/bchutil/merkleblock"
	"github.com/polynetwork/poly/common"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	crosscommon "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/header_sync/btc"
	"github.com/polynetwork/poly/native/service/utils"
	"golang.org/x/crypto/ripemd160"
)

const (
	OP_RETURN_SCRIPT_FLAG   = byte(0xcc)
	BTC_TX_PREFIX           = "btctx"
	BTC_FROM_TX_PREFIX      = "btcfromtx"
	UTXOS                   = "utxos"
	STXOS                   = "stxos"
	MULTI_SIGN_INFO         = "multiSignInfo"
	MAX_FEE_COST_PERCENTS   = 1.0
	MAX_SELECTING_TRY_LIMIT = 1000000
	SELECTING_K             = 4.0
)

func getNetParam(service *native.NativeService, chainId uint64) (*chaincfg.Params, error) {
	side, err := side_chain_manager.GetSideChain(service, chainId)
	if err != nil {
		return nil, fmt.Errorf("failed to get bitcoin net parameter: %v", err)
	}
	if side == nil {
		return nil, fmt.Errorf("side chain info for chainId: %d is not registered", chainId)
	}
	if side.CCMCAddress == nil || len(side.CCMCAddress) != 8 {
		return nil, fmt.Errorf("CCMCAddress is nil or its length is not 8")
	}
	switch utils.BtcNetType(binary.LittleEndian.Uint64(side.CCMCAddress)) {
	case utils.TyTestnet3:
		return &chaincfg.TestNet3Params, nil
	case utils.TyRegtest:
		return &chaincfg.RegressionNetParams, nil
	case utils.TySimnet:
		return &chaincfg.SimNetParams, nil
	default:
		return &chaincfg.MainNetParams, nil
	}
}

func verifyFromBtcTx(native *native.NativeService, proof, tx []byte, fromChainID uint64, height uint32) (*crosscommon.MakeTxParam, error) {
	// decode tx
	mtx := wire.NewMsgTx(wire.TxVersion)
	reader := bytes.NewReader(tx)
	err := mtx.BtcDecode(reader, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromBtcProof, failed to decode the transaction %s: %s", hex.EncodeToString(tx), err)
	}
	// check tx is legal format for btc cross chain transaction
	err = ifCanResolve(mtx.TxOut[1], mtx.TxOut[0].Value)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromBtcProof, not crosschain btc tx, since failed to resolve parameter: %v", err)
	}

	// make sure the header with height is already synced, meaning the tx is already confirmed in btc block chain
	bestHeader, err := btc.GetBestBlockHeader(native, fromChainID)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromBtcProof, get best block header error:%s", err)
	}
	sideChain, err := side_chain_manager.GetSideChain(native, fromChainID)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromBtcProof, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain == nil {
		return nil, fmt.Errorf("VerifyFromBtcProof, side chain is not registered")
	}
	bestHeight := bestHeader.Height
	if bestHeight < height || bestHeight-height < uint32(sideChain.BlocksToWait-1) {
		return nil, fmt.Errorf("verifyFromBtcTx, transaction is not confirmed, current height: %d, input height: %d", bestHeight, height)
	}

	// verify btc merkle proof
	header, err := btc.GetHeaderByHeight(native, fromChainID, height)
	if err != nil {
		return nil, fmt.Errorf("VerifyFromBtcProof, get header at height %d to verify btc merkle proof error:%s", height, err)
	}
	if verified, err := verifyBtcMerkleProof(mtx, header.Header, proof); !verified {
		return nil, fmt.Errorf("VerifyFromBtcProof, verify merkle proof error:%s", err)
	}

	// decode the extra data from tx and construct MakeTxParam
	var p targetChainParam
	err = p.resolve(mtx.TxOut[0].Value, mtx.TxOut[1])
	if err != nil {
		return nil, fmt.Errorf("verifyFromBtcTx, failed to resolve parameter: %v", err)
	}
	rk := GetUtxoKey(mtx.TxOut[0].PkScript)
	redeemKey, err := hex.DecodeString(rk)
	if err != nil {
		return nil, fmt.Errorf("verifyFromBtcTx, hex.DecodeString error: %v", err)
	}
	toContractAddress, err := side_chain_manager.GetContractBind(native, fromChainID, p.args.ToChainID, redeemKey)
	if err != nil {
		return nil, fmt.Errorf("verifyFromBtcTx, side_chain_manager.GetContractBind error: %v", err)
	}
	if toContractAddress == nil {
		return nil, fmt.Errorf("verifyFromBtcTx, no contract binding with redeem key %s", rk)
	}
	txHash := mtx.TxHash()
	return &crosscommon.MakeTxParam{
		TxHash:              txHash[:],
		CrossChainID:        txHash[:],
		FromContractAddress: redeemKey,
		ToChainID:           p.args.ToChainID,
		ToContractAddress:   toContractAddress.Contract,
		Method:              "unlock",
		Args:                p.AddrAndVal,
	}, nil
}

func verifyBtcMerkleProof(mtx *wire.MsgTx, blockHeader wire.BlockHeader, proof []byte) (bool, error) {
	merkleBlockMsg := wire_bch.MsgMerkleBlock{}
	err := merkleBlockMsg.BchDecode(bytes.NewReader(proof), wire_bch.ProtocolVersion, wire_bch.LatestEncoding)
	if err != nil {
		return false, fmt.Errorf("verify, failed to decode proof: %v", err)
	}
	merkleBlock := merkleblock.NewMerkleBlockFromMsg(merkleBlockMsg)
	merkleRootCalc := merkleBlock.ExtractMatches()
	if merkleRootCalc == nil || merkleBlock.BadTree() || len(merkleBlock.GetMatches()) == 0 {
		return false, fmt.Errorf("verify, bad merkle tree")
	}
	if !bytes.Equal(merkleRootCalc[:], blockHeader.MerkleRoot[:]) {
		return false, fmt.Errorf("verify, merkle root not equal, merkle root should be %s not %s, block hash in proof is %s",
			blockHeader.MerkleRoot.String(), merkleRootCalc.String(), merkleBlockMsg.Header.BlockHash().String())
	}

	// make sure txid exists in proof
	txid := mtx.TxHash()

	isExist := false
	for _, hash := range merkleBlockMsg.Hashes {
		if bytes.Equal(hash[:], txid[:]) {
			isExist = true
			break
		}
	}
	if !isExist {
		return false, fmt.Errorf("verify, transaction %s not found in proof", txid.String())
	}

	return true, nil

}

// not sure now
type targetChainParam struct {
	args       *Args
	AddrAndVal []byte
}

// func about OP_RETURN
func (p *targetChainParam) resolve(amount int64, paramOutput *wire.TxOut) error {
	script := paramOutput.PkScript

	if script[2] != OP_RETURN_SCRIPT_FLAG {
		return errors.New("Wrong flag")
	}
	inputArgs := new(Args)
	err := inputArgs.Deserialization(common.NewZeroCopySource(script[3:]))
	if err != nil {
		return fmt.Errorf("inputArgs.Deserialization fail: %v", err)
	}
	p.args = inputArgs

	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes(inputArgs.Address)
	sink.WriteUint64(uint64(amount))
	p.AddrAndVal = sink.Bytes()

	return nil
}

// This function needs to input the input and output information of the transaction
// and the lock time. Function build a raw transaction without signature and return it.
// This function uses the partial logic and code of btcd to finally return the
// reference of the transaction object.
func getUnsignedTx(txIns []*wire.TxIn, outs []*wire.TxOut, changeOut *wire.TxOut, locktime *int64) (*wire.MsgTx, error) {
	if locktime != nil && (*locktime < 0 || *locktime > int64(wire.MaxTxInSequenceNum)) {
		return nil, fmt.Errorf("getUnsignedTx, locktime %d out of range", *locktime)
	}

	// Add all transaction inputs to a new transaction after performing
	// some validity checks.
	mtx := wire.NewMsgTx(wire.TxVersion)
	for _, in := range txIns {
		if locktime != nil && *locktime != 0 {
			in.Sequence = wire.MaxTxInSequenceNum - 1
		}
		mtx.AddTxIn(in)
	}
	for _, out := range outs {
		mtx.AddTxOut(out)
	}
	if changeOut.Value > 0 {
		mtx.AddTxOut(changeOut)
	}
	// Set the Locktime, if given.
	if locktime != nil {
		mtx.LockTime = uint32(*locktime)
	}

	return mtx, nil
}

func getTxOuts(amounts map[string]int64, netParam *chaincfg.Params) ([]*wire.TxOut, error) {
	outs := make([]*wire.TxOut, 0)
	for encodedAddr, amount := range amounts {
		// Decode the provided address.
		addr, err := btcutil.DecodeAddress(encodedAddr, netParam)
		if err != nil {
			return nil, fmt.Errorf("getTxOuts, decode addr fail: %v", err)
		}

		if !addr.IsForNet(netParam) {
			return nil, fmt.Errorf("getTxOuts, addr is not for %s", netParam.Name)
		}

		// Create a new script which pays to the provided address.
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, fmt.Errorf("getTxOuts, failed to generate pay-to-address script: %v", err)
		}

		txOut := wire.NewTxOut(amount, pkScript)
		outs = append(outs, txOut)
	}

	return outs, nil
}

func getLockScript(redeem []byte, netParam *chaincfg.Params) ([]byte, error) {
	hasher := sha256.New()
	hasher.Write(redeem)
	witAddr, err := btcutil.NewAddressWitnessScriptHash(hasher.Sum(nil), netParam)
	if err != nil {
		return nil, fmt.Errorf("getChangeTxOut, failed to get witness address: %v", err)
	}
	script, err := txscript.PayToAddrScript(witAddr)
	if err != nil {
		return nil, fmt.Errorf("getChangeTxOut, failed to get p2sh script: %v", err)
	}
	return script, nil
}

func GetUtxoKey(scriptPk []byte) string {
	switch txscript.GetScriptClass(scriptPk) {
	case txscript.MultiSigTy:
		return hex.EncodeToString(btcutil.Hash160(scriptPk))
	case txscript.ScriptHashTy:
		return hex.EncodeToString(scriptPk[2:22])
	case txscript.WitnessV0ScriptHashTy:
		hasher := ripemd160.New()
		hasher.Write(scriptPk[2:34])
		return hex.EncodeToString(hasher.Sum(nil))
	default:
		return ""
	}
}

func addUtxos(native *native.NativeService, chainID uint64, height uint32, mtx *wire.MsgTx) error {
	utxoKey := GetUtxoKey(mtx.TxOut[0].PkScript)

	utxos, err := getUtxos(native, chainID, utxoKey)
	if err != nil {
		return fmt.Errorf("addUtxos, getUtxos err:%v", err)
	}
	txHash := mtx.TxHash()
	op := &OutPoint{
		Hash:  txHash[:],
		Index: 0,
	}
	newUtxo := &Utxo{
		Op:           op,
		AtHeight:     height,
		Value:        uint64(mtx.TxOut[0].Value),
		ScriptPubkey: mtx.TxOut[0].PkScript,
	}

	utxos.Utxos = append(utxos.Utxos, newUtxo)
	putUtxos(native, chainID, utxoKey, utxos)
	return nil
}

func chooseUtxos(native *native.NativeService, chainID uint64, amount int64, outs []*wire.TxOut, rk []byte, m, n int) ([]*Utxo, int64, int64, error) {
	utxoKey := hex.EncodeToString(rk)
	utxos, err := getUtxos(native, chainID, utxoKey)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("chooseUtxos, getUtxos error: %v", err)
	}
	sort.Sort(sort.Reverse(utxos))
	detail, err := side_chain_manager.GetBtcTxParam(native, rk, chainID)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("chooseUtxos, failed to get btcTxParam: %v", err)
	}
	if detail == nil {
		return nil, 0, 0, fmt.Errorf("chooseUtxos, no btcTxParam is set for redeem key %s", hex.EncodeToString(rk))
	}
	cs := &CoinSelector{
		sortedUtxos: utxos,
		target:      uint64(amount),
		maxP:        MAX_FEE_COST_PERCENTS,
		tries:       MAX_SELECTING_TRY_LIMIT,
		mc:          detail.MinChange,
		k:           SELECTING_K,
		txOuts:      outs,
		feeRate:     detail.FeeRate,
		m:           m,
		n:           n,
	}
	result, sum, fee := cs.Select()
	if result == nil || len(result) == 0 {
		return nil, 0, 0, fmt.Errorf("chooseUtxos, current utxo is not enough")
	}
	stxos, err := getStxos(native, chainID, utxoKey)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("chooseUtxos, failed to get stxos: %v", err)
	}
	stxos.Utxos = append(stxos.Utxos, result...)
	putStxos(native, chainID, utxoKey, stxos)

	toSort := new(Utxos)
	toSort.Utxos = result
	sort.Sort(sort.Reverse(toSort))
	idx := 0
	for _, v := range toSort.Utxos {
		for utxos.Utxos[idx].Op.String() != v.Op.String() {
			idx++
		}
		utxos.Utxos = append(utxos.Utxos[:idx], utxos.Utxos[idx+1:]...)
	}
	putUtxos(native, chainID, utxoKey, utxos)
	return result, int64(sum), int64(fee), nil
}

func putTxos(k string, native *native.NativeService, chainID uint64, txoKey string, txos *Utxos) {
	chainIDBytes := utils.GetUint64Bytes(chainID)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(k), chainIDBytes, []byte(txoKey))
	sink := common.NewZeroCopySink(nil)
	txos.Serialization(sink)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
}

func getTxos(k string, native *native.NativeService, chainID uint64, txoKey string) (*Utxos, error) {
	chainIDBytes := utils.GetUint64Bytes(chainID)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(k), chainIDBytes, []byte(txoKey))
	store, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("get%s, get btcTxStore error: %v", k, err)
	}
	txos := &Utxos{
		Utxos: make([]*Utxo, 0),
	}
	if store != nil {
		utxosBytes, err := cstates.GetValueFromRawStorageItem(store)
		if err != nil {
			return nil, fmt.Errorf("get%s, deserialize from raw storage item err:%v", k, err)
		}
		err = txos.Deserialization(common.NewZeroCopySource(utxosBytes))
		if err != nil {
			return nil, fmt.Errorf("get%s, utxos.Deserialization err:%v", k, err)
		}
	}
	return txos, nil
}

func putUtxos(native *native.NativeService, chainID uint64, utxoKey string, utxos *Utxos) {
	putTxos(UTXOS, native, chainID, utxoKey, utxos)
}

func getUtxos(native *native.NativeService, chainID uint64, utxoKey string) (*Utxos, error) {
	utxos, err := getTxos(UTXOS, native, chainID, utxoKey)
	return utxos, err
}

func putStxos(native *native.NativeService, chainID uint64, stxoKey string, stxos *Utxos) {
	putTxos(STXOS, native, chainID, stxoKey, stxos)
}

func getStxos(native *native.NativeService, chainID uint64, stxoKey string) (*Utxos, error) {
	stxos, err := getTxos(STXOS, native, chainID, stxoKey)
	return stxos, err
}

func getStxoAmts(service *native.NativeService, chainID uint64, txIns []*wire.TxIn, redeemKey string) ([]uint64, *Utxos, error) {
	stxos, err := getStxos(service, chainID, redeemKey)
	if err != nil {
		return nil, nil, fmt.Errorf("getStxoAmts, failed to get stxos: %v", err)
	}
	amts := make([]uint64, len(txIns))
	for i, in := range txIns {
		toDel := -1
		for j, v := range stxos.Utxos {
			if bytes.Equal(in.PreviousOutPoint.Hash[:], v.Op.Hash) && in.PreviousOutPoint.Index == v.Op.Index {
				amts[i] = v.Value
				toDel = j
				break
			}
		}
		if toDel < 0 {
			return nil, nil, fmt.Errorf("getStxoAmts, %d txIn not found in stxos", i)
		}
		stxos.Utxos = append(stxos.Utxos[:toDel], stxos.Utxos[toDel+1:]...)
	}

	return amts, stxos, nil
}

func verifySigs(sigs [][]byte, addr string, addrs []btcutil.Address, redeem []byte, tx *wire.MsgTx,
	pkScripts [][]byte, amts []uint64) error {
	if len(sigs) != len(tx.TxIn) {
		return fmt.Errorf("not enough sig, only %d sigs but %d required", len(sigs), len(tx.TxIn))
	}
	var signerAddr btcutil.Address = nil
	for _, a := range addrs {
		if a.EncodeAddress() == addr {
			signerAddr = a
		}
	}

	if signerAddr == nil {
		return fmt.Errorf("address %s not found in redeem script", addr)
	}

	for i, sig := range sigs {
		if len(sig) < 1 {
			return fmt.Errorf("length of no.%d sig is less than 1", i)
		}
		tSig := sig[:len(sig)-1]
		pSig, err := btcec.ParseDERSignature(tSig, btcec.S256())
		if err != nil {
			return fmt.Errorf("failed to parse no.%d sig: %v", i, err)
		}
		var hash []byte
		switch c := txscript.GetScriptClass(pkScripts[i]); c {
		case txscript.MultiSigTy, txscript.ScriptHashTy:
			hash, err = txscript.CalcSignatureHash(redeem, txscript.SigHashType(sig[len(sig)-1]), tx, i)
			if err != nil {
				return fmt.Errorf("failed to calculate sig hash: %v", err)
			}
		case txscript.WitnessV0ScriptHashTy:
			sh := txscript.NewTxSigHashes(tx)
			hash, err = txscript.CalcWitnessSigHash(redeem, sh, txscript.SigHashType(sig[len(sig)-1]), tx, i, int64(amts[i]))
			if err != nil {
				return fmt.Errorf("failed to calculate sig hash: %v", err)
			}
		default:
			return fmt.Errorf("script %s not supported", c)
		}
		if !pSig.Verify(hash, signerAddr.(*btcutil.AddressPubKey).PubKey()) {
			return fmt.Errorf("verify no.%d sig and not pass", i+1)
		}
	}

	return nil
}

func putBtcMultiSignInfo(native *native.NativeService, txid []byte, multiSignInfo *MultiSignInfo) error {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(MULTI_SIGN_INFO), txid)
	sink := common.NewZeroCopySink(nil)
	multiSignInfo.Serialization(sink)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getBtcMultiSignInfo(native *native.NativeService, txid []byte) (*MultiSignInfo, error) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(MULTI_SIGN_INFO), txid)
	multiSignInfoStore, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("getBtcMultiSignInfo, get multiSignInfoStore error: %v", err)
	}

	multiSignInfo := &MultiSignInfo{
		MultiSignInfo: make(map[string][][]byte),
	}
	if multiSignInfoStore != nil {
		multiSignInfoBytes, err := cstates.GetValueFromRawStorageItem(multiSignInfoStore)
		if err != nil {
			return nil, fmt.Errorf("getBtcMultiSignInfo, deserialize from raw storage item err:%v", err)
		}
		err = multiSignInfo.Deserialization(common.NewZeroCopySource(multiSignInfoBytes))
		if err != nil {
			return nil, fmt.Errorf("getBtcMultiSignInfo, deserialize multiSignInfo err:%v", err)
		}
	}
	return multiSignInfo, nil
}

func addSigToTx(sigMap *MultiSignInfo, addrs []btcutil.Address, redeem []byte, tx *wire.MsgTx, pkScripts [][]byte) error {
	for i := 0; i < len(tx.TxIn); i++ {
		var (
			script []byte
			err    error
		)
		builder := txscript.NewScriptBuilder()
		switch c := txscript.GetScriptClass(pkScripts[i]); c {
		case txscript.MultiSigTy, txscript.ScriptHashTy:
			builder.AddOp(txscript.OP_FALSE)
			for _, addr := range addrs {
				signs, ok := sigMap.MultiSignInfo[addr.EncodeAddress()]
				if !ok {
					continue
				}
				val := signs[i]
				builder.AddData(val)
			}
			if c == txscript.ScriptHashTy {
				builder.AddData(redeem)
			}
			script, err = builder.Script()
			if err != nil {
				return fmt.Errorf("failed to build sigscript for input %d: %v", i, err)
			}
			tx.TxIn[i].SignatureScript = script
		case txscript.WitnessV0ScriptHashTy:
			data := make([][]byte, len(sigMap.MultiSignInfo)+2)
			idx := 1
			for _, addr := range addrs {
				signs, ok := sigMap.MultiSignInfo[addr.EncodeAddress()]
				if !ok {
					continue
				}
				data[idx] = signs[i]
				idx++
			}
			data[idx] = redeem
			tx.TxIn[i].Witness = wire.TxWitness(data)
		default:
			return fmt.Errorf("addSigToTx, type of no.%d utxo is %s which is not supported", i, c)
		}
	}
	return nil
}

func putBtcFromInfo(native *native.NativeService, txid []byte, btcFromInfo *BtcFromInfo) error {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(BTC_FROM_TX_PREFIX), txid)
	sink := common.NewZeroCopySink(nil)
	btcFromInfo.Serialization(sink)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getBtcFromInfo(native *native.NativeService, txid []byte) (*BtcFromInfo, error) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(BTC_FROM_TX_PREFIX), txid)
	btcFromInfoStore, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("getBtcFromInfo, get multiSignInfoStore error: %v", err)
	}
	btcFromInfo := new(BtcFromInfo)
	if btcFromInfoStore == nil {
		return nil, fmt.Errorf("getBtcFromInfo, can not find any record")
	}
	multiSignInfoBytes, err := cstates.GetValueFromRawStorageItem(btcFromInfoStore)
	if err != nil {
		return nil, fmt.Errorf("getBtcFromInfo, deserialize from raw storage item err:%v", err)
	}
	err = btcFromInfo.Deserialization(common.NewZeroCopySource(multiSignInfoBytes))
	if err != nil {
		return nil, fmt.Errorf("getBtcFromInfo, deserialize multiSignInfo err:%v", err)
	}
	return btcFromInfo, nil
}

func ifCanResolve(paramOutput *wire.TxOut, value int64) error {
	script := paramOutput.PkScript
	if script[2] != OP_RETURN_SCRIPT_FLAG {
		return errors.New("wrong flag")
	}
	args := Args{}
	err := args.Deserialization(common.NewZeroCopySource(script[3:]))
	if err != nil {
		return err
	}
	if value < args.Fee && args.Fee >= 0 {
		return errors.New("the transfer amount cannot be less than the transaction fee")
	}
	return nil
}
