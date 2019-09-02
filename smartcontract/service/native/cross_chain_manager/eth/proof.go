package eth

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	ethComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Proof struct {
	AssetAddress string
	FromAddress  string
	ToChainID    uint64
	ToAddress    string
	Amount       *big.Int
	Decimal      int
}

type StorageProof struct {
	Key   string   `json:"key"`
	Value string   `json:"value"`
	Proof []string `json:"proof"`
}

type ETHProof struct {
	Address       string         `json:"address"`
	Balance       string         `json:"balance"`
	CodeHash      string         `json:"codeHash"`
	Nonce         string         `json:"nonce"`
	StorageHash   string         `json:"storageHash"`
	AccountProof  []string       `json:"accountProof"`
	StorageProofs []StorageProof `json:"storageProof"`
}

func (this *ETHProof) String() string {
	bs := bytes.NewBuffer([]byte("ETHProof:\n"))
	bs.WriteString("AccountProof:\n")
	for _, a := range this.AccountProof {
		bs.WriteString(a + "\n")
	}
	bs.WriteString("Address:")
	bs.WriteString(this.Address + "\n")
	bs.WriteString("StorageProof:\n")
	for _, s := range this.StorageProofs {
		bs.WriteString(s.Key + "\n")
		bs.WriteString("proofs:\n[")
		bs.WriteString(strings.Join(s.Proof, "\n"))
		bs.WriteString("]\n")

		bs.WriteString(s.Value + "\n")
	}
	return bs.String()
}

type RpcResponse struct {
	JsonRpc string `json:"jsonrpc"`
	Result  []byte `json:"result"`
	id      int    `json:"id"`
}

func GetProof() ([]byte, error) {

	params := []interface{}{"0xfa98bb293724fa6b012da0f39d4e185f0fe4a749", []string{"0x2a1543b4300f0f31df4d4ca5a28e30970d5e92ab3c4b01b8df45979ff2a863f5"}, "latest"}

	req := &Request{
		JsonRpc: "2.0",
		Id:      1,
		Method:  "eth_getProof",
		Params:  params,
	}

	reqbs, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://127.0.0.1:8545", "application/json", strings.NewReader(string(reqbs)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := &RpcResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, err
	}
	return response.Result, nil
}

func MappingKeyAt(position1 string, position2 string) ([]byte, error) {

	p1, err := hex.DecodeString(position1)
	if err != nil {
		return nil, err
	}

	p2, err := hex.DecodeString(position2)

	if err != nil {
		return nil, err
	}

	key := crypto.Keccak256(ethComm.LeftPadBytes(p1, 32), ethComm.LeftPadBytes(p2, 32))

	return key, nil
}

func (this *Proof) Deserialize(raw string) error {
	vals := strings.Split(raw, "#")
	if len(vals) != 6 {
		return fmt.Errorf("error count of proof deserialize")
	}
	this.AssetAddress = vals[0]
	this.FromAddress = vals[1]
	cid, err := strconv.Atoi(vals[2])
	if err != nil {
		return fmt.Errorf("chain id is not correct")
	}
	this.ToChainID = uint64(cid)
	this.ToAddress = vals[3]
	amt := new(big.Int)
	_, b := amt.SetString(vals[4], 10)
	if !b {
		return fmt.Errorf("amount is not correct")
	}
	this.Amount = amt
	decimal, err := strconv.Atoi(vals[5])
	if err != nil {
		return fmt.Errorf("decimal is not correct")
	}
	this.Decimal = decimal

	return nil
}
