package btc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
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
var addr1 = "mj3LUsSvk9ZQH1pSHvC8LBtsYXsZvbky8H"
var priv1 = "cTqbqa1YqCf4BaQTwYDGsPAB4VmWKUU67G5S1EtrHSWNRwY6QSag"
var addr2 = "mtNiC48WWbGRk2zLqiTMwKLhrCk6rBqBen"
var priv2 = "cT2HP4QvL8c6otn4LrzUWzgMBfTo1gzV2aobN1cTiuHPXH9Jk2ua"
var addr3 = "mi1bYK8SR3Qsf2cdrxgak3spzFx4EVH1pf"
var priv3 = "cSQmGg6spbhd23jHQ9HAtz3XU7GYJjYaBmFLWHbyKa9mWzTxEY5A"
var addr4 = "mz3bTZaQ2tNzsn4szNE8R6gp5zyHuqN29V"
var priv4 = "cPYAx61EjwshK5SQ6fqH7QGjc8L48xiJV7VRGpYzPSbkkZqrzQ5b"
var addr5 = "mfzbFf6njbEuyvZGDiAdfKamxWfAMv47NG"
var priv5 = "cVV9UmtnnhebmSQgHhbDZWCb7zBHbiAGDB9a5M2ffe1WpqvwD5zg"
var addr6 = "n4ESieuFJq5HCvE5GU8B35YTfShZmFrCKM"
var priv6 = "cNK7BwHmi8rZiqD2QfwJB1R6bF6qc7iVTMBNjTr2ACbsoq1vWau8"
var addr7 = "msK9xpuXn5xqr4UK7KyWi9VCaFhiwCqqq6"
var priv7 = "cUZdDF9sL11ya5civzMRYVYojoojjHbmWWm1yC5uRzfBRePVbQTZ"

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
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes([]byte(toAddr.ToBase58()))
	sink.WriteInt64(amount)
	p.AddrAndVal = sink.Bytes()
	return nil
}

func buildScript(pubks [][]byte, require int) ([]byte, error) {
	if len(pubks) == 0 || require <= 0 {
		return nil, errors.New("Wrong public keys or require number")
	}
	var addrPks []*btcutil.AddressPubKey
	for _, v := range pubks {
		addrPk, err := btcutil.NewAddressPubKey(v, netParam)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse address pubkey: %v", err)
		}
		addrPks = append(addrPks, addrPk)
	}
	s, err := txscript.MultiSigScript(addrPks, require)
	if err != nil {
		return nil, fmt.Errorf("Failed to build multi-sig script: %v", err)
	}

	return s, nil
}

func getPubKeys() [][]byte {
	_, pubk1 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv1))
	_, pubk2 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv2))
	_, pubk3 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv3))
	_, pubk4 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv4))
	_, pubk5 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv5))
	_, pubk6 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv6))
	_, pubk7 := btcec.PrivKeyFromBytes(btcec.S256(), base58.Decode(priv7))

	pubks := make([][]byte, 0)
	pubks = append(pubks, pubk1.SerializeCompressed(), pubk2.SerializeCompressed(), pubk3.SerializeCompressed(),
		pubk4.SerializeCompressed(), pubk5.SerializeCompressed(), pubk6.SerializeCompressed(), pubk7.SerializeCompressed())
	return pubks
}

func checkTxOutputs(tx *wire.MsgTx, pubKeys [][]byte, require int) (ret bool, err error) {
	// has to be 2?
	if len(tx.TxOut) < 2 {
		return false, errors.New("Number of transaction's outputs is at least greater than 2")
	}
	if tx.TxOut[0].Value <= 0 {
		return false, fmt.Errorf("The value of crosschain transaction must be bigger than 0, but value is %d",
			tx.TxOut[0].Value)
	}

	redeem, err := buildScript(pubKeys, require)
	if err != nil {
		return false, fmt.Errorf("Failed to build redeem script: %v", err)
	}
	c1 := txscript.GetScriptClass(tx.TxOut[0].PkScript)
	if c1 == txscript.MultiSigTy {
		if !bytes.Equal(redeem, tx.TxOut[0].PkScript) {
			return false, fmt.Errorf("Wrong script: \"%x\" is not same as our \"%x\"",
				tx.TxOut[0].PkScript, redeem)
		}
	} else if c1 == txscript.ScriptHashTy {
		addr, err := btcutil.NewAddressScriptHash(redeem, netParam)
		if err != nil {
			return false, err
		}
		h, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return false, err
		}
		if !bytes.Equal(h, tx.TxOut[0].PkScript) {
			return false, fmt.Errorf("Wrong script: \"%x\" is not same as our \"%x\"", tx.TxOut[0].PkScript, h)
		}
	} else {
		return false, errors.New("First output's pkScript is not supported")
	}

	c2 := txscript.GetScriptClass(tx.TxOut[1].PkScript)
	if c2 != txscript.NullDataTy {
		return false, errors.New("Second output's pkScript is not NullData type")
	}

	return true, nil
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
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("putUtxo, utils.GetBytesUint64 err:%v", err)
	}
	key := utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(UTXOS), chainIDBytes)
	sink := common.NewZeroCopySink(nil)
	utxos.Serialization(sink)
	native.GetCacheDB().Put(key, cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getUtxos(native *native.NativeService, chainID uint64) (*Utxos, error) {
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return nil, fmt.Errorf("getUtxos, utils.GetBytesUint64 err:%v", err)
	}
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
