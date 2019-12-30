package btc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	crosscommon "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/utils"
	"sort"
)

const (
	// TODO: Temporary setting
	OP_RETURN_SCRIPT_FLAG                   = byte(0x66)
	BTC_TX_PREFIX                           = "btctx"
	BTC_FROM_TX_PREFIX                      = "btcfromtx"
	REDEEM_P2SH_5_OF_7_MULTISIG_SCRIPT_SIZE = 1 + 5*(1+75) + 1 + 1 + 7*(1+33) + 1 + 1
	MIN_SATOSHI_TO_RELAY_PER_BYTE           = 1
	WEIGHT                                  = 1.2
	MIN_CHANGE                              = 2000
	BTC_ADDRESS                             = "btc"
	NOTIFY_BTC_PROOF                        = "notifyBtcProof"
	UTXOS                                   = "utxos"
	STXOS                                   = "stxos"
	REDEEM_SCRIPT                           = "redeemScript"
	MULTI_SIGN_INFO                         = "multiSignInfo"
	MAX_FEE_COST_PERCENTS = 0.4
	MAX_SELECTING_TRY_LIMIT = 1000000
	SELECTING_K = 2.0
)

var netParam = &chaincfg.TestNet3Params

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

func prefixAppendUint256(src []byte) []byte {
	x := make([]byte, 32)
	for i := 0; i < len(src); i++ {
		x[32-len(src)+i] = byte(src[i])
	}
	return x
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

func getTxOuts(amounts map[string]int64) ([]*wire.TxOut, error) {
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

func getChangeTxOut(change int64, redeem []byte) (*wire.TxOut, error) {
	script, err := getLockScript(redeem)
	return wire.NewTxOut(change, script), err
}

func getLockScript(redeem []byte) ([]byte, error) {
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

func putBtcProof(native *native.NativeService, txHash, proof []byte) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(crosscommon.KEY_PREFIX_BTC), txHash)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(proof))
}

func getBtcProof(native *native.NativeService, txHash []byte) ([]byte, error) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(crosscommon.KEY_PREFIX_BTC), txHash)
	btcProofStore, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("getBtcProof, get btcProofStore error: %v", err)
	}
	if btcProofStore == nil {
		return nil, fmt.Errorf("getBtcProof, can not find any records")
	}
	btcProofBytes, err := cstates.GetValueFromRawStorageItem(btcProofStore)
	if err != nil {
		return nil, fmt.Errorf("getBtcProof, deserialize from raw storage item err:%v", err)
	}
	return btcProofBytes, nil
}

func checkBtcProof(native *native.NativeService, txHash []byte) (bool, error) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(crosscommon.KEY_PREFIX_BTC), txHash)
	btcProofStore, err := native.GetCacheDB().Get(key)
	if err != nil {
		return false, fmt.Errorf("getBtcProof, get btcProofStore error: %v", err)
	}
	if btcProofStore == nil {
		return true, nil
	}
	return false, nil
}

func putBtcVote(native *native.NativeService, txHash []byte, vote *crosscommon.Vote) error {
	sink := common.NewZeroCopySink(nil)
	vote.Serialization(sink)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(crosscommon.KEY_PREFIX_BTC_VOTE), txHash)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getBtcVote(native *native.NativeService, txHash []byte) (*crosscommon.Vote, error) {
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(crosscommon.KEY_PREFIX_BTC_VOTE), txHash)
	btcVoteStore, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("getBtcVote, get btcTxStore error: %v", err)
	}
	vote := &crosscommon.Vote{
		VoteMap: make(map[string]string),
	}
	if btcVoteStore != nil {
		btcVoteBytes, err := cstates.GetValueFromRawStorageItem(btcVoteStore)
		if err != nil {
			return nil, fmt.Errorf("getBtcVote, deserialize from raw storage item err:%v", err)
		}
		err = vote.Deserialization(common.NewZeroCopySource(btcVoteBytes))
		if err != nil {
			return nil, fmt.Errorf("getBtcVote, vote.Deserialization err:%v", err)
		}
	}
	return vote, nil
}

func addUtxos(native *native.NativeService, chainID uint64, height uint32, mtx *wire.MsgTx) error {
	utxos, err := getUtxos(native, chainID)
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
	putUtxos(native, chainID, utxos)
	return nil
}

func chooseUtxos(native *native.NativeService, chainID uint64, amount int64, outs []*wire.TxOut) ([]*Utxo, int64, int64, error) {
	utxos, err := getUtxos(native, chainID)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("chooseUtxos, getUtxos error: %v", err)
	}
	sort.Sort(sort.Reverse(utxos))
	cs := &CoinSelector {
		SortedUtxos: utxos,
		Target: uint64(amount),
		MaxP: MAX_FEE_COST_PERCENTS,
		Tries: MAX_SELECTING_TRY_LIMIT,
		Mc: MIN_CHANGE,
		K: SELECTING_K,
		TxOuts: outs,
	}
	result, sum, fee := cs.Select()
	if result == nil || len(result) == 0 {
		return nil, 0, 0, fmt.Errorf("chooseUtxos, current utxo is not enough")
	}
	stxos, err := getStxos(native, chainID)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("chooseUtxos, failed to get stxos: %v", err)
	}
	stxos.Utxos = append(stxos.Utxos, result...)
	putStxos(native, chainID, stxos)

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
	putUtxos(native, chainID, utxos)
	return result, int64(sum), int64(fee), nil
}

