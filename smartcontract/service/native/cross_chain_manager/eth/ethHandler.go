package eth

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethComm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/eth/locker"
	"github.com/ontio/multi-chain/smartcontract/service/native/cross_chain_manager/inf"
	"github.com/ontio/multi-chain/smartcontract/service/native/side_chain_manager"
	"strings"
	"encoding/json"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"github.com/ethereum/go-ethereum/rlp"
	"bytes"
)

type ETHHandler struct {
}

func NewETHHandler() *ETHHandler {
	return &ETHHandler{}
}

func (this *ETHHandler) Verify(service *native.NativeService) (*inf.MakeTxParam, error) {
	params := new(inf.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.Input)); err != nil {
		return nil, fmt.Errorf("Verify, contract params deserialize error: %v", err)
	}

	blockdata, err := GetEthBlockByNumber(params.Height)
	if err != nil {
		return nil, fmt.Errorf("Verify, GetEthBlockByNumber error:%v", err)
	}

	proofdata, err := hex.DecodeString(params.Proof)
	if err != nil {
		return nil, err
	}

	ethproof := new(ETHProof)
	err = json.Unmarshal(proofdata,ethproof)
	if err != nil {
		return nil, err
	}

	if len(ethproof.StorageProofs) != 1 {
		return nil, fmt.Errorf("[Verify] incorrect proof format")
	}
	keybytes := ethComm.Hex2Bytes( inf.Key_prefix_ETH + replace0x(ethproof.StorageProofs[0].Key))
	val ,err := service.CacheDB.Get(keybytes)
	if err != nil{
		return nil,err
	}
	if val != nil{
		return nil,fmt.Errorf("[Verify] key:%s already solved ",ethproof.StorageProofs[0].Key)
	}

	//todo 1. verify the proof with header
	//determine where the k and v from
	proofresult, err := verifyMerkleProof(ethproof, blockdata)
	if err != nil {
		return nil, fmt.Errorf("Verify, verifyMerkleProof error:%v", err)
	}
	if proofresult == nil {
		return nil, fmt.Errorf("Verify, verifyMerkleProof failed!")
	}

	proof := &Proof{}
	if err := proof.Deserialize(proofresult); err != nil {
		return nil, fmt.Errorf("Verify, eth proof deserialize error: %v", err)
	}

	//todo does the proof data too big??
	service.CacheDB.Put(keybytes,proofdata)

	ret := &inf.MakeTxParam{}
	ret.ToChainID = proof.ToChainID
	ret.FromChainID = params.SourceChainID
	ret.ToAddress = proof.ToAddress
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

	bindaddr := ethComm.HexToAddress(param.ToAddress)
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

func verifyMerkleProof(ethproof *ETHProof, blockdata *EthBlock) ([]byte, error) {

	//1. prepare verify account
	nodelist := new(light.NodeList)

	for _,s:= range ethproof.AccountProof{
		p :=  replace0x(s)
		nodelist.Put(nil,ethComm.Hex2Bytes(p))
	}
	ns := nodelist.NodeSet()

	acctkey := crypto.Keccak256(ethComm.Hex2Bytes(replace0x(ethproof.Address)))

	//2. verify account proof
	acctval,_,err:=trie.VerifyProof(ethComm.HexToHash(replace0x(blockdata.StateRoot)),acctkey,ns)
	if err != nil{
		return nil, err
	}

	nounce := new(big.Int)
	err = rlp.DecodeBytes(ethComm.Hex2Bytes( replace0x(ethproof.Nonce)),nounce)
	if err != nil{
		return nil,err
	}
	balance := new(big.Int)
	err = rlp.DecodeBytes(ethComm.Hex2Bytes( replace0x(ethproof.Balance)),balance)
	if err != nil{
		return nil,err
	}

	storagehash := ethComm.HexToHash(replace0x(ethproof.StorageHash))
	codehash := ethComm.HexToHash(replace0x(ethproof.CodeHash))
	//construct the account value
	acct := &ProofAccount{
		Nounce:nounce,
		Balance:balance,
		Storage:storagehash,
		Codehash:codehash,
	}
	acctrlp ,err:= rlp.EncodeToBytes(acct)
	if err != nil{
		return nil,err
	}
	if !bytes.Equal(acctrlp,acctval){
		return nil, fmt.Errorf("[verifyMerkleProof]: verify account proof failed, wanted:%v, get:%v",acctrlp,acctval)
	}

	//3.verify storage proof
	nodelist = new(light.NodeList)

	if len(ethproof.StorageProofs) != 1 {
		return nil,fmt.Errorf("[verifyMerkleProof]: storage proof fmt error")
	}

	sp := ethproof.StorageProofs[0]
	storagekey := crypto.Keccak256(ethComm.Hex2Bytes(replace0x(sp.Key)))
	for _, prf := range sp.Proof{
		nodelist.Put(nil,ethComm.Hex2Bytes(replace0x(prf)))
	}

	ns = nodelist.NodeSet()
	val, _, err := trie.VerifyProof(storagehash, storagekey, ns)
	if err != nil{
		return nil,err
	}

	return val, nil
}

func replace0x(s string) string {
	p :=  strings.Replace(s,"0x","",0)
	return strings.Replace(p,"0X","",0)
}
