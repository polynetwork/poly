package btc

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
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
	"io/ioutil"
	"net/http"
	"time"
)

const (
	// TODO: Temporary setting
	OP_RETURN_DATA_LEN           = 42
	OP_RETURN_SCRIPT_FLAG        = byte(0x66)
	FEE                          = int64(1e3)
	REQUIRE                      = 5
	CONFRIMATION                 = 6
	BTC_TX_PREFIX         string = "btctx"
	VERIFIED_TX           string = "verified"
	IP                    string = "192.168.3.122:50071" // "0.0.0.0:50071" //"172.168.3.73:50071"
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

// request
type QueryHeaderByHeightParam struct {
	Height uint32 `json:"height"`
}

type QueryUtxosReq struct {
	Addr      string `json:"addr"`
	Amount    int64  `json:"amount"`
	Fee       int64  `json:"fee"`
	IsPreExec bool   `json:"is_pre_exec"`
}

type ChangeAddressReq struct {
	Aciton string `json:"aciton"`
	Addr   string `json:"addr"`
}

type BroadcastTxReq struct {
	RawTx string `json:"raw_tx"`
}

type UnlockUtxoReq struct {
	Hash  string `json:"hash"`
	Index uint32 `json:"index"`
}

type GetFeePerByteReq struct {
	Level int `json:"level"`
}

type RollbackReq struct {
	Time string `json:"time"`
}

type UtxoInfo struct {
	Outpoint string `json:"outpoint"`
	Val      int64  `json:"val"`
	IsLock   bool   `json:"is_lock"`
	Height   int32  `json:"height"`
	Script   string `json:"script"`
}

type GetAllUtxosResp struct {
	Infos []UtxoInfo `json:"infos"`
}

// response
type Response struct {
	Action string      `json:"action"`
	Desc   string      `json:"desc"`
	Error  uint32      `json:"error"`
	Result interface{} `json:"result"`
}

type ResponseAllUtxos struct {
	Action string          `json:"action"`
	Desc   string          `json:"desc"`
	Error  uint32          `json:"error"`
	Result GetAllUtxosResp `json:"result"`
}

type RestClient struct {
	Addr       string
	restClient *http.Client
}

func NewRestClient(addr string) *RestClient {
	return &RestClient{
		restClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   5,
				DisableKeepAlives:     false,
				IdleConnTimeout:       time.Second * 300,
				ResponseHeaderTimeout: time.Second * 300,
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: time.Second * 300,
		},
		Addr: addr,
	}
}

func (self *RestClient) SetAddr(addr string) *RestClient {
	self.Addr = addr
	return self
}

func (self *RestClient) SetRestClient(restClient *http.Client) *RestClient {
	self.restClient = restClient
	return self
}

func (self *RestClient) SendRestRequest(addr string, data []byte) ([]byte, error) {
	resp, err := self.restClient.Post(addr, "application/json;charset=UTF-8",
		bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("rest post request:%s error:%s", data, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rest response body error:%s", err)
	}
	return body, nil
}

func (self *RestClient) SendGetRequst(addr string) ([]byte, error) {
	resp, err := self.restClient.Get(addr)
	if err != nil {
		return nil, fmt.Errorf("rest get request: error: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read get response body error:%s", err)
	}
	return body, nil
}

func (self *RestClient) GetHeaderFromSpv(height uint32) (*wire.BlockHeader, error) {
	query, err := json.Marshal(QueryHeaderByHeightParam{
		Height: height,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to parse query parameter: %v", err)
	}

	// how to config it???
	data, err := self.SendRestRequest("http://"+self.Addr+"/api/v1/queryheaderbyheight", query)
	if err != nil {
		return nil, fmt.Errorf("Failed to send request: %v", err)
	}
	var resp Response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal resp to json: %v", err)
	}

	if resp.Error != 0 || resp.Desc != "SUCCESS" {
		return nil, fmt.Errorf("Response shows failure: %s", resp.Desc)
	}

	hbs, err := hex.DecodeString(resp.Result.(map[string]interface{})["header"].(string))
	if err != nil {
		return nil, fmt.Errorf("Failed to decode hex string from response: %v", err)
	}

	header := wire.BlockHeader{}
	buf := bytes.NewReader(hbs)
	err = header.BtcDecode(buf, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode header: %v", err)
	}
	return &header, nil
}

func (self *RestClient) GetUtxosFromSpv(addr string, amount int64, fee int64, isPreExec bool) ([]btcjson.TransactionInput, int64, error) {
	query, err := json.Marshal(QueryUtxosReq{
		Addr:      addr,
		Amount:    amount,
		Fee:       fee,
		IsPreExec: isPreExec,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to parse parameter: %v", err)
	}
	data, err := self.SendRestRequest("http://"+self.Addr+"/api/v1/queryutxos", query)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to send request: %v", err)
	}

	var resp Response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to unmarshal resp to json: %v", err)
	}
	if resp.Error != 0 || resp.Desc != "SUCCESS" {
		return nil, 0, fmt.Errorf("Response shows failure: %s", resp.Desc)
	}
	var ins []btcjson.TransactionInput
	for _, v := range resp.Result.(map[string]interface{})["inputs"].([]interface{}) {
		m := v.(map[string]interface{})
		ins = append(ins, btcjson.TransactionInput{
			Txid: m["txid"].(string),
			Vout: uint32(m["vout"].(float64)),
		})
	}

	return ins, int64(resp.Result.(map[string]interface{})["sum"].(float64)), nil
}