func putTxos(k string, native *native.NativeService, chainID uint64, txos *Utxos) {
	chainIDBytes := utils.GetUint64Bytes(chainID)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(k), chainIDBytes)
	sink := common.NewZeroCopySink(nil)
	txos.Serialization(sink)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
}

func getTxos(k string, native *native.NativeService, chainID uint64) (*Utxos, error) {
	chainIDBytes := utils.GetUint64Bytes(chainID)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(k), chainIDBytes)
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

func putUtxos(native *native.NativeService, chainID uint64, utxos *Utxos) {
	putTxos(UTXOS, native, chainID, utxos)
}

func getUtxos(native *native.NativeService, chainID uint64) (*Utxos, error) {
	utxos, err := getTxos(UTXOS, native, chainID)
	return utxos, err
}

func putStxos(native *native.NativeService, chainID uint64, stxos *Utxos) {
	putTxos(STXOS, native, chainID, stxos)
}

func getStxos(native *native.NativeService, chainID uint64) (*Utxos, error) {
	stxos, err := getTxos(STXOS, native, chainID)
	return stxos, err
}

func getStxoAmts(service *native.NativeService, chainID uint64, txIns []*wire.TxIn) ([]uint64, *Utxos, error) {
	stxos, err := getStxos(service, chainID)
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
			return nil, nil, fmt.Errorf("getStxoAmts, %d txIn not found in stxos: %v", i, err)
		}
		stxos.Utxos = append(stxos.Utxos[:toDel], stxos.Utxos[toDel+1:]...)
	}

	return amts, stxos, nil
}

func putBtcRedeemScript(native *native.NativeService, redeemScript string) error {
	chainIDBytes := utils.GetUint64Bytes(0)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(REDEEM_SCRIPT), chainIDBytes)
	redeem, err := hex.DecodeString(redeemScript)
	if err != nil {
		return fmt.Errorf("putBtcRedeemScript, failed to decode redeem script: %v", err)
	}

	cls := txscript.GetScriptClass(redeem)
	if cls.String() != "multisig" {
		return fmt.Errorf("putBtcRedeemScript, wrong type of redeem: %s", cls)
	}
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(redeem))
	return nil
}

func getBtcRedeemScript(native *native.NativeService) (string, error) {
	redeem, err := getBtcRedeemScriptBytes(native)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(redeem), nil
}

func getBtcRedeemScriptBytes(native *native.NativeService) ([]byte, error) {
	chainIDBytes := utils.GetUint64Bytes(0)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(REDEEM_SCRIPT), chainIDBytes)
	redeemStore, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("getBtcRedeemScript, get btcProofStore error: %v", err)
	}
	if redeemStore == nil {
		return nil, fmt.Errorf("getBtcRedeemScript, can not find any records")
	}
	redeemBytes, err := cstates.GetValueFromRawStorageItem(redeemStore)
	if err != nil {
		return nil, fmt.Errorf("getBtcRedeemScript, deserialize from raw storage item err:%v", err)
	}
	return redeemBytes, nil
}

func notifyBtcProof(native *native.NativeService, txid, btcProof string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{NOTIFY_BTC_PROOF, txid, btcProof},
		})
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
			return fmt.Errorf("verify no.%d sig and not pass", i)
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

func checkTxOuts(tx *wire.MsgTx, redeem []byte) error {
	if len(tx.TxOut) < 2 {
		return errors.New("checkTxOuts, number of transaction's outputs is at least greater" +
			" than 2")
	}
	if tx.TxOut[0].Value <= 0 {
		return fmt.Errorf("checkTxOuts, the value of crosschain transaction must be bigger "+
			"than 0, but value is %d", tx.TxOut[0].Value)
	}

	switch txscript.GetScriptClass(tx.TxOut[0].PkScript) {
	case txscript.MultiSigTy:
		if !bytes.Equal(redeem, tx.TxOut[0].PkScript) {
			return fmt.Errorf("wrong script: \"%x\" is not same as our \"%x\"",
				tx.TxOut[0].PkScript, redeem)
		}
	case txscript.ScriptHashTy:
		addr, err := btcutil.NewAddressScriptHash(redeem, netParam)
		if err != nil {
			return err
		}
		script, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return err
		}
		if !bytes.Equal(script, tx.TxOut[0].PkScript) {
			return fmt.Errorf("wrong script: \"%x\" is not same as our \"%x\"", tx.TxOut[0].PkScript, script)
		}
	case txscript.WitnessV0ScriptHashTy:
		hasher := sha256.New()
		hasher.Write(redeem)
		addr, err := btcutil.NewAddressWitnessScriptHash(hasher.Sum(nil), netParam)
		if err != nil {
			return err
		}
		script, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return err
		}
		if !bytes.Equal(script, tx.TxOut[0].PkScript) {
			return fmt.Errorf("wrong script: \"%x\" is not same as our \"%x\"", tx.TxOut[0].PkScript, script)
		}
	default:
		return errors.New("first output's pkScript is not supported")
	}

	c2 := txscript.GetScriptClass(tx.TxOut[1].PkScript)
	if c2 != txscript.NullDataTy {
		return errors.New("second output's pkScript is not NullData type")
	}

	return nil
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
