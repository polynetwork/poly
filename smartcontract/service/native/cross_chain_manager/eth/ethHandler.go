package eth

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

type ETHHandler struct {
}

func NewETHHandler() *ETHHandler {
	return &ETHHandler{}
}

func (this *ETHHandler) Verify(service *native.NativeService) (*inf.EntranceParam, error) {
	//todo add logic
	return nil, nil
}

func (this *ETHHandler) MakeTransaction(service *native.NativeService, param *inf.EntranceParam) error {
	//todo add logic
	return nil
}

func EthHandler(native *native.NativeService) ([]byte, error) {

	params := new(inf.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, contract params deserialize error: %v", err)
	}

	txdata := []byte(params.TxData)

	//1. get txhash of txdata

	tx := types.Transaction{}
	err := rlp.DecodeBytes(txdata, tx)
	if err != nil {
		return nil, err
	}
	//todo get from input??
	//ethclientChainid := big.NewInt(0)
	////get from address
	//msg, err := tx.AsMessage(types.NewEIP155Signer(ethclientChainid))
	//if err != nil{
	//	return nil, err
	//}

	//trie.VerifyProof()
	//
	//fromaddress := msg.From()
	//toaddress := msg.To()
	//value := msg.Value()
	//
	//proof := params.Proof
	//height := params.Height
	//2. validate the proof

	//3. parse the tx content

	//4. make the invoke tx bytes of the dest chain

	//5. save the tx in txpool

	return nil, nil
}
