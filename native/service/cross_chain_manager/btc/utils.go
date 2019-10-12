package btc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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
)

const (
	// TODO: Temporary setting
	OP_RETURN_DATA_LEN           = 37
	OP_RETURN_SCRIPT_FLAG        = byte(0x66)
	FEE                          = int64(1e3)
	REQUIRE                      = 5
	BTC_TX_PREFIX         string = "btctx"
	IP                    string = "0.0.0.0:30336" //
)

var netParam = &chaincfg.TestNet3Params

// not sure now
type targetChainParam struct {
	ChainId    uint64
	Fee        int64
	AddrAndVal []byte
}

// func about OP_RETURN
func (p *targetChainParam) resolve(amount int64, paramOutput *wire.TxOut) error {
	script := paramOutput.PkScript
	if int(script[1]) != OP_RETURN_DATA_LEN {
		return errors.New("Length of script is wrong")
	}

	if script[2] != OP_RETURN_SCRIPT_FLAG {
		return errors.New("Wrong flag")
	}
	p.ChainId = binary.BigEndian.Uint64(script[3:11])
	p.Fee = int64(binary.BigEndian.Uint64(script[11:19]))
	// TODO:need to check the addr format?
	if amount < p.Fee && p.Fee >= 0 {
		return errors.New("The transfer amount cannot be less than the transaction fee")
	}
	toAddr, err := common.AddressParseFromBytes(script[19:])
	if err != nil {
		return fmt.Errorf("Failed to parse address from bytes: %v", err)
	}
	//sink := common.NewZeroCopySink(nil)
	//sink.WriteVarBytes([]byte(toAddr.ToBase58()))
	//sink.WriteInt64(amount)
	//p.AddrAndVal = sink.Bytes()

	buf := bytes.NewBuffer(nil)
	err = utils.NeoVmSerializeArray(buf, 2)
	if err != nil {
		return fmt.Errorf("btc resolve, utils.NeoVmSerializeArray length error: %v", err)
	}
	err = utils.NeoVmSerializeAddress(buf, toAddr)
	if err != nil {
		return fmt.Errorf("btc resolve, utils.NeoVmSerializeAddress address error: %v", err)
	}
	err = utils.NeoVmSerializeInteger(buf, new(big.Int).SetInt64(amount))
	if err != nil {
		return fmt.Errorf("btc resolve, utils.NeoVmSerializeInteger amount error: %v", err)
	}
	p.AddrAndVal = buf.Bytes()
	return nil
}

// This function needs to input the input and output information of the transaction
// and the lock time. Function build a raw transaction without signature and return it.
// This function uses the partial logic and code of btcd to finally return the
// reference of the transaction object.
func getUnsignedTx(txIns []btcjson.TransactionInput, amounts map[string]int64, change int64, multiScript []byte,
	locktime *int64) (*wire.MsgTx, error) {
	if locktime != nil &&
		(*locktime < 0 || *locktime > int64(wire.MaxTxInSequenceNum)) {
		return nil, fmt.Errorf("getUnsignedTx, locktime %d out of range", *locktime)
	}

	// Add all transaction inputs to a new transaction after performing
	// some validity checks.
	mtx := wire.NewMsgTx(wire.TxVersion)
	for _, input := range txIns {
		txHash, err := chainhash.NewHashFromStr(input.Txid)
		if err != nil {
			return nil, fmt.Errorf("getUnsignedTx, decode txid fail: %v", err)
		}

		prevOut := wire.NewOutPoint(txHash, input.Vout)
		txIn := wire.NewTxIn(prevOut, []byte{}, nil)
		if locktime != nil && *locktime != 0 {
			txIn.Sequence = wire.MaxTxInSequenceNum - 1
		}
		mtx.AddTxIn(txIn)
	}

	// Add all transaction outputs to the transaction after performing
	// some validity checks.
	for encodedAddr, amount := range amounts {
		// Decode the provided address.
		addr, err := btcutil.DecodeAddress(encodedAddr, netParam)
		if err != nil {
			return nil, fmt.Errorf("getUnsignedTx, decode addr fail: %v", err)
		}

		// Ensure the address is one of the supported types and that
		// the network encoded with the address matches the network the
		// server is currently on.
		switch addr.(type) {
		case *btcutil.AddressPubKeyHash:
		case *btcutil.AddressScriptHash:
		default:
			return nil, fmt.Errorf("getUnsignedTx, type of addr is not found")
		}
		if !addr.IsForNet(netParam) {
			return nil, fmt.Errorf("getUnsignedTx, addr is not for mainnet")
		}

		// Create a new script which pays to the provided address.
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, fmt.Errorf("getUnsignedTx, failed to generate pay-to-address script: %v", err)
		}

		txOut := wire.NewTxOut(amount, pkScript)
		mtx.AddTxOut(txOut)
	}

	if change > 0 {
		p2shAddr, err := btcutil.NewAddressScriptHash(multiScript, netParam)
		if err != nil {
			return nil, fmt.Errorf("getRawTxToMultiAddr, failed to get p2sh: %v", err)
		}
		p2shScript, err := txscript.PayToAddrScript(p2shAddr)
		if err != nil {
			return nil, fmt.Errorf("getRawTxToMultiAddr, failed to get p2sh script: %v", err)
		}
		mtx.AddTxOut(wire.NewTxOut(change, p2shScript))
	}

	// Set the Locktime, if given.
	if locktime != nil {
		mtx.LockTime = uint32(*locktime)
	}

	return mtx, nil
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
	err = putUtxos(native, chainID, utxos)
	if err != nil {
		return fmt.Errorf("addUtxos, putUtxos err:%v", err)
	}
	return nil
}

