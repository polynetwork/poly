package btc

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	// TODO: Temporary setting
	OP_RETURN_DATA_LEN    = 42
	OP_RETURN_SCRIPT_FLAG = byte(0x66)
)

type queryHeaderByHeightParam struct {
	Height uint32 `json:"height"`
}

type QueryHeaderByHeightResp struct {
	Header string `json:"header"`
}

type Response struct {
	Action string                  `json:"action"`
	Desc   string                  `json:"desc"`
	Error  uint32                  `json:"error"`
	Result QueryHeaderByHeightResp `json:"result"`
}

type RestClient struct {
	Addr       string
	restClient *http.Client
}

func NewRestClient() *RestClient {
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

func (self *RestClient) GetHeaderFromSpv(height uint32) (*wire.BlockHeader, error) {
	query, err := json.Marshal(queryHeaderByHeightParam{
		Height: height,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to parse query parameter: %v", err)
	}

	// how to config it???
	data, err := self.SendRestRequest("http://0.0.0.0:20335/api/v1/queryheaderbyheight", query)

	var resp Response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal resp to json: %v", err)
	}

	if resp.Error != 0 || resp.Desc != "SUCCESS" {
		return nil, errors.New("Response shows failure")
	}

	hbs, err := hex.DecodeString(resp.Result.Header)
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
		addrPk, err := btcutil.NewAddressPubKey(v, &chaincfg.TestNet3Params)
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

func checkTxOutputs(tx *wire.MsgTx, pubKeys [][]byte, require int) (ret bool, err error) {
	// has to be 2?
	if len(tx.TxOut) >= 2 {
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
		addr, err := btcutil.NewAddressScriptHash(redeem, &chaincfg.TestNet3Params)
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
func getUnsignedTx(txIns []btcjson.TransactionInput, amounts map[string]float64, locktime *int64) (*wire.MsgTx, error) {
	if locktime != nil &&
		(*locktime < 0 || *locktime > int64(wire.MaxTxInSequenceNum)) {
		return nil, fmt.Errorf("getRawTx, locktime %d out of range", *locktime)
	}

	// Add all transaction inputs to a new transaction after performing
	// some validity checks.
	mtx := wire.NewMsgTx(wire.TxVersion)
	for _, input := range txIns {
		txHash, err := chainhash.NewHashFromStr(input.Txid)
		if err != nil {
			return nil, fmt.Errorf("getRawTx, decode txid fail: %v", err)
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
	params := &chaincfg.TestNet3Params
	for encodedAddr, amount := range amounts {
		// Ensure amount is in the valid range for monetary amounts.
		if amount <= 0 || amount > btcutil.MaxSatoshi {
			return nil, fmt.Errorf("getRawTx, wrong amount: %f", amount)
		}

		// Decode the provided address.
		addr, err := btcutil.DecodeAddress(encodedAddr, params)
		if err != nil {
			return nil, fmt.Errorf("getRawTx, decode addr fail: %v", err)
		}

		// Ensure the address is one of the supported types and that
		// the network encoded with the address matches the network the
		// server is currently on.
		switch addr.(type) {
		case *btcutil.AddressPubKeyHash:
		default:
			return nil, fmt.Errorf("getRawTx, type of addr is not found")
		}
		if !addr.IsForNet(params) {
			return nil, fmt.Errorf("getRawTx, addr is not for mainnet")
		}

		// Create a new script which pays to the provided address.
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, fmt.Errorf("getRawTx, failed to generate pay-to-address script: %v", err)
		}

		// Convert the amount to satoshi.
		satoshi, err := btcutil.NewAmount(amount)
		if err != nil {
			return nil, fmt.Errorf("getRawTx, failed to convert amount: %v", err)
		}

		txOut := wire.NewTxOut(int64(satoshi), pkScript)
		mtx.AddTxOut(txOut)
	}

	// Set the Locktime, if given.
	if locktime != nil {
		mtx.LockTime = uint32(*locktime)
	}

	return mtx, nil
}
