package eth

import (
	"testing"
	"fmt"
)

func TestGetEthBlockByNumber(t *testing.T) {
	num := uint32(6097203)
	blockData,err := GetEthBlockByNumber(num)
	if err != nil{
		fmt.Printf("err:%v",err)
	}
	fmt.Printf("blockData:%v\n",blockData)
}
