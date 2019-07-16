package cross_chain_manager

import (
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"fmt"
	"github.com/ontio/multi-chain/common"
)
const(
	ImportExTransfer_Name = "ImportOuterTransfer"
	RegisterChainHandler_ID = "RegisterChainHandler"

)

type CrossChainHandler func(native *native.NativeService)([]byte, error)
var mapping = make(map[uint32]CrossChainHandler)



func InitEntrance(){
	native.Contracts[utils.CrossChainManagerContractAddress] =RegisterCrossChianManagerContract
	//RegisterChainHandler(0,BTCHandler)
	//RegisterChainHandler(1,ETHHandler)
}

func RegisterCrossChianManagerContract(native *native.NativeService){
	native.Register(ImportExTransfer_Name, ImportExTransfer)
}

func RegisterChainHandler(chainid uint32,handler CrossChainHandler ){
	mapping[chainid] = handler
}

func GetChainHandler(chainid uint32)(CrossChainHandler,error){
	handler,ok := mapping[chainid]
	if !ok {
		return nil, fmt.Errorf("no handler for chainID:%d",chainid)
	}
	return handler,nil
}


func ImportExTransfer(native *native.NativeService)([]byte ,error){
	params := new(EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, contract params deserialize error: %v", err)
	}

	chainid := params.SourceChainID
	handler,err := GetChainHandler(chainid)
	if err != nil{
		return nil, err
	}
	return handler(native)
}

