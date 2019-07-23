package btc

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"io/ioutil"
	"net/http"
	"time"
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
