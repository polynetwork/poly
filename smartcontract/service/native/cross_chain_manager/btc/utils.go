package btc

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/wire"
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
	if p.Value < p.Fee {
		return errors.New("The transfer amount cannot be less than the transaction fee")
	}
	return nil
}
