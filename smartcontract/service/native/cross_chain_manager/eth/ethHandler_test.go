package eth

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethComm "github.com/ethereum/go-ethereum/common"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/eth/locker"
	"strings"
	"testing"

	"math/big"
)

func TestETHHandler_MakeTransaction(t *testing.T) {
	contractabi, err := abi.JSON(strings.NewReader(locker.EthereumCrossChainABI))
	if err != nil {
		return
	}

	tokenAddress := ethComm.HexToAddress("0x0000000000000000000000000000000000000000")
	txid := "1"
	bindaddr := ethComm.HexToAddress("0xfA98bb293724fA6b012DA0F39D4e185f0fE4A749")
	amount := big.NewInt(100)

	v := []uint8{0}
	r := [][32]byte{[32]byte{0}}
	s := [][32]byte{[32]byte{0}}

	txData, err := contractabi.Pack("Withdraw", tokenAddress, txid, bindaddr, amount, v, r, s)
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
		return
	}
	fmt.Printf("%v\n", txData)
}
