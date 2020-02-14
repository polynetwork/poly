package btc

import (
	"bytes"
	"crypto/sha256"
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
	"github.com/ontio/multi-chain/common"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native"
	crosscommon "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/governance/side_chain_manager"
	"github.com/ontio/multi-chain/native/service/header_sync/btc"
	"github.com/ontio/multi-chain/native/service/utils"
	"golang.org/x/crypto/ripemd160"
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
	MAX_FEE_COST_PERCENTS                   = 0.4
	MAX_SELECTING_TRY_LIMIT                 = 1000000
	SELECTING_K                             = 2.0
	BTC_CHAIN_ID                            = 1
)

var netParam = &chaincfg.TestNet3Params

func GetNetParam() *chaincfg.Params {
	return netParam
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
	fromContractAddress, err := hex.DecodeString(GetUtxoKey(mtx.TxOut[0].PkScript))
	if err != nil {
		return nil, fmt.Errorf("verifyFromBtcTx, hex.DecodeString error: %v", err)
	}
	txHash := mtx.TxHash()
	return &crosscommon.MakeTxParam{
		TxHash:              txHash[:],
		CrossChainID:        txHash[:],
		FromContractAddress: fromContractAddress,
		ToChainID:           p.args.ToChainID,
		ToContractAddress:   p.args.ToContractAddress,
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

func GetUtxoKey(scriptPk []byte) string {
	switch txscript.GetScriptClass(scriptPk) {
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

func chooseUtxos(native *native.NativeService, chainID uint64, amount int64, outs []*wire.TxOut) ([]*Utxo, int64, int64, error) {
	utxoKey := GetUtxoKey(outs[len(outs)-1].PkScript)
	utxos, err := getUtxos(native, chainID, utxoKey)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("chooseUtxos, getUtxos error: %v", err)
	}
	sort.Sort(sort.Reverse(utxos))
	cs := &CoinSelector{
		SortedUtxos: utxos,
		Target:      uint64(amount),
		MaxP:        MAX_FEE_COST_PERCENTS,
		Tries:       MAX_SELECTING_TRY_LIMIT,
		Mc:          MIN_CHANGE,
		K:           SELECTING_K,
		TxOuts:      outs,
		feeWeight:   WEIGHT,
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

func putBtcRedeemScript(native *native.NativeService, redeemScriptKey string, redeemScriptBytes []byte) error {
	chainIDBytes := utils.GetUint64Bytes(0)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(REDEEM_SCRIPT), chainIDBytes, []byte(redeemScriptKey))

	cls := txscript.GetScriptClass(redeemScriptBytes)
	if cls.String() != "multisig" {
		return fmt.Errorf("putBtcRedeemScript, wrong type of redeem: %s", cls)
	}
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(redeemScriptBytes))
	return nil
}

func getBtcRedeemScript(native *native.NativeService, redeemScriptKey string) (string, error) {
	redeem, err := getBtcRedeemScriptBytes(native, redeemScriptKey)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(redeem), nil
}

func getBtcRedeemScriptBytes(native *native.NativeService, redeemScriptKey string) ([]byte, error) {
	chainIDBytes := utils.GetUint64Bytes(0)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(REDEEM_SCRIPT), chainIDBytes, []byte(redeemScriptKey))
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