func chooseUtxos(native *native.NativeService, chainID uint64, amount int64, fee int64) ([]*Utxo, int64, error) {
	utxos, err := getUtxos(native, chainID)
	if err != nil {
		return nil, 0, fmt.Errorf("chooseUtxos, getUtxos error: %v", err)
	}
	total := amount + fee
	result := make([]*Utxo, 0)
	var sum int64 = 0
	var j int
	for i := 0; i < len(utxos.Utxos); i++ {
		sum = sum + int64(utxos.Utxos[i].Value)
		result = append(result, utxos.Utxos[i])
		if sum >= total {
			j = i
			break
		}
	}
	if sum < total {
		return nil, sum, fmt.Errorf("chooseUtxos, current utxo sum %d is not enough %d", sum, total)
	}
	utxos.Utxos = utxos.Utxos[j:]
	err = putUtxos(native, chainID, utxos)
	if err != nil {
		return nil, sum, fmt.Errorf("chooseUtxos, putUtxos err:%v", err)
	}
	return result, sum, nil
}

func putUtxos(native *native.NativeService, chainID uint64, utxos *Utxos) error {
	chainIDBytes := utils.GetUint64Bytes(chainID)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(UTXOS), chainIDBytes)
	sink := common.NewZeroCopySink(nil)
	utxos.Serialization(sink)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getUtxos(native *native.NativeService, chainID uint64) (*Utxos, error) {
	chainIDBytes := utils.GetUint64Bytes(chainID)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(UTXOS), chainIDBytes)
	utxosStore, err := native.GetCacheDB().Get(key)
	if err != nil {
		return nil, fmt.Errorf("getUtxos, get btcTxStore error: %v", err)
	}
	utxos := &Utxos{
		Utxos: make([]*Utxo, 0),
	}
	if utxosStore != nil {
		utxosBytes, err := cstates.GetValueFromRawStorageItem(utxosStore)
		if err != nil {
			return nil, fmt.Errorf("getUtxos, deserialize from raw storage item err:%v", err)
		}
		err = utxos.Deserialization(common.NewZeroCopySource(utxosBytes))
		if err != nil {
			return nil, fmt.Errorf("getUtxos, utxos.Deserialization err:%v", err)
		}
	}
	return utxos, nil
}

func putBtcRedeemScript(native *native.NativeService, redeemScript string) error {
	chainIDBytes := utils.GetUint64Bytes(0)
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(REDEEM_SCRIPT), chainIDBytes)
	redeem, err := hex.DecodeString(redeemScript)
	if err != nil {
		return fmt.Errorf("putBtcRedeemScript, failed to decode redeem script: %v", err)
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

func verifySigs(sigs [][]byte, addr string, addrs []btcutil.Address, redeem []byte, tx *wire.MsgTx) error {
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

		hash, err := txscript.CalcSignatureHash(redeem, txscript.SigHashType(sig[len(sig)-1]), tx, i)
		if err != nil {
			return fmt.Errorf("failed to calculate sig hash: %v", err)
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
	multiSignInfoBytes, err := cstates.GetValueFromRawStorageItem(multiSignInfoStore)
	if err != nil {
		return nil, fmt.Errorf("getBtcMultiSignInfo, deserialize from raw storage item err:%v", err)
	}
	multiSignInfo := &MultiSignInfo{
		MultiSignInfo: make(map[uint64]*MultiSignItem),
	}
	if multiSignInfoStore != nil {
		err = multiSignInfo.Deserialization(common.NewZeroCopySource(multiSignInfoBytes))
		if err != nil {
			return nil, fmt.Errorf("getBtcMultiSignInfo, deserialize multiSignInfo err:%v", err)
		}
	}
	return multiSignInfo, nil
}

func addSigToTx(sigMap map[uint64]*MultiSignItem, addrs []btcutil.Address, redeem []byte, tx *wire.MsgTx) (*wire.MsgTx, error) {
	for idx, item := range sigMap {
		builder := txscript.NewScriptBuilder().AddOp(txscript.OP_FALSE)
		for _, addr := range addrs {
			val, ok := item.MultiSignItem[addr.EncodeAddress()]
			if !ok {
				continue
			}
			builder.AddData(val)
		}

		builder.AddData(redeem)
		script, err := builder.Script()
		if err != nil {
			return nil, fmt.Errorf("failed to build sigscript for input %d: %v", idx, err)
		}

		tx.TxIn[idx].SignatureScript = script
	}

	return tx, nil
}