func (self *RestClient) GetCurrentHeightFromSpv() (uint32, error) {
	data, err := self.SendGetRequst("http://" + self.Addr + "/api/v1/getcurrentheight")
	if err != nil {
		return 0, fmt.Errorf("Failed to send request: %v", err)
	}

	var resp Response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return 0, fmt.Errorf("Failed to unmarshal resp to json: %v", err)
	}
	if resp.Error != 0 || resp.Desc != "SUCCESS" {
		return 0, fmt.Errorf("Response shows failure: %s", resp.Desc)
	}

	return uint32(resp.Result.(map[string]interface{})["height"].(float64)), nil
}

func (self *RestClient) ChangeSpvWatchedAddr(addr string, action string) error {
	req, err := json.Marshal(ChangeAddressReq{
		Addr:   addr,
		Aciton: action,
	})
	if err != nil {
		return fmt.Errorf("Failed to parse parameter: %v", err)
	}
	data, err := self.SendRestRequest("http://"+self.Addr+"/api/v1/changeaddress", req)
	if err != nil {
		return fmt.Errorf("Failed to send request: %v", err)
	}

	var resp Response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal resp to json: %v", err)
	}
	if resp.Error != 0 || resp.Desc != "SUCCESS" {
		return fmt.Errorf("Response shows failure: %s", resp.Desc)
	}

	return nil
}

func (self *RestClient) GetWatchedAddrsFromSpv() ([]string, error) {
	data, err := self.SendGetRequst("http://" + self.Addr + "/api/v1/getalladdress")
	if err != nil {
		return nil, fmt.Errorf("Failed to send request: %v", err)
	}

	var resp Response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal resp to json: %v", err)
	}
	if resp.Error != 0 || resp.Desc != "SUCCESS" {
		return nil, fmt.Errorf("Response shows failure: %s", resp.Desc)
	}
	var addrs []string
	for _, v := range resp.Result.(map[string]interface{})["addresses"].([]interface{}) {
		addrs = append(addrs, v.(string))
	}
	return addrs, nil
}

func (self *RestClient) UnlockUtxoInSpv(hash string, index uint32) error {
	req, err := json.Marshal(UnlockUtxoReq{
		Hash:  hash,
		Index: index,
	})
	if err != nil {
		return fmt.Errorf("Failed to parse parameter: %v", err)
	}
	data, err := self.SendRestRequest("http://"+self.Addr+"/api/v1/unlockutxo", req)
	if err != nil {
		return fmt.Errorf("Failed to send request: %v", err)
	}

	var resp Response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal resp to json: %v", err)
	}
	if resp.Error != 0 || resp.Desc != "SUCCESS" {
		return fmt.Errorf("Response shows failure: %s", resp.Desc)
	}

	return nil
}

