package eth

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethComm "github.com/ethereum/go-ethereum/common"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/eth/locker"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/smartcontract/service/native/side_chain_manager"
	"strings"
)

type ETHHandler struct {
}

func NewETHHandler() *ETHHandler {
	return &ETHHandler{}
}

func (this *ETHHandler) Verify(service *native.NativeService) (*inf.MakeTxParam, error) {
	//todo add logic
	params := new(inf.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.Input)); err != nil {
		return nil, fmt.Errorf("Verify, contract params deserialize error: %v", err)
	}

	proof := &Proof{}
	if err := proof.Deserialize(params.Proof); err != nil {
		return nil, fmt.Errorf("Verify, eth proof deserialize error: %v", err)
	}

	blockdata,err := GetEthBlockByNumber(params.Height)
	if err != nil{
		return nil,fmt.Errorf("Verify, GetEthBlockByNumber error:%v",err)
	}
	//todo 1. verify the proof with header
	proofresult ,err :=verifyMerkleProof(params.Proof,blockdata)
	if err != nil{
		return nil,fmt.Errorf("Verify, verifyMerkleProof error:%v",err)
	}
	if !proofresult{
		return nil,fmt.Errorf("Verify, verifyMerkleProof failed!")
	}
	ret := &inf.MakeTxParam{}
	ret.ToChainID = proof.ToChainID
	ret.FromChainID = params.SourceChainID
	ret.Address = proof.ToAddress
	//todo 2. transform the decimal if needed
	ret.Amount = proof.Amount

	return ret, nil
}

func (this *ETHHandler) MakeTransaction(service *native.NativeService, param *inf.MakeTxParam) error {
	//todo add logic
	//1 construct tx
	contractabi, err := abi.JSON(strings.NewReader(locker.LockerABI))
	if err != nil {
		return err
	}

	bindaddr := ethComm.HexToAddress(param.Address)
	amount := param.Amount
	//lockAddress := ethComm.HexToAddress(LOCKER_CONTRACT_ADDR)

	targetTokenAddr, err := side_chain_manager.GetAssetContractAddress(service, param.FromChainID, param.ToChainID, param.FromContractAddress)
	if err != nil {
		return err
	}

	tokenAddress := ethComm.HexToAddress(targetTokenAddr)
	txData, err := contractabi.Pack("SendToken", tokenAddress, bindaddr, amount)
	if err != nil {
		return err
	}

	//todo store the txData in storage
	//determin the key format
	service.CacheDB.Put([]byte("TEST_KEY"), txData)

	return nil
}

func verifyMerkleProof(proof string, blockdata *EthBlock)(bool, error){
	//todo add verify logic here
	return true,nil
}