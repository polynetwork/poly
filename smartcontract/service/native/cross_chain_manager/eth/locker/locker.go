// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package locker

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)


	
	// EthereumCrossChainABI is the input ABI used to generate the binding from.
	const EthereumCrossChainABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"owners\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"transactionId\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"transactions\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_txId\",\"type\":\"string\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_v\",\"type\":\"uint8[]\"},{\"name\":\"_r\",\"type\":\"bytes32[]\"},{\"name\":\"_s\",\"type\":\"bytes32[]\"}],\"name\":\"Withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"required\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_toChainId\",\"type\":\"string\"},{\"name\":\"_toAddress\",\"type\":\"string\"}],\"name\":\"CrossChain\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_owners\",\"type\":\"address[]\"},{\"name\":\"_required\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"\",\"type\":\"string\"}],\"name\":\"Debugs\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"DebugByte\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"topic\",\"type\":\"string\"},{\"indexed\":true,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"txId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"decimal\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"toChainId\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"toAddress\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"rawdata\",\"type\":\"string\"}],\"name\":\"CrossChainEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"topic\",\"type\":\"string\"},{\"indexed\":true,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"toAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"WithDrawEvent\",\"type\":\"event\"}]"

	
		// EthereumCrossChainFuncSigs maps the 4-byte function signature to its string representation.
		var EthereumCrossChainFuncSigs = map[string]string{
			"fafa0765": "CrossChain(address,uint256,string,string)",
			"b70d4ee9": "Withdraw(address,string,address,uint256,uint8[],bytes32[],bytes32[])",
			"025e7c27": "owners(uint256)",
			"dc8452cd": "required()",
			"7e2f42e7": "transactionId()",
			"9ace38c2": "transactions(uint256)",
			
		}
	

	
		// EthereumCrossChainBin is the compiled bytecode used for deploying new contracts.
		var EthereumCrossChainBin = "0x60806040523480156200001157600080fd5b5060405162001bb138038062001bb183398101604052805160208201519101805190919060009081908311156200004757600080fd5b8351603210156200005757600080fd5b600091505b8351821015620000ef5783828151811015156200007557fe5b602090810290910101519050600160a060020a03811615156200009757600080fd5b600160a060020a03811660009081526003602052604090205460ff1615620000be57600080fd5b600160a060020a0381166000908152600360205260409020805460ff1916600190811790915591909101906200005c565b83516200010490600290602087019062000111565b50505060015550620001a5565b82805482825590600052602060002090810192821562000169579160200282015b82811115620001695782518254600160a060020a031916600160a060020a0390911617825560209092019160019091019062000132565b50620001779291506200017b565b5090565b620001a291905b8082111562000177578054600160a060020a031916815560010162000182565b90565b6119fc80620001b56000396000f30060806040526004361061005e5763ffffffff60e060020a600035041663025e7c2781146100635780637e2f42e7146100975780639ace38c2146100be578063b70d4ee9146100d6578063dc8452cd14610202578063fafa076514610217575b600080fd5b34801561006f57600080fd5b5061007b6004356102b1565b60408051600160a060020a039092168252519081900360200190f35b3480156100a357600080fd5b506100ac6102d9565b60408051918252519081900360200190f35b3480156100ca57600080fd5b506100ac6004356102df565b3480156100e257600080fd5b5060408051602060046024803582810135601f8101859004850286018501909652858552610200958335600160a060020a03169536956044949193909101919081908401838280828437505060408051818801358901803560208181028481018201909552818452989b600160a060020a038b35169b8a8c01359b919a90995060609091019750929550908201935091829185019084908082843750506040805187358901803560208181028481018201909552818452989b9a998901989297509082019550935083925085019084908082843750506040805187358901803560208181028481018201909552818452989b9a9989019892975090820195509350839250850190849080828437509497506102f19650505050505050565b005b34801561020e57600080fd5b506100ac61097d565b604080516020600460443581810135601f8101849004840285018401909552848452610200948235600160a060020a031694602480359536959460649492019190819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a9998810197919650918201945092508291508401838280828437509497506109839650505050505050565b60028054829081106102bf57fe5b600091825260209091200154600160a060020a0316905081565b60045481565b60006020819052908152604090205481565b600080600160a060020a038716151561030957600080fd5b6000861161031657600080fd5b600160a060020a03891615156104e357876103308861107b565b61033988611281565b6040516020018084805190602001908083835b6020831061036b5780518252601f19909201916020918201910161034c565b51815160209384036101000a600019018019909216911617905286519190930192860191508083835b602083106103b35780518252601f199092019160209182019101610394565b51815160209384036101000a600019018019909216911617905285519190930192850191508083835b602083106103fb5780518252601f1990920191602091820191016103dc565b6001836020036101000a03801982511681845116808217855250505050505090500193505050506040516020818303038152906040526040518082805190602001908083835b602083106104605780518252601f199092019160209182019101610441565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040518091039020915061049c8883878787611367565b15156104a757600080fd5b604051600160a060020a0388169087156108fc029088906000818181858888f193505050501580156104dd573d6000803e3d6000fd5b506107cb565b600160a060020a0389161561076457876104fc8a61107b565b6105058961107b565b61050e89611281565b6040516020018085805190602001908083835b602083106105405780518252601f199092019160209182019101610521565b51815160209384036101000a600019018019909216911617905287519190930192870191508083835b602083106105885780518252601f199092019160209182019101610569565b51815160209384036101000a600019018019909216911617905286519190930192860191508083835b602083106105d05780518252601f1990920191602091820191016105b1565b51815160209384036101000a600019018019909216911617905285519190930192850191508083835b602083106106185780518252601f1990920191602091820191016105f9565b6001836020036101000a0380198251168184511680821785525050505050509050019450505050506040516020818303038152906040526040518082805190602001908083835b6020831061067e5780518252601f19909201916020918201910161065f565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051809103902091506106ba8883878787611367565b15156106c557600080fd5b88600160a060020a031663a9059cbb88886040518363ffffffff1660e060020a0281526004018083600160a060020a0316600160a060020a0316815260200182815260200192505050602060405180830381600087803b15801561072857600080fd5b505af115801561073c573d6000803e3d6000fd5b505050506040513d602081101561075257600080fd5b5051151561075f57600080fd5b6107cb565b604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600e60248201527f696e76616c696420746f6b656e2e000000000000000000000000000000000000604482015290519081900360640190fd5b876040516020018082805190602001908083835b602083106107fe5780518252601f1990920191602091820191016107df565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040526040518082805190602001908083835b602083106108615780518252601f199092019160209182019101610842565b6001836020036101000a03801982511681845116808217855250505050505090500191505060405180910390209050600160066000836000191660001916815260200190815260200160002060006101000a81548160ff02191690831515021790555033600160a060020a03167fe1e95cd62612ab6fe5f5ff7a378b06c8bcc1c30a0892301266af4e34ff8b075f8a8989604051808060200185600160a060020a0316600160a060020a0316815260200184600160a060020a0316600160a060020a03168152602001838152602001828103825260088152602001807f576974684472617700000000000000000000000000000000000000000000000081525060200194505050505060405180910390a2505050505050505050565b60015481565b61098b611987565b8251606090600090151561099e57600080fd5b835115156109ab57600080fd5b600160a060020a0387161515610c7557600034116109c857600080fd5b6040805160c0810182526000815233602082015290810186905260608101859052346080820152601260a08201529250610a01836115a8565b9150816040518082805190602001908083835b60208310610a335780518252601f199092019160209182019101610a14565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051809103902060008060045481526020019081526020016000208160001916905550600160046000828254019250508190555033600160a060020a03167fda4769f5ce852c913b48cfbb0d994d395902782cb1efd516641ab1cfc8c71fbb60016004540360003460128a8a89604051808060200189815260200188600160a060020a0316600160a060020a031681526020018781526020018681526020018060200180602001806020018581038552600a8152602001807f43726f7373436861696e00000000000000000000000000000000000000000000815250602001858103845288818151815260200191508051906020019080838360005b83811015610b6d578181015183820152602001610b55565b50505050905090810190601f168015610b9a5780820380516001836020036101000a031916815260200191505b50858103835287518152875160209182019189019080838360005b83811015610bcd578181015183820152602001610bb5565b50505050905090810190601f168015610bfa5780820380516001836020036101000a031916815260200191505b50858103825286518152865160209182019188019080838360005b83811015610c2d578181015183820152602001610c15565b50505050905090810190601f168015610c5a5780820380516001836020036101000a031916815260200191505b509b50505050505050505050505060405180910390a2611072565b600160a060020a038716156110725760008611610c9157600080fd5b3415610c9c57600080fd5b86600160a060020a031663313ce5676040518163ffffffff1660e060020a028152600401602060405180830381600087803b158015610cda57600080fd5b505af1158015610cee573d6000803e3d6000fd5b505050506040513d6020811015610d0457600080fd5b50519050600060ff821611610d1857600080fd5b604080517f23b872dd000000000000000000000000000000000000000000000000000000008152336004820152306024820152604481018890529051600160a060020a038916916323b872dd9160648083019260209291908290030181600087803b158015610d8657600080fd5b505af1158015610d9a573d6000803e3d6000fd5b505050506040513d6020811015610db057600080fd5b50511515610dbd57600080fd5b6040805160c081018252600160a060020a0389168152336020820152908101869052606081018590526080810187905260ff821660a08201529250610e01836115a8565b9150816040518082805190602001908083835b60208310610e335780518252601f199092019160209182019101610e14565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051809103902060008060045481526020019081526020016000208160001916905550600160046000828254019250508190555033600160a060020a03167fda4769f5ce852c913b48cfbb0d994d395902782cb1efd516641ab1cfc8c71fbb6001600454038989858a8a89604051808060200189815260200188600160a060020a0316600160a060020a031681526020018781526020018660ff1681526020018060200180602001806020018581038552600a8152602001807f43726f7373436861696e00000000000000000000000000000000000000000000815250602001858103845288818151815260200191508051906020019080838360005b83811015610f6e578181015183820152602001610f56565b50505050905090810190601f168015610f9b5780820380516001836020036101000a031916815260200191505b50858103835287518152875160209182019189019080838360005b83811015610fce578181015183820152602001610fb6565b50505050905090810190601f168015610ffb5780820380516001836020036101000a031916815260200191505b50858103825286518152865160209182019188019080838360005b8381101561102e578181015183820152602001611016565b50505050905090810190601f16801561105b5780820380516001836020036101000a031916815260200191505b509b50505050505050505050505060405180910390a25b50505050505050565b604080518082018252601081527f303132333435363738396162636465660000000000000000000000000000000060208201528151602a8082526060828101909452600160a060020a03851692918491600091908160200160208202803883390190505091507f300000000000000000000000000000000000000000000000000000000000000082600081518110151561111157fe5b906020010190600160f860020a031916908160001a9053507f780000000000000000000000000000000000000000000000000000000000000082600181518110151561115957fe5b906020010190600160f860020a031916908160001a905350600090505b60148110156112745782600485600c84016020811061119157fe5b1a60f860020a02600160f860020a0319169060020a900460f860020a90048151811015156111bb57fe5b90602001015160f860020a900460f860020a0282826002026002018151811015156111e257fe5b906020010190600160f860020a031916908160001a9053508284600c83016020811061120a57fe5b1a60f860020a02600f60f860020a021660f860020a900481518110151561122d57fe5b90602001015160f860020a900460f860020a02828260020260030181518110151561125457fe5b906020010190600160f860020a031916908160001a905350600101611176565b8194505b50505050919050565b606060008082818515156112ca5760408051808201909152600181527f300000000000000000000000000000000000000000000000000000000000000060208201529450611278565b8593505b83156112e557600190920191600a840493506112ce565b826040519080825280601f01601f191660200182016040528015611313578160200160208202803883390190505b5091505060001982015b851561127457815160001982019160f860020a6030600a8a06010291849190811061134457fe5b906020010190600160f860020a031916908160001a905350600a8604955061131d565b60008060008060008951602414151561137f57600080fd5b896040516020018082805190602001908083835b602083106113b25780518252601f199092019160209182019101611393565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040526040518082805190602001908083835b602083106114155780518252601f1990920191602091820191016113f6565b51815160209384036101000a6000190180199092169116179052604080519290940182900390912060008181526006909252929020549197505060ff16159150611460905057600080fd5b86518851148015611472575085518751145b151561147d57600080fd5b6001548851101561148d57600080fd5b60009250600091505b600154821015611589576114f18989848151811015156114b257fe5b9060200190602002015189858151811015156114ca57fe5b9060200190602002015189868151811015156114e257fe5b90602001906020020151611826565b600160a060020a03811660009081526003602052604090205490915060ff16151561151b57600080fd5b6000848152600560209081526040808320600160a060020a038516845290915290205460ff161561154b57600080fd5b6000848152600560209081526040808320600160a060020a03851684529091529020805460ff19166001908117909155928301929190910190611496565b60015483101561159857600080fd5b5060019998505050505050505050565b60606115b7826000015161107b565b6115c4836020015161107b565b836040015184606001516115db8660800151611281565b6115e88760a00151611281565b6040516020018087805190602001908083835b6020831061161a5780518252601f1990920191602091820191016115fb565b6001836020036101000a0380198251168184511680821785525050505050509050018060f860020a60230281525060010186805190602001908083835b602083106116765780518252601f199092019160209182019101611657565b6001836020036101000a0380198251168184511680821785525050505050509050018060f860020a60230281525060010185805190602001908083835b602083106116d25780518252601f1990920191602091820191016116b3565b6001836020036101000a0380198251168184511680821785525050505050509050018060f860020a60230281525060010184805190602001908083835b6020831061172e5780518252601f19909201916020918201910161170f565b6001836020036101000a0380198251168184511680821785525050505050509050018060f860020a60230281525060010183805190602001908083835b6020831061178a5780518252601f19909201916020918201910161176b565b6001836020036101000a0380198251168184511680821785525050505050509050018060f860020a60230281525060010182805190602001908083835b602083106117e65780518252601f1990920191602091820191016117c7565b6001836020036101000a03801982511681845116808217855250505050505090500196505050505050506040516020818303038152906040529050919050565b604080518082018252601c8082527f19457468657265756d205369676e6564204d6573736167653a0a33320000000060208084019182529351600094859385938b939092019182918083835b602083106118915780518252601f199092019160209182019101611872565b51815160209384036101000a600019018019909216911617905292019384525060408051808503815293820190819052835193945092839250908401908083835b602083106118f15780518252601f1990920191602091820191016118d2565b51815160209384036101000a600019018019909216911617905260408051929094018290038220600080845283830180875282905260ff8e1684870152606084018d9052608084018c905294519097506001965060a080840196509194601f19820194509281900390910191865af1158015611971573d6000803e3d6000fd5b5050604051601f19015198975050505050505050565b60c0604051908101604052806000600160a060020a031681526020016000600160a060020a031681526020016060815260200160608152602001600081526020016000815250905600a165627a7a72305820412b61db524c7c67a22b1ee1b63469a8b5321df56b7eb465f9f1ca881911cb770029"

		// DeployEthereumCrossChain deploys a new Ethereum contract, binding an instance of EthereumCrossChain to it.
		func DeployEthereumCrossChain(auth *bind.TransactOpts, backend bind.ContractBackend , _owners []common.Address, _required *big.Int) (common.Address, *types.Transaction, *EthereumCrossChain, error) {
		  parsed, err := abi.JSON(strings.NewReader(EthereumCrossChainABI))
		  if err != nil {
		    return common.Address{}, nil, nil, err
		  }
		  
		  address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(EthereumCrossChainBin), backend , _owners, _required)
		  if err != nil {
		    return common.Address{}, nil, nil, err
		  }
		  return address, tx, &EthereumCrossChain{ EthereumCrossChainCaller: EthereumCrossChainCaller{contract: contract}, EthereumCrossChainTransactor: EthereumCrossChainTransactor{contract: contract}, EthereumCrossChainFilterer: EthereumCrossChainFilterer{contract: contract} }, nil
		}
	

	// EthereumCrossChain is an auto generated Go binding around an Ethereum contract.
	type EthereumCrossChain struct {
	  EthereumCrossChainCaller     // Read-only binding to the contract
	  EthereumCrossChainTransactor // Write-only binding to the contract
	  EthereumCrossChainFilterer   // Log filterer for contract events
	}

	// EthereumCrossChainCaller is an auto generated read-only Go binding around an Ethereum contract.
	type EthereumCrossChainCaller struct {
	  contract *bind.BoundContract // Generic contract wrapper for the low level calls
	}

	// EthereumCrossChainTransactor is an auto generated write-only Go binding around an Ethereum contract.
	type EthereumCrossChainTransactor struct {
	  contract *bind.BoundContract // Generic contract wrapper for the low level calls
	}

	// EthereumCrossChainFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
	type EthereumCrossChainFilterer struct {
	  contract *bind.BoundContract // Generic contract wrapper for the low level calls
	}

	// EthereumCrossChainSession is an auto generated Go binding around an Ethereum contract,
	// with pre-set call and transact options.
	type EthereumCrossChainSession struct {
	  Contract     *EthereumCrossChain        // Generic contract binding to set the session for
	  CallOpts     bind.CallOpts     // Call options to use throughout this session
	  TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
	}

	// EthereumCrossChainCallerSession is an auto generated read-only Go binding around an Ethereum contract,
	// with pre-set call options.
	type EthereumCrossChainCallerSession struct {
	  Contract *EthereumCrossChainCaller // Generic contract caller binding to set the session for
	  CallOpts bind.CallOpts    // Call options to use throughout this session
	}

	// EthereumCrossChainTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
	// with pre-set transact options.
	type EthereumCrossChainTransactorSession struct {
	  Contract     *EthereumCrossChainTransactor // Generic contract transactor binding to set the session for
	  TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
	}

	// EthereumCrossChainRaw is an auto generated low-level Go binding around an Ethereum contract.
	type EthereumCrossChainRaw struct {
	  Contract *EthereumCrossChain // Generic contract binding to access the raw methods on
	}

	// EthereumCrossChainCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
	type EthereumCrossChainCallerRaw struct {
		Contract *EthereumCrossChainCaller // Generic read-only contract binding to access the raw methods on
	}

	// EthereumCrossChainTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
	type EthereumCrossChainTransactorRaw struct {
		Contract *EthereumCrossChainTransactor // Generic write-only contract binding to access the raw methods on
	}

	// NewEthereumCrossChain creates a new instance of EthereumCrossChain, bound to a specific deployed contract.
	func NewEthereumCrossChain(address common.Address, backend bind.ContractBackend) (*EthereumCrossChain, error) {
	  contract, err := bindEthereumCrossChain(address, backend, backend, backend)
	  if err != nil {
	    return nil, err
	  }
	  return &EthereumCrossChain{ EthereumCrossChainCaller: EthereumCrossChainCaller{contract: contract}, EthereumCrossChainTransactor: EthereumCrossChainTransactor{contract: contract}, EthereumCrossChainFilterer: EthereumCrossChainFilterer{contract: contract} }, nil
	}

	// NewEthereumCrossChainCaller creates a new read-only instance of EthereumCrossChain, bound to a specific deployed contract.
	func NewEthereumCrossChainCaller(address common.Address, caller bind.ContractCaller) (*EthereumCrossChainCaller, error) {
	  contract, err := bindEthereumCrossChain(address, caller, nil, nil)
	  if err != nil {
	    return nil, err
	  }
	  return &EthereumCrossChainCaller{contract: contract}, nil
	}

	// NewEthereumCrossChainTransactor creates a new write-only instance of EthereumCrossChain, bound to a specific deployed contract.
	func NewEthereumCrossChainTransactor(address common.Address, transactor bind.ContractTransactor) (*EthereumCrossChainTransactor, error) {
	  contract, err := bindEthereumCrossChain(address, nil, transactor, nil)
	  if err != nil {
	    return nil, err
	  }
	  return &EthereumCrossChainTransactor{contract: contract}, nil
	}

	// NewEthereumCrossChainFilterer creates a new log filterer instance of EthereumCrossChain, bound to a specific deployed contract.
 	func NewEthereumCrossChainFilterer(address common.Address, filterer bind.ContractFilterer) (*EthereumCrossChainFilterer, error) {
 	  contract, err := bindEthereumCrossChain(address, nil, nil, filterer)
 	  if err != nil {
 	    return nil, err
 	  }
 	  return &EthereumCrossChainFilterer{contract: contract}, nil
 	}

	// bindEthereumCrossChain binds a generic wrapper to an already deployed contract.
	func bindEthereumCrossChain(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	  parsed, err := abi.JSON(strings.NewReader(EthereumCrossChainABI))
	  if err != nil {
	    return nil, err
	  }
	  return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
	}

	// Call invokes the (constant) contract method with params as input values and
	// sets the output to result. The result type might be a single field for simple
	// returns, a slice of interfaces for anonymous returns and a struct for named
	// returns.
	func (_EthereumCrossChain *EthereumCrossChainRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
		return _EthereumCrossChain.Contract.EthereumCrossChainCaller.contract.Call(opts, result, method, params...)
	}

	// Transfer initiates a plain transaction to move funds to the contract, calling
	// its default method if one is available.
	func (_EthereumCrossChain *EthereumCrossChainRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
		return _EthereumCrossChain.Contract.EthereumCrossChainTransactor.contract.Transfer(opts)
	}

	// Transact invokes the (paid) contract method with params as input values.
	func (_EthereumCrossChain *EthereumCrossChainRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
		return _EthereumCrossChain.Contract.EthereumCrossChainTransactor.contract.Transact(opts, method, params...)
	}

	// Call invokes the (constant) contract method with params as input values and
	// sets the output to result. The result type might be a single field for simple
	// returns, a slice of interfaces for anonymous returns and a struct for named
	// returns.
	func (_EthereumCrossChain *EthereumCrossChainCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
		return _EthereumCrossChain.Contract.contract.Call(opts, result, method, params...)
	}

	// Transfer initiates a plain transaction to move funds to the contract, calling
	// its default method if one is available.
	func (_EthereumCrossChain *EthereumCrossChainTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
		return _EthereumCrossChain.Contract.contract.Transfer(opts)
	}

	// Transact invokes the (paid) contract method with params as input values.
	func (_EthereumCrossChain *EthereumCrossChainTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
		return _EthereumCrossChain.Contract.contract.Transact(opts, method, params...)
	}

	

	
		// Owners is a free data retrieval call binding the contract method 0x025e7c27.
		//
		// Solidity: function owners(uint256 ) constant returns(address)
		func (_EthereumCrossChain *EthereumCrossChainCaller) Owners(opts *bind.CallOpts , arg0 *big.Int ) (common.Address, error) {
			var (
				ret0 = new(common.Address)
				
			)
			out := ret0
			err := _EthereumCrossChain.contract.Call(opts, out, "owners" , arg0)
			return *ret0, err
		}

		// Owners is a free data retrieval call binding the contract method 0x025e7c27.
		//
		// Solidity: function owners(uint256 ) constant returns(address)
		func (_EthereumCrossChain *EthereumCrossChainSession) Owners( arg0 *big.Int ) ( common.Address,  error) {
		  return _EthereumCrossChain.Contract.Owners(&_EthereumCrossChain.CallOpts , arg0)
		}

		// Owners is a free data retrieval call binding the contract method 0x025e7c27.
		//
		// Solidity: function owners(uint256 ) constant returns(address)
		func (_EthereumCrossChain *EthereumCrossChainCallerSession) Owners( arg0 *big.Int ) ( common.Address,  error) {
		  return _EthereumCrossChain.Contract.Owners(&_EthereumCrossChain.CallOpts , arg0)
		}
	
		// Required is a free data retrieval call binding the contract method 0xdc8452cd.
		//
		// Solidity: function required() constant returns(uint256)
		func (_EthereumCrossChain *EthereumCrossChainCaller) Required(opts *bind.CallOpts ) (*big.Int, error) {
			var (
				ret0 = new(*big.Int)
				
			)
			out := ret0
			err := _EthereumCrossChain.contract.Call(opts, out, "required" )
			return *ret0, err
		}

		// Required is a free data retrieval call binding the contract method 0xdc8452cd.
		//
		// Solidity: function required() constant returns(uint256)
		func (_EthereumCrossChain *EthereumCrossChainSession) Required() ( *big.Int,  error) {
		  return _EthereumCrossChain.Contract.Required(&_EthereumCrossChain.CallOpts )
		}

		// Required is a free data retrieval call binding the contract method 0xdc8452cd.
		//
		// Solidity: function required() constant returns(uint256)
		func (_EthereumCrossChain *EthereumCrossChainCallerSession) Required() ( *big.Int,  error) {
		  return _EthereumCrossChain.Contract.Required(&_EthereumCrossChain.CallOpts )
		}
	
		// TransactionId is a free data retrieval call binding the contract method 0x7e2f42e7.
		//
		// Solidity: function transactionId() constant returns(uint256)
		func (_EthereumCrossChain *EthereumCrossChainCaller) TransactionId(opts *bind.CallOpts ) (*big.Int, error) {
			var (
				ret0 = new(*big.Int)
				
			)
			out := ret0
			err := _EthereumCrossChain.contract.Call(opts, out, "transactionId" )
			return *ret0, err
		}

		// TransactionId is a free data retrieval call binding the contract method 0x7e2f42e7.
		//
		// Solidity: function transactionId() constant returns(uint256)
		func (_EthereumCrossChain *EthereumCrossChainSession) TransactionId() ( *big.Int,  error) {
		  return _EthereumCrossChain.Contract.TransactionId(&_EthereumCrossChain.CallOpts )
		}

		// TransactionId is a free data retrieval call binding the contract method 0x7e2f42e7.
		//
		// Solidity: function transactionId() constant returns(uint256)
		func (_EthereumCrossChain *EthereumCrossChainCallerSession) TransactionId() ( *big.Int,  error) {
		  return _EthereumCrossChain.Contract.TransactionId(&_EthereumCrossChain.CallOpts )
		}
	
		// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
		//
		// Solidity: function transactions(uint256 ) constant returns(bytes32)
		func (_EthereumCrossChain *EthereumCrossChainCaller) Transactions(opts *bind.CallOpts , arg0 *big.Int ) ([32]byte, error) {
			var (
				ret0 = new([32]byte)
				
			)
			out := ret0
			err := _EthereumCrossChain.contract.Call(opts, out, "transactions" , arg0)
			return *ret0, err
		}

		// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
		//
		// Solidity: function transactions(uint256 ) constant returns(bytes32)
		func (_EthereumCrossChain *EthereumCrossChainSession) Transactions( arg0 *big.Int ) ( [32]byte,  error) {
		  return _EthereumCrossChain.Contract.Transactions(&_EthereumCrossChain.CallOpts , arg0)
		}

		// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
		//
		// Solidity: function transactions(uint256 ) constant returns(bytes32)
		func (_EthereumCrossChain *EthereumCrossChainCallerSession) Transactions( arg0 *big.Int ) ( [32]byte,  error) {
		  return _EthereumCrossChain.Contract.Transactions(&_EthereumCrossChain.CallOpts , arg0)
		}
	

	
		// CrossChain is a paid mutator transaction binding the contract method 0xfafa0765.
		//
		// Solidity: function CrossChain(address _token, uint256 _value, string _toChainId, string _toAddress) returns()
		func (_EthereumCrossChain *EthereumCrossChainTransactor) CrossChain(opts *bind.TransactOpts , _token common.Address , _value *big.Int , _toChainId string , _toAddress string ) (*types.Transaction, error) {
			return _EthereumCrossChain.contract.Transact(opts, "CrossChain" , _token, _value, _toChainId, _toAddress)
		}

		// CrossChain is a paid mutator transaction binding the contract method 0xfafa0765.
		//
		// Solidity: function CrossChain(address _token, uint256 _value, string _toChainId, string _toAddress) returns()
		func (_EthereumCrossChain *EthereumCrossChainSession) CrossChain( _token common.Address , _value *big.Int , _toChainId string , _toAddress string ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.CrossChain(&_EthereumCrossChain.TransactOpts , _token, _value, _toChainId, _toAddress)
		}

		// CrossChain is a paid mutator transaction binding the contract method 0xfafa0765.
		//
		// Solidity: function CrossChain(address _token, uint256 _value, string _toChainId, string _toAddress) returns()
		func (_EthereumCrossChain *EthereumCrossChainTransactorSession) CrossChain( _token common.Address , _value *big.Int , _toChainId string , _toAddress string ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.CrossChain(&_EthereumCrossChain.TransactOpts , _token, _value, _toChainId, _toAddress)
		}
	
		// Withdraw is a paid mutator transaction binding the contract method 0xb70d4ee9.
		//
		// Solidity: function Withdraw(address _token, string _txId, address _to, uint256 _value, uint8[] _v, bytes32[] _r, bytes32[] _s) returns()
		func (_EthereumCrossChain *EthereumCrossChainTransactor) Withdraw(opts *bind.TransactOpts , _token common.Address , _txId string , _to common.Address , _value *big.Int , _v []uint8 , _r [][32]byte , _s [][32]byte ) (*types.Transaction, error) {
			return _EthereumCrossChain.contract.Transact(opts, "Withdraw" , _token, _txId, _to, _value, _v, _r, _s)
		}

		// Withdraw is a paid mutator transaction binding the contract method 0xb70d4ee9.
		//
		// Solidity: function Withdraw(address _token, string _txId, address _to, uint256 _value, uint8[] _v, bytes32[] _r, bytes32[] _s) returns()
		func (_EthereumCrossChain *EthereumCrossChainSession) Withdraw( _token common.Address , _txId string , _to common.Address , _value *big.Int , _v []uint8 , _r [][32]byte , _s [][32]byte ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.Withdraw(&_EthereumCrossChain.TransactOpts , _token, _txId, _to, _value, _v, _r, _s)
		}

		// Withdraw is a paid mutator transaction binding the contract method 0xb70d4ee9.
		//
		// Solidity: function Withdraw(address _token, string _txId, address _to, uint256 _value, uint8[] _v, bytes32[] _r, bytes32[] _s) returns()
		func (_EthereumCrossChain *EthereumCrossChainTransactorSession) Withdraw( _token common.Address , _txId string , _to common.Address , _value *big.Int , _v []uint8 , _r [][32]byte , _s [][32]byte ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.Withdraw(&_EthereumCrossChain.TransactOpts , _token, _txId, _to, _value, _v, _r, _s)
		}
	

	
		// EthereumCrossChainCrossChainEventIterator is returned from FilterCrossChainEvent and is used to iterate over the raw logs and unpacked data for CrossChainEvent events raised by the EthereumCrossChain contract.
		type EthereumCrossChainCrossChainEventIterator struct {
			Event *EthereumCrossChainCrossChainEvent // Event containing the contract specifics and raw log

			contract *bind.BoundContract // Generic contract to use for unpacking event data
			event    string              // Event name to use for unpacking event data

			logs chan types.Log        // Log channel receiving the found contract events
			sub  ethereum.Subscription // Subscription for errors, completion and termination
			done bool                  // Whether the subscription completed delivering logs
			fail error                 // Occurred error to stop iteration
		}
		// Next advances the iterator to the subsequent event, returning whether there
		// are any more events found. In case of a retrieval or parsing error, false is
		// returned and Error() can be queried for the exact failure.
		func (it *EthereumCrossChainCrossChainEventIterator) Next() bool {
			// If the iterator failed, stop iterating
			if (it.fail != nil) {
				return false
			}
			// If the iterator completed, deliver directly whatever's available
			if (it.done) {
				select {
				case log := <-it.logs:
					it.Event = new(EthereumCrossChainCrossChainEvent)
					if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
						it.fail = err
						return false
					}
					it.Event.Raw = log
					return true

				default:
					return false
				}
			}
			// Iterator still in progress, wait for either a data or an error event
			select {
			case log := <-it.logs:
				it.Event = new(EthereumCrossChainCrossChainEvent)
				if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
					it.fail = err
					return false
				}
				it.Event.Raw = log
				return true

			case err := <-it.sub.Err():
				it.done = true
				it.fail = err
				return it.Next()
			}
		}
		// Error returns any retrieval or parsing error occurred during filtering.
		func (it *EthereumCrossChainCrossChainEventIterator) Error() error {
			return it.fail
		}
		// Close terminates the iteration process, releasing any pending underlying
		// resources.
		func (it *EthereumCrossChainCrossChainEventIterator) Close() error {
			it.sub.Unsubscribe()
			return nil
		}

		// EthereumCrossChainCrossChainEvent represents a CrossChainEvent event raised by the EthereumCrossChain contract.
		type EthereumCrossChainCrossChainEvent struct { 
			Topic string; 
			Sender common.Address; 
			TxId *big.Int; 
			Token common.Address; 
			Value *big.Int; 
			Decimal *big.Int; 
			ToChainId string; 
			ToAddress string; 
			Rawdata string; 
			Raw types.Log // Blockchain specific contextual infos
		}

		// FilterCrossChainEvent is a free log retrieval operation binding the contract event 0xda4769f5ce852c913b48cfbb0d994d395902782cb1efd516641ab1cfc8c71fbb.
		//
		// Solidity: event CrossChainEvent(string topic, address indexed sender, uint256 txId, address token, uint256 value, uint256 decimal, string toChainId, string toAddress, string rawdata)
 		func (_EthereumCrossChain *EthereumCrossChainFilterer) FilterCrossChainEvent(opts *bind.FilterOpts, sender []common.Address) (*EthereumCrossChainCrossChainEventIterator, error) {
			
			
			var senderRule []interface{}
			for _, senderItem := range sender {
				senderRule = append(senderRule, senderItem)
			}
			
			
			
			
			
			
			

			logs, sub, err := _EthereumCrossChain.contract.FilterLogs(opts, "CrossChainEvent", senderRule)
			if err != nil {
				return nil, err
			}
			return &EthereumCrossChainCrossChainEventIterator{contract: _EthereumCrossChain.contract, event: "CrossChainEvent", logs: logs, sub: sub}, nil
 		}

		// WatchCrossChainEvent is a free log subscription operation binding the contract event 0xda4769f5ce852c913b48cfbb0d994d395902782cb1efd516641ab1cfc8c71fbb.
		//
		// Solidity: event CrossChainEvent(string topic, address indexed sender, uint256 txId, address token, uint256 value, uint256 decimal, string toChainId, string toAddress, string rawdata)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) WatchCrossChainEvent(opts *bind.WatchOpts, sink chan<- *EthereumCrossChainCrossChainEvent, sender []common.Address) (event.Subscription, error) {
			
			
			var senderRule []interface{}
			for _, senderItem := range sender {
				senderRule = append(senderRule, senderItem)
			}
			
			
			
			
			
			
			

			logs, sub, err := _EthereumCrossChain.contract.WatchLogs(opts, "CrossChainEvent", senderRule)
			if err != nil {
				return nil, err
			}
			return event.NewSubscription(func(quit <-chan struct{}) error {
				defer sub.Unsubscribe()
				for {
					select {
					case log := <-logs:
						// New log arrived, parse the event and forward to the user
						event := new(EthereumCrossChainCrossChainEvent)
						if err := _EthereumCrossChain.contract.UnpackLog(event, "CrossChainEvent", log); err != nil {
							return err
						}
						event.Raw = log

						select {
						case sink <- event:
						case err := <-sub.Err():
							return err
						case <-quit:
							return nil
						}
					case err := <-sub.Err():
						return err
					case <-quit:
						return nil
					}
				}
			}), nil
		}

		// ParseCrossChainEvent is a log parse operation binding the contract event 0xda4769f5ce852c913b48cfbb0d994d395902782cb1efd516641ab1cfc8c71fbb.
		//
		// Solidity: event CrossChainEvent(string topic, address indexed sender, uint256 txId, address token, uint256 value, uint256 decimal, string toChainId, string toAddress, string rawdata)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) ParseCrossChainEvent(log types.Log) (*EthereumCrossChainCrossChainEvent, error) {
			event := new(EthereumCrossChainCrossChainEvent)
			if err := _EthereumCrossChain.contract.UnpackLog(event, "CrossChainEvent", log); err != nil {
				return nil, err
			}
			return event, nil
		}

 	
		// EthereumCrossChainDebugByteIterator is returned from FilterDebugByte and is used to iterate over the raw logs and unpacked data for DebugByte events raised by the EthereumCrossChain contract.
		type EthereumCrossChainDebugByteIterator struct {
			Event *EthereumCrossChainDebugByte // Event containing the contract specifics and raw log

			contract *bind.BoundContract // Generic contract to use for unpacking event data
			event    string              // Event name to use for unpacking event data

			logs chan types.Log        // Log channel receiving the found contract events
			sub  ethereum.Subscription // Subscription for errors, completion and termination
			done bool                  // Whether the subscription completed delivering logs
			fail error                 // Occurred error to stop iteration
		}
		// Next advances the iterator to the subsequent event, returning whether there
		// are any more events found. In case of a retrieval or parsing error, false is
		// returned and Error() can be queried for the exact failure.
		func (it *EthereumCrossChainDebugByteIterator) Next() bool {
			// If the iterator failed, stop iterating
			if (it.fail != nil) {
				return false
			}
			// If the iterator completed, deliver directly whatever's available
			if (it.done) {
				select {
				case log := <-it.logs:
					it.Event = new(EthereumCrossChainDebugByte)
					if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
						it.fail = err
						return false
					}
					it.Event.Raw = log
					return true

				default:
					return false
				}
			}
			// Iterator still in progress, wait for either a data or an error event
			select {
			case log := <-it.logs:
				it.Event = new(EthereumCrossChainDebugByte)
				if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
					it.fail = err
					return false
				}
				it.Event.Raw = log
				return true

			case err := <-it.sub.Err():
				it.done = true
				it.fail = err
				return it.Next()
			}
		}
		// Error returns any retrieval or parsing error occurred during filtering.
		func (it *EthereumCrossChainDebugByteIterator) Error() error {
			return it.fail
		}
		// Close terminates the iteration process, releasing any pending underlying
		// resources.
		func (it *EthereumCrossChainDebugByteIterator) Close() error {
			it.sub.Unsubscribe()
			return nil
		}

		// EthereumCrossChainDebugByte represents a DebugByte event raised by the EthereumCrossChain contract.
		type EthereumCrossChainDebugByte struct { 
			b []byte;
			Raw types.Log // Blockchain specific contextual infos
		}

		// FilterDebugByte is a free log retrieval operation binding the contract event 0xe257d31e31b20bc48128c69b5a1b884d412f5d92048696b9a29e0b15625755da.
		//
		// Solidity: event DebugByte(bytes )
 		func (_EthereumCrossChain *EthereumCrossChainFilterer) FilterDebugByte(opts *bind.FilterOpts) (*EthereumCrossChainDebugByteIterator, error) {
			
			

			logs, sub, err := _EthereumCrossChain.contract.FilterLogs(opts, "DebugByte")
			if err != nil {
				return nil, err
			}
			return &EthereumCrossChainDebugByteIterator{contract: _EthereumCrossChain.contract, event: "DebugByte", logs: logs, sub: sub}, nil
 		}

		// WatchDebugByte is a free log subscription operation binding the contract event 0xe257d31e31b20bc48128c69b5a1b884d412f5d92048696b9a29e0b15625755da.
		//
		// Solidity: event DebugByte(bytes )
		func (_EthereumCrossChain *EthereumCrossChainFilterer) WatchDebugByte(opts *bind.WatchOpts, sink chan<- *EthereumCrossChainDebugByte) (event.Subscription, error) {
			
			

			logs, sub, err := _EthereumCrossChain.contract.WatchLogs(opts, "DebugByte")
			if err != nil {
				return nil, err
			}
			return event.NewSubscription(func(quit <-chan struct{}) error {
				defer sub.Unsubscribe()
				for {
					select {
					case log := <-logs:
						// New log arrived, parse the event and forward to the user
						event := new(EthereumCrossChainDebugByte)
						if err := _EthereumCrossChain.contract.UnpackLog(event, "DebugByte", log); err != nil {
							return err
						}
						event.Raw = log

						select {
						case sink <- event:
						case err := <-sub.Err():
							return err
						case <-quit:
							return nil
						}
					case err := <-sub.Err():
						return err
					case <-quit:
						return nil
					}
				}
			}), nil
		}

		// ParseDebugByte is a log parse operation binding the contract event 0xe257d31e31b20bc48128c69b5a1b884d412f5d92048696b9a29e0b15625755da.
		//
		// Solidity: event DebugByte(bytes )
		func (_EthereumCrossChain *EthereumCrossChainFilterer) ParseDebugByte(log types.Log) (*EthereumCrossChainDebugByte, error) {
			event := new(EthereumCrossChainDebugByte)
			if err := _EthereumCrossChain.contract.UnpackLog(event, "DebugByte", log); err != nil {
				return nil, err
			}
			return event, nil
		}

 	
		// EthereumCrossChainDebugsIterator is returned from FilterDebugs and is used to iterate over the raw logs and unpacked data for Debugs events raised by the EthereumCrossChain contract.
		type EthereumCrossChainDebugsIterator struct {
			Event *EthereumCrossChainDebugs // Event containing the contract specifics and raw log

			contract *bind.BoundContract // Generic contract to use for unpacking event data
			event    string              // Event name to use for unpacking event data

			logs chan types.Log        // Log channel receiving the found contract events
			sub  ethereum.Subscription // Subscription for errors, completion and termination
			done bool                  // Whether the subscription completed delivering logs
			fail error                 // Occurred error to stop iteration
		}
		// Next advances the iterator to the subsequent event, returning whether there
		// are any more events found. In case of a retrieval or parsing error, false is
		// returned and Error() can be queried for the exact failure.
		func (it *EthereumCrossChainDebugsIterator) Next() bool {
			// If the iterator failed, stop iterating
			if (it.fail != nil) {
				return false
			}
			// If the iterator completed, deliver directly whatever's available
			if (it.done) {
				select {
				case log := <-it.logs:
					it.Event = new(EthereumCrossChainDebugs)
					if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
						it.fail = err
						return false
					}
					it.Event.Raw = log
					return true

				default:
					return false
				}
			}
			// Iterator still in progress, wait for either a data or an error event
			select {
			case log := <-it.logs:
				it.Event = new(EthereumCrossChainDebugs)
				if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
					it.fail = err
					return false
				}
				it.Event.Raw = log
				return true

			case err := <-it.sub.Err():
				it.done = true
				it.fail = err
				return it.Next()
			}
		}
		// Error returns any retrieval or parsing error occurred during filtering.
		func (it *EthereumCrossChainDebugsIterator) Error() error {
			return it.fail
		}
		// Close terminates the iteration process, releasing any pending underlying
		// resources.
		func (it *EthereumCrossChainDebugsIterator) Close() error {
			it.sub.Unsubscribe()
			return nil
		}

		// EthereumCrossChainDebugs represents a Debugs event raised by the EthereumCrossChain contract.
		type EthereumCrossChainDebugs struct { 
			 string; 
			Raw types.Log // Blockchain specific contextual infos
		}

		// FilterDebugs is a free log retrieval operation binding the contract event 0xe406140f9f0bc5d159086a5162bf2b79e5e7f6c4413b136022c1307dd062aa3c.
		//
		// Solidity: event Debugs(string )
 		func (_EthereumCrossChain *EthereumCrossChainFilterer) FilterDebugs(opts *bind.FilterOpts) (*EthereumCrossChainDebugsIterator, error) {
			
			

			logs, sub, err := _EthereumCrossChain.contract.FilterLogs(opts, "Debugs")
			if err != nil {
				return nil, err
			}
			return &EthereumCrossChainDebugsIterator{contract: _EthereumCrossChain.contract, event: "Debugs", logs: logs, sub: sub}, nil
 		}

		// WatchDebugs is a free log subscription operation binding the contract event 0xe406140f9f0bc5d159086a5162bf2b79e5e7f6c4413b136022c1307dd062aa3c.
		//
		// Solidity: event Debugs(string )
		func (_EthereumCrossChain *EthereumCrossChainFilterer) WatchDebugs(opts *bind.WatchOpts, sink chan<- *EthereumCrossChainDebugs) (event.Subscription, error) {
			
			

			logs, sub, err := _EthereumCrossChain.contract.WatchLogs(opts, "Debugs")
			if err != nil {
				return nil, err
			}
			return event.NewSubscription(func(quit <-chan struct{}) error {
				defer sub.Unsubscribe()
				for {
					select {
					case log := <-logs:
						// New log arrived, parse the event and forward to the user
						event := new(EthereumCrossChainDebugs)
						if err := _EthereumCrossChain.contract.UnpackLog(event, "Debugs", log); err != nil {
							return err
						}
						event.Raw = log

						select {
						case sink <- event:
						case err := <-sub.Err():
							return err
						case <-quit:
							return nil
						}
					case err := <-sub.Err():
						return err
					case <-quit:
						return nil
					}
				}
			}), nil
		}

		// ParseDebugs is a log parse operation binding the contract event 0xe406140f9f0bc5d159086a5162bf2b79e5e7f6c4413b136022c1307dd062aa3c.
		//
		// Solidity: event Debugs(string )
		func (_EthereumCrossChain *EthereumCrossChainFilterer) ParseDebugs(log types.Log) (*EthereumCrossChainDebugs, error) {
			event := new(EthereumCrossChainDebugs)
			if err := _EthereumCrossChain.contract.UnpackLog(event, "Debugs", log); err != nil {
				return nil, err
			}
			return event, nil
		}

 	
		// EthereumCrossChainWithDrawEventIterator is returned from FilterWithDrawEvent and is used to iterate over the raw logs and unpacked data for WithDrawEvent events raised by the EthereumCrossChain contract.
		type EthereumCrossChainWithDrawEventIterator struct {
			Event *EthereumCrossChainWithDrawEvent // Event containing the contract specifics and raw log

			contract *bind.BoundContract // Generic contract to use for unpacking event data
			event    string              // Event name to use for unpacking event data

			logs chan types.Log        // Log channel receiving the found contract events
			sub  ethereum.Subscription // Subscription for errors, completion and termination
			done bool                  // Whether the subscription completed delivering logs
			fail error                 // Occurred error to stop iteration
		}
		// Next advances the iterator to the subsequent event, returning whether there
		// are any more events found. In case of a retrieval or parsing error, false is
		// returned and Error() can be queried for the exact failure.
		func (it *EthereumCrossChainWithDrawEventIterator) Next() bool {
			// If the iterator failed, stop iterating
			if (it.fail != nil) {
				return false
			}
			// If the iterator completed, deliver directly whatever's available
			if (it.done) {
				select {
				case log := <-it.logs:
					it.Event = new(EthereumCrossChainWithDrawEvent)
					if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
						it.fail = err
						return false
					}
					it.Event.Raw = log
					return true

				default:
					return false
				}
			}
			// Iterator still in progress, wait for either a data or an error event
			select {
			case log := <-it.logs:
				it.Event = new(EthereumCrossChainWithDrawEvent)
				if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
					it.fail = err
					return false
				}
				it.Event.Raw = log
				return true

			case err := <-it.sub.Err():
				it.done = true
				it.fail = err
				return it.Next()
			}
		}
		// Error returns any retrieval or parsing error occurred during filtering.
		func (it *EthereumCrossChainWithDrawEventIterator) Error() error {
			return it.fail
		}
		// Close terminates the iteration process, releasing any pending underlying
		// resources.
		func (it *EthereumCrossChainWithDrawEventIterator) Close() error {
			it.sub.Unsubscribe()
			return nil
		}

		// EthereumCrossChainWithDrawEvent represents a WithDrawEvent event raised by the EthereumCrossChain contract.
		type EthereumCrossChainWithDrawEvent struct { 
			Topic string; 
			Sender common.Address; 
			Token common.Address; 
			ToAddress common.Address; 
			Value *big.Int; 
			Raw types.Log // Blockchain specific contextual infos
		}

		// FilterWithDrawEvent is a free log retrieval operation binding the contract event 0xe1e95cd62612ab6fe5f5ff7a378b06c8bcc1c30a0892301266af4e34ff8b075f.
		//
		// Solidity: event WithDrawEvent(string topic, address indexed sender, address token, address toAddress, uint256 value)
 		func (_EthereumCrossChain *EthereumCrossChainFilterer) FilterWithDrawEvent(opts *bind.FilterOpts, sender []common.Address) (*EthereumCrossChainWithDrawEventIterator, error) {
			
			
			var senderRule []interface{}
			for _, senderItem := range sender {
				senderRule = append(senderRule, senderItem)
			}
			
			
			

			logs, sub, err := _EthereumCrossChain.contract.FilterLogs(opts, "WithDrawEvent", senderRule)
			if err != nil {
				return nil, err
			}
			return &EthereumCrossChainWithDrawEventIterator{contract: _EthereumCrossChain.contract, event: "WithDrawEvent", logs: logs, sub: sub}, nil
 		}

		// WatchWithDrawEvent is a free log subscription operation binding the contract event 0xe1e95cd62612ab6fe5f5ff7a378b06c8bcc1c30a0892301266af4e34ff8b075f.
		//
		// Solidity: event WithDrawEvent(string topic, address indexed sender, address token, address toAddress, uint256 value)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) WatchWithDrawEvent(opts *bind.WatchOpts, sink chan<- *EthereumCrossChainWithDrawEvent, sender []common.Address) (event.Subscription, error) {
			
			
			var senderRule []interface{}
			for _, senderItem := range sender {
				senderRule = append(senderRule, senderItem)
			}
			
			
			

			logs, sub, err := _EthereumCrossChain.contract.WatchLogs(opts, "WithDrawEvent", senderRule)
			if err != nil {
				return nil, err
			}
			return event.NewSubscription(func(quit <-chan struct{}) error {
				defer sub.Unsubscribe()
				for {
					select {
					case log := <-logs:
						// New log arrived, parse the event and forward to the user
						event := new(EthereumCrossChainWithDrawEvent)
						if err := _EthereumCrossChain.contract.UnpackLog(event, "WithDrawEvent", log); err != nil {
							return err
						}
						event.Raw = log

						select {
						case sink <- event:
						case err := <-sub.Err():
							return err
						case <-quit:
							return nil
						}
					case err := <-sub.Err():
						return err
					case <-quit:
						return nil
					}
				}
			}), nil
		}

		// ParseWithDrawEvent is a log parse operation binding the contract event 0xe1e95cd62612ab6fe5f5ff7a378b06c8bcc1c30a0892301266af4e34ff8b075f.
		//
		// Solidity: event WithDrawEvent(string topic, address indexed sender, address token, address toAddress, uint256 value)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) ParseWithDrawEvent(log types.Log) (*EthereumCrossChainWithDrawEvent, error) {
			event := new(EthereumCrossChainWithDrawEvent)
			if err := _EthereumCrossChain.contract.UnpackLog(event, "WithDrawEvent", log); err != nil {
				return nil, err
			}
			return event, nil
		}

 	

	
	// IERC20ABI is the input ABI used to generate the binding from.
	const IERC20ABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"spender\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"sender\",\"type\":\"address\"},{\"name\":\"recipient\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"decimals\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"recipient\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

	
		// IERC20FuncSigs maps the 4-byte function signature to its string representation.
		var IERC20FuncSigs = map[string]string{
			"dd62ed3e": "allowance(address,address)",
			"095ea7b3": "approve(address,uint256)",
			"70a08231": "balanceOf(address)",
			"313ce567": "decimals()",
			"06fdde03": "name()",
			"95d89b41": "symbol()",
			"18160ddd": "totalSupply()",
			"a9059cbb": "transfer(address,uint256)",
			"23b872dd": "transferFrom(address,address,uint256)",
			
		}
	

	

	// IERC20 is an auto generated Go binding around an Ethereum contract.
	type IERC20 struct {
	  IERC20Caller     // Read-only binding to the contract
	  IERC20Transactor // Write-only binding to the contract
	  IERC20Filterer   // Log filterer for contract events
	}

	// IERC20Caller is an auto generated read-only Go binding around an Ethereum contract.
	type IERC20Caller struct {
	  contract *bind.BoundContract // Generic contract wrapper for the low level calls
	}

	// IERC20Transactor is an auto generated write-only Go binding around an Ethereum contract.
	type IERC20Transactor struct {
	  contract *bind.BoundContract // Generic contract wrapper for the low level calls
	}

	// IERC20Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
	type IERC20Filterer struct {
	  contract *bind.BoundContract // Generic contract wrapper for the low level calls
	}

	// IERC20Session is an auto generated Go binding around an Ethereum contract,
	// with pre-set call and transact options.
	type IERC20Session struct {
	  Contract     *IERC20        // Generic contract binding to set the session for
	  CallOpts     bind.CallOpts     // Call options to use throughout this session
	  TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
	}

	// IERC20CallerSession is an auto generated read-only Go binding around an Ethereum contract,
	// with pre-set call options.
	type IERC20CallerSession struct {
	  Contract *IERC20Caller // Generic contract caller binding to set the session for
	  CallOpts bind.CallOpts    // Call options to use throughout this session
	}

	// IERC20TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
	// with pre-set transact options.
	type IERC20TransactorSession struct {
	  Contract     *IERC20Transactor // Generic contract transactor binding to set the session for
	  TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
	}

	// IERC20Raw is an auto generated low-level Go binding around an Ethereum contract.
	type IERC20Raw struct {
	  Contract *IERC20 // Generic contract binding to access the raw methods on
	}

	// IERC20CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
	type IERC20CallerRaw struct {
		Contract *IERC20Caller // Generic read-only contract binding to access the raw methods on
	}

	// IERC20TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
	type IERC20TransactorRaw struct {
		Contract *IERC20Transactor // Generic write-only contract binding to access the raw methods on
	}

	// NewIERC20 creates a new instance of IERC20, bound to a specific deployed contract.
	func NewIERC20(address common.Address, backend bind.ContractBackend) (*IERC20, error) {
	  contract, err := bindIERC20(address, backend, backend, backend)
	  if err != nil {
	    return nil, err
	  }
	  return &IERC20{ IERC20Caller: IERC20Caller{contract: contract}, IERC20Transactor: IERC20Transactor{contract: contract}, IERC20Filterer: IERC20Filterer{contract: contract} }, nil
	}

	// NewIERC20Caller creates a new read-only instance of IERC20, bound to a specific deployed contract.
	func NewIERC20Caller(address common.Address, caller bind.ContractCaller) (*IERC20Caller, error) {
	  contract, err := bindIERC20(address, caller, nil, nil)
	  if err != nil {
	    return nil, err
	  }
	  return &IERC20Caller{contract: contract}, nil
	}

	// NewIERC20Transactor creates a new write-only instance of IERC20, bound to a specific deployed contract.
	func NewIERC20Transactor(address common.Address, transactor bind.ContractTransactor) (*IERC20Transactor, error) {
	  contract, err := bindIERC20(address, nil, transactor, nil)
	  if err != nil {
	    return nil, err
	  }
	  return &IERC20Transactor{contract: contract}, nil
	}

	// NewIERC20Filterer creates a new log filterer instance of IERC20, bound to a specific deployed contract.
 	func NewIERC20Filterer(address common.Address, filterer bind.ContractFilterer) (*IERC20Filterer, error) {
 	  contract, err := bindIERC20(address, nil, nil, filterer)
 	  if err != nil {
 	    return nil, err
 	  }
 	  return &IERC20Filterer{contract: contract}, nil
 	}

	// bindIERC20 binds a generic wrapper to an already deployed contract.
	func bindIERC20(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	  parsed, err := abi.JSON(strings.NewReader(IERC20ABI))
	  if err != nil {
	    return nil, err
	  }
	  return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
	}

	// Call invokes the (constant) contract method with params as input values and
	// sets the output to result. The result type might be a single field for simple
	// returns, a slice of interfaces for anonymous returns and a struct for named
	// returns.
	func (_IERC20 *IERC20Raw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
		return _IERC20.Contract.IERC20Caller.contract.Call(opts, result, method, params...)
	}

	// Transfer initiates a plain transaction to move funds to the contract, calling
	// its default method if one is available.
	func (_IERC20 *IERC20Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
		return _IERC20.Contract.IERC20Transactor.contract.Transfer(opts)
	}

	// Transact invokes the (paid) contract method with params as input values.
	func (_IERC20 *IERC20Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
		return _IERC20.Contract.IERC20Transactor.contract.Transact(opts, method, params...)
	}

	// Call invokes the (constant) contract method with params as input values and
	// sets the output to result. The result type might be a single field for simple
	// returns, a slice of interfaces for anonymous returns and a struct for named
	// returns.
	func (_IERC20 *IERC20CallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
		return _IERC20.Contract.contract.Call(opts, result, method, params...)
	}

	// Transfer initiates a plain transaction to move funds to the contract, calling
	// its default method if one is available.
	func (_IERC20 *IERC20TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
		return _IERC20.Contract.contract.Transfer(opts)
	}

	// Transact invokes the (paid) contract method with params as input values.
	func (_IERC20 *IERC20TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
		return _IERC20.Contract.contract.Transact(opts, method, params...)
	}

	

	
		// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
		//
		// Solidity: function allowance(address owner, address spender) constant returns(uint256)
		func (_IERC20 *IERC20Caller) Allowance(opts *bind.CallOpts , owner common.Address , spender common.Address ) (*big.Int, error) {
			var (
				ret0 = new(*big.Int)
				
			)
			out := ret0
			err := _IERC20.contract.Call(opts, out, "allowance" , owner, spender)
			return *ret0, err
		}

		// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
		//
		// Solidity: function allowance(address owner, address spender) constant returns(uint256)
		func (_IERC20 *IERC20Session) Allowance( owner common.Address , spender common.Address ) ( *big.Int,  error) {
		  return _IERC20.Contract.Allowance(&_IERC20.CallOpts , owner, spender)
		}

		// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
		//
		// Solidity: function allowance(address owner, address spender) constant returns(uint256)
		func (_IERC20 *IERC20CallerSession) Allowance( owner common.Address , spender common.Address ) ( *big.Int,  error) {
		  return _IERC20.Contract.Allowance(&_IERC20.CallOpts , owner, spender)
		}
	
		// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
		//
		// Solidity: function balanceOf(address account) constant returns(uint256)
		func (_IERC20 *IERC20Caller) BalanceOf(opts *bind.CallOpts , account common.Address ) (*big.Int, error) {
			var (
				ret0 = new(*big.Int)
				
			)
			out := ret0
			err := _IERC20.contract.Call(opts, out, "balanceOf" , account)
			return *ret0, err
		}

		// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
		//
		// Solidity: function balanceOf(address account) constant returns(uint256)
		func (_IERC20 *IERC20Session) BalanceOf( account common.Address ) ( *big.Int,  error) {
		  return _IERC20.Contract.BalanceOf(&_IERC20.CallOpts , account)
		}

		// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
		//
		// Solidity: function balanceOf(address account) constant returns(uint256)
		func (_IERC20 *IERC20CallerSession) BalanceOf( account common.Address ) ( *big.Int,  error) {
		  return _IERC20.Contract.BalanceOf(&_IERC20.CallOpts , account)
		}
	
		// Decimals is a free data retrieval call binding the contract method 0x313ce567.
		//
		// Solidity: function decimals() constant returns(uint8 decimals)
		func (_IERC20 *IERC20Caller) Decimals(opts *bind.CallOpts ) (uint8, error) {
			var (
				ret0 = new(uint8)
				
			)
			out := ret0
			err := _IERC20.contract.Call(opts, out, "decimals" )
			return *ret0, err
		}

		// Decimals is a free data retrieval call binding the contract method 0x313ce567.
		//
		// Solidity: function decimals() constant returns(uint8 decimals)
		func (_IERC20 *IERC20Session) Decimals() ( uint8,  error) {
		  return _IERC20.Contract.Decimals(&_IERC20.CallOpts )
		}

		// Decimals is a free data retrieval call binding the contract method 0x313ce567.
		//
		// Solidity: function decimals() constant returns(uint8 decimals)
		func (_IERC20 *IERC20CallerSession) Decimals() ( uint8,  error) {
		  return _IERC20.Contract.Decimals(&_IERC20.CallOpts )
		}
	
		// Name is a free data retrieval call binding the contract method 0x06fdde03.
		//
		// Solidity: function name() constant returns(string)
		func (_IERC20 *IERC20Caller) Name(opts *bind.CallOpts ) (string, error) {
			var (
				ret0 = new(string)
				
			)
			out := ret0
			err := _IERC20.contract.Call(opts, out, "name" )
			return *ret0, err
		}

		// Name is a free data retrieval call binding the contract method 0x06fdde03.
		//
		// Solidity: function name() constant returns(string)
		func (_IERC20 *IERC20Session) Name() ( string,  error) {
		  return _IERC20.Contract.Name(&_IERC20.CallOpts )
		}

		// Name is a free data retrieval call binding the contract method 0x06fdde03.
		//
		// Solidity: function name() constant returns(string)
		func (_IERC20 *IERC20CallerSession) Name() ( string,  error) {
		  return _IERC20.Contract.Name(&_IERC20.CallOpts )
		}
	
		// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
		//
		// Solidity: function symbol() constant returns(string)
		func (_IERC20 *IERC20Caller) Symbol(opts *bind.CallOpts ) (string, error) {
			var (
				ret0 = new(string)
				
			)
			out := ret0
			err := _IERC20.contract.Call(opts, out, "symbol" )
			return *ret0, err
		}

		// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
		//
		// Solidity: function symbol() constant returns(string)
		func (_IERC20 *IERC20Session) Symbol() ( string,  error) {
		  return _IERC20.Contract.Symbol(&_IERC20.CallOpts )
		}

		// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
		//
		// Solidity: function symbol() constant returns(string)
		func (_IERC20 *IERC20CallerSession) Symbol() ( string,  error) {
		  return _IERC20.Contract.Symbol(&_IERC20.CallOpts )
		}
	
		// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
		//
		// Solidity: function totalSupply() constant returns(uint256)
		func (_IERC20 *IERC20Caller) TotalSupply(opts *bind.CallOpts ) (*big.Int, error) {
			var (
				ret0 = new(*big.Int)
				
			)
			out := ret0
			err := _IERC20.contract.Call(opts, out, "totalSupply" )
			return *ret0, err
		}

		// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
		//
		// Solidity: function totalSupply() constant returns(uint256)
		func (_IERC20 *IERC20Session) TotalSupply() ( *big.Int,  error) {
		  return _IERC20.Contract.TotalSupply(&_IERC20.CallOpts )
		}

		// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
		//
		// Solidity: function totalSupply() constant returns(uint256)
		func (_IERC20 *IERC20CallerSession) TotalSupply() ( *big.Int,  error) {
		  return _IERC20.Contract.TotalSupply(&_IERC20.CallOpts )
		}
	

	
		// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
		//
		// Solidity: function approve(address spender, uint256 amount) returns(bool)
		func (_IERC20 *IERC20Transactor) Approve(opts *bind.TransactOpts , spender common.Address , amount *big.Int ) (*types.Transaction, error) {
			return _IERC20.contract.Transact(opts, "approve" , spender, amount)
		}

		// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
		//
		// Solidity: function approve(address spender, uint256 amount) returns(bool)
		func (_IERC20 *IERC20Session) Approve( spender common.Address , amount *big.Int ) (*types.Transaction, error) {
		  return _IERC20.Contract.Approve(&_IERC20.TransactOpts , spender, amount)
		}

		// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
		//
		// Solidity: function approve(address spender, uint256 amount) returns(bool)
		func (_IERC20 *IERC20TransactorSession) Approve( spender common.Address , amount *big.Int ) (*types.Transaction, error) {
		  return _IERC20.Contract.Approve(&_IERC20.TransactOpts , spender, amount)
		}
	
		// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
		//
		// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
		func (_IERC20 *IERC20Transactor) Transfer(opts *bind.TransactOpts , recipient common.Address , amount *big.Int ) (*types.Transaction, error) {
			return _IERC20.contract.Transact(opts, "transfer" , recipient, amount)
		}

		// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
		//
		// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
		func (_IERC20 *IERC20Session) Transfer( recipient common.Address , amount *big.Int ) (*types.Transaction, error) {
		  return _IERC20.Contract.Transfer(&_IERC20.TransactOpts , recipient, amount)
		}

		// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
		//
		// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
		func (_IERC20 *IERC20TransactorSession) Transfer( recipient common.Address , amount *big.Int ) (*types.Transaction, error) {
		  return _IERC20.Contract.Transfer(&_IERC20.TransactOpts , recipient, amount)
		}
	
		// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
		//
		// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
		func (_IERC20 *IERC20Transactor) TransferFrom(opts *bind.TransactOpts , sender common.Address , recipient common.Address , amount *big.Int ) (*types.Transaction, error) {
			return _IERC20.contract.Transact(opts, "transferFrom" , sender, recipient, amount)
		}

		// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
		//
		// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
		func (_IERC20 *IERC20Session) TransferFrom( sender common.Address , recipient common.Address , amount *big.Int ) (*types.Transaction, error) {
		  return _IERC20.Contract.TransferFrom(&_IERC20.TransactOpts , sender, recipient, amount)
		}

		// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
		//
		// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
		func (_IERC20 *IERC20TransactorSession) TransferFrom( sender common.Address , recipient common.Address , amount *big.Int ) (*types.Transaction, error) {
		  return _IERC20.Contract.TransferFrom(&_IERC20.TransactOpts , sender, recipient, amount)
		}
	

	
		// IERC20ApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the IERC20 contract.
		type IERC20ApprovalIterator struct {
			Event *IERC20Approval // Event containing the contract specifics and raw log

			contract *bind.BoundContract // Generic contract to use for unpacking event data
			event    string              // Event name to use for unpacking event data

			logs chan types.Log        // Log channel receiving the found contract events
			sub  ethereum.Subscription // Subscription for errors, completion and termination
			done bool                  // Whether the subscription completed delivering logs
			fail error                 // Occurred error to stop iteration
		}
		// Next advances the iterator to the subsequent event, returning whether there
		// are any more events found. In case of a retrieval or parsing error, false is
		// returned and Error() can be queried for the exact failure.
		func (it *IERC20ApprovalIterator) Next() bool {
			// If the iterator failed, stop iterating
			if (it.fail != nil) {
				return false
			}
			// If the iterator completed, deliver directly whatever's available
			if (it.done) {
				select {
				case log := <-it.logs:
					it.Event = new(IERC20Approval)
					if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
						it.fail = err
						return false
					}
					it.Event.Raw = log
					return true

				default:
					return false
				}
			}
			// Iterator still in progress, wait for either a data or an error event
			select {
			case log := <-it.logs:
				it.Event = new(IERC20Approval)
				if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
					it.fail = err
					return false
				}
				it.Event.Raw = log
				return true

			case err := <-it.sub.Err():
				it.done = true
				it.fail = err
				return it.Next()
			}
		}
		// Error returns any retrieval or parsing error occurred during filtering.
		func (it *IERC20ApprovalIterator) Error() error {
			return it.fail
		}
		// Close terminates the iteration process, releasing any pending underlying
		// resources.
		func (it *IERC20ApprovalIterator) Close() error {
			it.sub.Unsubscribe()
			return nil
		}

		// IERC20Approval represents a Approval event raised by the IERC20 contract.
		type IERC20Approval struct { 
			Owner common.Address; 
			Spender common.Address; 
			Value *big.Int; 
			Raw types.Log // Blockchain specific contextual infos
		}

		// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
		//
		// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
 		func (_IERC20 *IERC20Filterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*IERC20ApprovalIterator, error) {
			
			var ownerRule []interface{}
			for _, ownerItem := range owner {
				ownerRule = append(ownerRule, ownerItem)
			}
			var spenderRule []interface{}
			for _, spenderItem := range spender {
				spenderRule = append(spenderRule, spenderItem)
			}
			

			logs, sub, err := _IERC20.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
			if err != nil {
				return nil, err
			}
			return &IERC20ApprovalIterator{contract: _IERC20.contract, event: "Approval", logs: logs, sub: sub}, nil
 		}

		// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
		//
		// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
		func (_IERC20 *IERC20Filterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *IERC20Approval, owner []common.Address, spender []common.Address) (event.Subscription, error) {
			
			var ownerRule []interface{}
			for _, ownerItem := range owner {
				ownerRule = append(ownerRule, ownerItem)
			}
			var spenderRule []interface{}
			for _, spenderItem := range spender {
				spenderRule = append(spenderRule, spenderItem)
			}
			

			logs, sub, err := _IERC20.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
			if err != nil {
				return nil, err
			}
			return event.NewSubscription(func(quit <-chan struct{}) error {
				defer sub.Unsubscribe()
				for {
					select {
					case log := <-logs:
						// New log arrived, parse the event and forward to the user
						event := new(IERC20Approval)
						if err := _IERC20.contract.UnpackLog(event, "Approval", log); err != nil {
							return err
						}
						event.Raw = log

						select {
						case sink <- event:
						case err := <-sub.Err():
							return err
						case <-quit:
							return nil
						}
					case err := <-sub.Err():
						return err
					case <-quit:
						return nil
					}
				}
			}), nil
		}

		// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
		//
		// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
		func (_IERC20 *IERC20Filterer) ParseApproval(log types.Log) (*IERC20Approval, error) {
			event := new(IERC20Approval)
			if err := _IERC20.contract.UnpackLog(event, "Approval", log); err != nil {
				return nil, err
			}
			return event, nil
		}

 	
		// IERC20TransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the IERC20 contract.
		type IERC20TransferIterator struct {
			Event *IERC20Transfer // Event containing the contract specifics and raw log

			contract *bind.BoundContract // Generic contract to use for unpacking event data
			event    string              // Event name to use for unpacking event data

			logs chan types.Log        // Log channel receiving the found contract events
			sub  ethereum.Subscription // Subscription for errors, completion and termination
			done bool                  // Whether the subscription completed delivering logs
			fail error                 // Occurred error to stop iteration
		}
		// Next advances the iterator to the subsequent event, returning whether there
		// are any more events found. In case of a retrieval or parsing error, false is
		// returned and Error() can be queried for the exact failure.
		func (it *IERC20TransferIterator) Next() bool {
			// If the iterator failed, stop iterating
			if (it.fail != nil) {
				return false
			}
			// If the iterator completed, deliver directly whatever's available
			if (it.done) {
				select {
				case log := <-it.logs:
					it.Event = new(IERC20Transfer)
					if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
						it.fail = err
						return false
					}
					it.Event.Raw = log
					return true

				default:
					return false
				}
			}
			// Iterator still in progress, wait for either a data or an error event
			select {
			case log := <-it.logs:
				it.Event = new(IERC20Transfer)
				if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
					it.fail = err
					return false
				}
				it.Event.Raw = log
				return true

			case err := <-it.sub.Err():
				it.done = true
				it.fail = err
				return it.Next()
			}
		}
		// Error returns any retrieval or parsing error occurred during filtering.
		func (it *IERC20TransferIterator) Error() error {
			return it.fail
		}
		// Close terminates the iteration process, releasing any pending underlying
		// resources.
		func (it *IERC20TransferIterator) Close() error {
			it.sub.Unsubscribe()
			return nil
		}

		// IERC20Transfer represents a Transfer event raised by the IERC20 contract.
		type IERC20Transfer struct { 
			From common.Address; 
			To common.Address; 
			Value *big.Int; 
			Raw types.Log // Blockchain specific contextual infos
		}

		// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
		//
		// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
 		func (_IERC20 *IERC20Filterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*IERC20TransferIterator, error) {
			
			var fromRule []interface{}
			for _, fromItem := range from {
				fromRule = append(fromRule, fromItem)
			}
			var toRule []interface{}
			for _, toItem := range to {
				toRule = append(toRule, toItem)
			}
			

			logs, sub, err := _IERC20.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
			if err != nil {
				return nil, err
			}
			return &IERC20TransferIterator{contract: _IERC20.contract, event: "Transfer", logs: logs, sub: sub}, nil
 		}

		// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
		//
		// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
		func (_IERC20 *IERC20Filterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *IERC20Transfer, from []common.Address, to []common.Address) (event.Subscription, error) {
			
			var fromRule []interface{}
			for _, fromItem := range from {
				fromRule = append(fromRule, fromItem)
			}
			var toRule []interface{}
			for _, toItem := range to {
				toRule = append(toRule, toItem)
			}
			

			logs, sub, err := _IERC20.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
			if err != nil {
				return nil, err
			}
			return event.NewSubscription(func(quit <-chan struct{}) error {
				defer sub.Unsubscribe()
				for {
					select {
					case log := <-logs:
						// New log arrived, parse the event and forward to the user
						event := new(IERC20Transfer)
						if err := _IERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
							return err
						}
						event.Raw = log

						select {
						case sink <- event:
						case err := <-sub.Err():
							return err
						case <-quit:
							return nil
						}
					case err := <-sub.Err():
						return err
					case <-quit:
						return nil
					}
				}
			}), nil
		}

		// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
		//
		// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
		func (_IERC20 *IERC20Filterer) ParseTransfer(log types.Log) (*IERC20Transfer, error) {
			event := new(IERC20Transfer)
			if err := _IERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
				return nil, err
			}
			return event, nil
		}

 	

	
	// SafeMathABI is the input ABI used to generate the binding from.
	const SafeMathABI = "[]"

	

	
		// SafeMathBin is the compiled bytecode used for deploying new contracts.
		var SafeMathBin = "0x604c602c600b82828239805160001a60731460008114601c57601e565bfe5b5030600052607381538281f30073000000000000000000000000000000000000000030146080604052600080fd00a165627a7a7230582058afa72077c199fa2c77cdcceb0ecc972fc6d1e9046df5a7f8d465dd48112eaa0029"

		// DeploySafeMath deploys a new Ethereum contract, binding an instance of SafeMath to it.
		func DeploySafeMath(auth *bind.TransactOpts, backend bind.ContractBackend ) (common.Address, *types.Transaction, *SafeMath, error) {
		  parsed, err := abi.JSON(strings.NewReader(SafeMathABI))
		  if err != nil {
		    return common.Address{}, nil, nil, err
		  }
		  
		  address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SafeMathBin), backend )
		  if err != nil {
		    return common.Address{}, nil, nil, err
		  }
		  return address, tx, &SafeMath{ SafeMathCaller: SafeMathCaller{contract: contract}, SafeMathTransactor: SafeMathTransactor{contract: contract}, SafeMathFilterer: SafeMathFilterer{contract: contract} }, nil
		}
	

	// SafeMath is an auto generated Go binding around an Ethereum contract.
	type SafeMath struct {
	  SafeMathCaller     // Read-only binding to the contract
	  SafeMathTransactor // Write-only binding to the contract
	  SafeMathFilterer   // Log filterer for contract events
	}

	// SafeMathCaller is an auto generated read-only Go binding around an Ethereum contract.
	type SafeMathCaller struct {
	  contract *bind.BoundContract // Generic contract wrapper for the low level calls
	}

	// SafeMathTransactor is an auto generated write-only Go binding around an Ethereum contract.
	type SafeMathTransactor struct {
	  contract *bind.BoundContract // Generic contract wrapper for the low level calls
	}

	// SafeMathFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
	type SafeMathFilterer struct {
	  contract *bind.BoundContract // Generic contract wrapper for the low level calls
	}

	// SafeMathSession is an auto generated Go binding around an Ethereum contract,
	// with pre-set call and transact options.
	type SafeMathSession struct {
	  Contract     *SafeMath        // Generic contract binding to set the session for
	  CallOpts     bind.CallOpts     // Call options to use throughout this session
	  TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
	}

	// SafeMathCallerSession is an auto generated read-only Go binding around an Ethereum contract,
	// with pre-set call options.
	type SafeMathCallerSession struct {
	  Contract *SafeMathCaller // Generic contract caller binding to set the session for
	  CallOpts bind.CallOpts    // Call options to use throughout this session
	}

	// SafeMathTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
	// with pre-set transact options.
	type SafeMathTransactorSession struct {
	  Contract     *SafeMathTransactor // Generic contract transactor binding to set the session for
	  TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
	}

	// SafeMathRaw is an auto generated low-level Go binding around an Ethereum contract.
	type SafeMathRaw struct {
	  Contract *SafeMath // Generic contract binding to access the raw methods on
	}

	// SafeMathCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
	type SafeMathCallerRaw struct {
		Contract *SafeMathCaller // Generic read-only contract binding to access the raw methods on
	}

	// SafeMathTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
	type SafeMathTransactorRaw struct {
		Contract *SafeMathTransactor // Generic write-only contract binding to access the raw methods on
	}

	// NewSafeMath creates a new instance of SafeMath, bound to a specific deployed contract.
	func NewSafeMath(address common.Address, backend bind.ContractBackend) (*SafeMath, error) {
	  contract, err := bindSafeMath(address, backend, backend, backend)
	  if err != nil {
	    return nil, err
	  }
	  return &SafeMath{ SafeMathCaller: SafeMathCaller{contract: contract}, SafeMathTransactor: SafeMathTransactor{contract: contract}, SafeMathFilterer: SafeMathFilterer{contract: contract} }, nil
	}

	// NewSafeMathCaller creates a new read-only instance of SafeMath, bound to a specific deployed contract.
	func NewSafeMathCaller(address common.Address, caller bind.ContractCaller) (*SafeMathCaller, error) {
	  contract, err := bindSafeMath(address, caller, nil, nil)
	  if err != nil {
	    return nil, err
	  }
	  return &SafeMathCaller{contract: contract}, nil
	}

	// NewSafeMathTransactor creates a new write-only instance of SafeMath, bound to a specific deployed contract.
	func NewSafeMathTransactor(address common.Address, transactor bind.ContractTransactor) (*SafeMathTransactor, error) {
	  contract, err := bindSafeMath(address, nil, transactor, nil)
	  if err != nil {
	    return nil, err
	  }
	  return &SafeMathTransactor{contract: contract}, nil
	}

	// NewSafeMathFilterer creates a new log filterer instance of SafeMath, bound to a specific deployed contract.
 	func NewSafeMathFilterer(address common.Address, filterer bind.ContractFilterer) (*SafeMathFilterer, error) {
 	  contract, err := bindSafeMath(address, nil, nil, filterer)
 	  if err != nil {
 	    return nil, err
 	  }
 	  return &SafeMathFilterer{contract: contract}, nil
 	}

	// bindSafeMath binds a generic wrapper to an already deployed contract.
	func bindSafeMath(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	  parsed, err := abi.JSON(strings.NewReader(SafeMathABI))
	  if err != nil {
	    return nil, err
	  }
	  return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
	}

	// Call invokes the (constant) contract method with params as input values and
	// sets the output to result. The result type might be a single field for simple
	// returns, a slice of interfaces for anonymous returns and a struct for named
	// returns.
	func (_SafeMath *SafeMathRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
		return _SafeMath.Contract.SafeMathCaller.contract.Call(opts, result, method, params...)
	}

	// Transfer initiates a plain transaction to move funds to the contract, calling
	// its default method if one is available.
	func (_SafeMath *SafeMathRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
		return _SafeMath.Contract.SafeMathTransactor.contract.Transfer(opts)
	}

	// Transact invokes the (paid) contract method with params as input values.
	func (_SafeMath *SafeMathRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
		return _SafeMath.Contract.SafeMathTransactor.contract.Transact(opts, method, params...)
	}

	// Call invokes the (constant) contract method with params as input values and
	// sets the output to result. The result type might be a single field for simple
	// returns, a slice of interfaces for anonymous returns and a struct for named
	// returns.
	func (_SafeMath *SafeMathCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
		return _SafeMath.Contract.contract.Call(opts, result, method, params...)
	}

	// Transfer initiates a plain transaction to move funds to the contract, calling
	// its default method if one is available.
	func (_SafeMath *SafeMathTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
		return _SafeMath.Contract.contract.Transfer(opts)
	}

	// Transact invokes the (paid) contract method with params as input values.
	func (_SafeMath *SafeMathTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
		return _SafeMath.Contract.contract.Transact(opts, method, params...)
	}