func (self *RestClient) GetFeeRateFromSpv(level int) (int64, error) {
	req, err := json.Marshal(GetFeePerByteReq{
		Level: level,
	})
	if err != nil {
		return -1, fmt.Errorf("Failed to parse parameter: %v", err)
	}

	data, err := self.SendRestRequest("http://"+self.Addr+"/api/v1/getfeeperbyte", req)
	if err != nil {
		return -1, fmt.Errorf("Failed to send request: %v", err)
	}

	var resp Response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return -1, fmt.Errorf("Failed to unmarshal resp to json: %v", err)
	}
	if resp.Error != 0 || resp.Desc != "SUCCESS" {
		return -1, fmt.Errorf("Response shows failure: %s", resp.Desc)
	}

	return int64(resp.Result.(map[string]interface{})["feepb"].(float64)), nil
}

func (self *RestClient) GetAllUtxosFromSpv() ([]UtxoInfo, error) {
	data, err := self.SendGetRequst("http://" + self.Addr + "/api/v1/getallutxos")
	if err != nil {
		return nil, fmt.Errorf("Failed to send request: %v", err)
	}

	var resp ResponseAllUtxos
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal resp to json: %v", err)
	}
	if resp.Error != 0 || resp.Desc != "SUCCESS" {
		return nil, fmt.Errorf("Response shows failure: %s", resp.Desc)
	}

	return resp.Result.Infos, nil
}

func (self *RestClient) BroadcastTxBySpv(mtx *wire.MsgTx) error {
	var buf bytes.Buffer
	err := mtx.BtcEncode(&buf, wire.ProtocolVersion, wire.LatestEncoding)
	if err != nil {
		return err
	}
	req, err := json.Marshal(BroadcastTxReq{
		RawTx: hex.EncodeToString(buf.Bytes()),
	})
	if err != nil {
		return fmt.Errorf("Failed to parse parameter: %v", err)
	}

	data, err := self.SendRestRequest("http://"+self.Addr+"/api/v1/broadcasttx", req)
	if err != nil {
		return fmt.Errorf("Failed to send request: %v", err)
	}

	var resp Response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal resp to json: %v", err)
	}
	if resp.Error != 0 || resp.Desc != "SUCCESS" {
		return fmt.Errorf("Response shows failure: %s", resp.Desc)
	}

	return nil
}

func (self *RestClient) RollbackSpv(time string) error {
	req, err := json.Marshal(RollbackReq{
		Time: time,
	})
	if err != nil {
		return fmt.Errorf("Failed to parse parameter: %v", err)
	}
	data, err := self.SendRestRequest("http://"+self.Addr+"/api/v1/rollback", req)
	if err != nil {
		return fmt.Errorf("Failed to send request: %v", err)
	}

	var resp Response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal resp to json: %v", err)
	}
	if resp.Error != 0 || resp.Desc != "SUCCESS" {
		return fmt.Errorf("Response shows failure: %s", resp.Desc)
	}

	return nil
}

// not sure now
type targetChainParam struct {
	ChainId uint64
	Fee     int64
	Addr    []byte // 25 bytes
	Value   int64
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
	p.Addr = script[19:]
	p.Value = amount
	if p.Value < p.Fee && p.Fee >= 0 {
		return errors.New("The transfer amount cannot be less than the transaction fee")
	}
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

//func getTxOuts(amounts map[string]int64) ([]*wire.TxOut, error) {
//	outs := make([]*wire.TxOut, 0)
//	for encodedAddr, amount := range amounts {
//		// Decode the provided address.
//		addr, err := btcutil.DecodeAddress(encodedAddr, netParam)
//		if err != nil {
//			return nil, fmt.Errorf("getTxOuts, decode addr fail: %v", err)
//		}
//
//		// Ensure the address is one of the supported types and that
//		// the network encoded with the address matches the network the
//		// server is currently on.
//		switch addr.(type) {
//		case *btcutil.AddressPubKeyHash:
//		case *btcutil.AddressScriptHash:
//		default:
//			return nil, fmt.Errorf("getTxOuts, type of addr is not found")
//		}
//		if !addr.IsForNet(netParam) {
//			return nil, fmt.Errorf("getTxOuts, addr is not for mainnet")
//		}
//
//		// Create a new script which pays to the provided address.
//		pkScript, err := txscript.PayToAddrScript(addr)
//		if err != nil {
//			return nil, fmt.Errorf("getTxOuts, failed to generate pay-to-address script: %v", err)
//		}
//
//		txOut := wire.NewTxOut(amount, pkScript)
//		outs = append(outs, txOut)
//	}
//
//	return outs, nil
//}
