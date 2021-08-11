/*
 * Copyright (C) 2021 The poly network Authors
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
package eth

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/polynetwork/poly/common"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

type proofReq struct {
	JsonRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      uint          `json:"id"`
}

type ProofResult struct {
	JsonRPC string   `json:"jsonrpc"`
	Result  ETHProof `json:"result"`
	Id      uint     `json:"id"`
}

type SProof struct {
	Key   string   `json:"key"`
	Proof []string `json:"proof"`
	Value string   `json:value`
}

func TestProof(t *testing.T) {
	//address := `0xfa98bb293724fa6b012da0f39d4e185f0fe4a749`
	url := "http://42.159.153.121:10331"
	contractAddress := `0x0ec1eeef149b277100b287e6d9991472c191d369`
	blockheight := "6544946 "

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   5,
			DisableKeepAlives:     false,
			IdleConnTimeout:       time.Second * 300,
			ResponseHeaderTimeout: time.Second * 300,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Second * 300,
	}

	val, err := GetProof(url, contractAddress, "", blockheight, client)
	assert.Nil(t, err)
	fmt.Println("value:", val)
}

func GetProof(url string, contractAddress string, key string, blockheight string, restClient *http.Client) (string, error) {
	req := &proofReq{
		JsonRPC: "2.0",
		Method:  "eth_getProof",
		Params:  []interface{}{contractAddress, []string{key}, blockheight},
		Id:      1,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("get_ethproof: marshal req err: %s", err)
	}

	fmt.Printf("req is %s\n", data)
	resp, err := SendRestRequest(restClient, url, data)
	if err != nil {
		return "", fmt.Errorf("GetProof: send request err: %s", err)
	}
	proofRes := &ProofResult{}
	err = json.Unmarshal(resp, proofRes)
	if err != nil {
		return "", fmt.Errorf("GetProof, unmarshal resp err: %s", err)
	}

	fmt.Printf("proof res is:%v\n", proofRes)

	result, err := json.Marshal(proofRes.Result)
	if err != nil {
		return "", fmt.Errorf("GetProof, Marshal result err: %s", err)
	}
	return common.ToHexString([]byte(result)), nil
}

func SendRestRequest(restClient *http.Client, addr string, data []byte) ([]byte, error) {
	resp, err := restClient.Post(addr, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("http post request:%s error:%s", data, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rest response body error:%s", err)
	}
	return body, nil
}
