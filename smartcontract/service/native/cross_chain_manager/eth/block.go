package eth

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
)

type EthBlock struct {
	Author          string           `json:"author"`
	Difficulty      string           `json:"difficulty"`
	ExtraData       string           `json:"extraData"`
	Gaslimit        string           `json:"gaslimit"`
	GasUsed         string           `json:"gasUsed"`
	Hash            string           `json:"hash"`
	LogBloom        string           `json:"logBloom"`
	Miner           string           `json:"miner"`
	MixHash         string           `json:"mixHash"`
	Nounce          string           `json:"nounce"`
	ParentHash      string           `json:"parentHash"`
	ReceiptRoot     string           `json:"receiptRoot"`
	SealFields      []string         `json:"sealFields"`
	Sha3Uncles      string           `json:"sha3Uncles"`
	Size            string           `json:"size"`
	StateRoot       string           `json:"stateRoot"`
	TimeStamp       string           `json:"timeStamp"`
	TotalDifficulty string           `json:"totalDifficulty"`
	Transactions    []ETHTransaction `json:"transactions"`
	TransactionRoot string           `json:"transactionRoot"`
	Uncles          []interface{}    `json:"uncles"`
}

type ETHTransaction struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	ChainID          string `json:"chainId"`
	Condition        string `json:"condition"`
	Creates          string `json:"creates"`
	From             string `json:"from"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nounce           string `json:"nounce"`
	PublicKey        string `json:"publicKey"`
	R                string `json:"r"`
	Raw              string `json:"raw"`
	S                string `json:"s"`
	StandardV        string `json:"standardV"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	V                string `json:"v"`
	Value            string `json:"value"`
}

type Response struct {
	JsonRpc string    `json:"jsonrpc"`
	Result  *EthBlock `json:"result"`
	id      int       `json:"id"`
}

type Request struct {
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
	JsonRpc string        `json:"jsonrpc"`
}

type ProofAccount struct {
	Nounce   *big.Int
	Balance  *big.Int
	Storage  common.Hash
	Codehash common.Hash
}

func GetEthBlockByNumber(num uint32) (*EthBlock, error) {

	hexnum := hexutil.EncodeUint64(uint64(num))
	fmt.Printf("hexnum:%s\n", hexnum)
	req := &Request{
		JsonRpc: "2.0",
		Id:      1,
		Method:  "eth_getBlockByNumber",
		Params:  []interface{}{hexnum, true},
	}

	reqbs, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	fmt.Printf("req is %s\n", reqbs)

	resp, err := http.Post("http://127.0.0.1:8545", "application/json", strings.NewReader(string(reqbs)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("body is %s\n", body)
	response := &Response{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, err
	}
	return response.Result, nil
}
