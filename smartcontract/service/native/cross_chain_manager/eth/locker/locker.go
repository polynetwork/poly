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
	const EthereumCrossChainABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"owners\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_decimal\",\"type\":\"uint256\"},{\"name\":\"_toChainId\",\"type\":\"string\"},{\"name\":\"_toAddress\",\"type\":\"string\"}],\"name\":\"ERC20CrossChain\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_txId\",\"type\":\"string\"},{\"name\":\"_v\",\"type\":\"uint8[]\"},{\"name\":\"_r\",\"type\":\"bytes32[]\"},{\"name\":\"_s\",\"type\":\"bytes32[]\"}],\"name\":\"WithDrawERC20\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"msg\",\"type\":\"string\"},{\"name\":\"v\",\"type\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"VerifyWeb3ETHSigner\",\"outputs\":[{\"name\":\"signer\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_toChainId\",\"type\":\"string\"},{\"name\":\"_toAddress\",\"type\":\"string\"}],\"name\":\"ETHCrossChain\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"transactionId\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"sha3Hash\",\"type\":\"bytes32\"},{\"name\":\"v\",\"type\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"_verifySigner\",\"outputs\":[{\"name\":\"signer\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"transactions\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_txId\",\"type\":\"string\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"},{\"name\":\"_v\",\"type\":\"uint8[]\"},{\"name\":\"_r\",\"type\":\"bytes32[]\"},{\"name\":\"_s\",\"type\":\"bytes32[]\"}],\"name\":\"WithDrawETH\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_txId\",\"type\":\"string\"},{\"name\":\"sha3Hash\",\"type\":\"bytes32\"},{\"name\":\"_v\",\"type\":\"uint8[]\"},{\"name\":\"_r\",\"type\":\"bytes32[]\"},{\"name\":\"_s\",\"type\":\"bytes32[]\"}],\"name\":\"_verifyMultiSigner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"required\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_owners\",\"type\":\"address[]\"},{\"name\":\"_required\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"topic\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"Debug\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"DebugN\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"\",\"type\":\"string\"}],\"name\":\"Debugs\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"DebugByte\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"topic\",\"type\":\"string\"},{\"indexed\":true,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"txId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"decimal\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"toChainId\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"toAddress\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"rawdata\",\"type\":\"string\"}],\"name\":\"ERC20CrossChainEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"topic\",\"type\":\"string\"},{\"indexed\":true,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"txId\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"toChainId\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"toAddress\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"rawdata\",\"type\":\"string\"}],\"name\":\"ETHCrossChainEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"topic\",\"type\":\"string\"},{\"indexed\":true,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"toAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"WithDrawETHEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"topic\",\"type\":\"string\"},{\"indexed\":true,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"toAddress\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"WithDrawERC20Event\",\"type\":\"event\"}]"

	
		// EthereumCrossChainFuncSigs maps the 4-byte function signature to its string representation.
		var EthereumCrossChainFuncSigs = map[string]string{
			"0af8890c": "ERC20CrossChain(address,uint256,uint256,string,string)",
			"690ad5b7": "ETHCrossChain(string,string)",
			"44fcddb9": "VerifyWeb3ETHSigner(string,uint8,bytes32,bytes32)",
			"1cc98f77": "WithDrawERC20(address,address,uint256,string,uint8[],bytes32[],bytes32[])",
			"b03450fe": "WithDrawETH(string,address,uint256,uint8[],bytes32[],bytes32[])",
			"c281eee6": "_verifyMultiSigner(string,bytes32,uint8[],bytes32[],bytes32[])",
			"94187159": "_verifySigner(bytes32,uint8,bytes32,bytes32)",
			"025e7c27": "owners(uint256)",
			"dc8452cd": "required()",
			"7e2f42e7": "transactionId()",
			"9ace38c2": "transactions(uint256)",
			
		}
	

	
		// EthereumCrossChainBin is the compiled bytecode used for deploying new contracts.
		var EthereumCrossChainBin = "0x60806040523480156200001157600080fd5b5060405162001eaf38038062001eaf83398101604052805160208201519101805190919060009081908311156200004757600080fd5b8351603210156200005757600080fd5b600091505b8351821015620000ef5783828151811015156200007557fe5b602090810290910101519050600160a060020a03811615156200009757600080fd5b600160a060020a03811660009081526003602052604090205460ff1615620000be57600080fd5b600160a060020a0381166000908152600360205260409020805460ff1916600190811790915591909101906200005c565b83516200010490600290602087019062000111565b50505060015550620001a5565b82805482825590600052602060002090810192821562000169579160200282015b82811115620001695782518254600160a060020a031916600160a060020a0390911617825560209092019160019091019062000132565b50620001779291506200017b565b5090565b620001a291905b8082111562000177578054600160a060020a031916815560010162000182565b90565b611cfa80620001b56000396000f3006080604052600436106100955763ffffffff60e060020a600035041663025e7c27811461009a5780630af8890c146100ce5780631cc98f771461017a57806344fcddb914610294578063690ad5b7146102fc5780637e2f42e71461038657806394187159146103ad5780639ace38c2146103d1578063b03450fe146103e9578063c281eee614610505578063dc8452cd14610624575b600080fd5b3480156100a657600080fd5b506100b2600435610639565b60408051600160a060020a039092168252519081900360200190f35b3480156100da57600080fd5b50604080516020601f60643560048181013592830184900484028501840190955281845261017894600160a060020a03813516946024803595604435953695608494930191819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a9998810197919650918201945092508291508401838280828437509497506106619650505050505050565b005b34801561018657600080fd5b50604080516020601f60643560048181013592830184900484028501840190955281845261017894600160a060020a038135811695602480359092169560443595369560849401918190840183828082843750506040805187358901803560208181028481018201909552818452989b9a998901989297509082019550935083925085019084908082843750506040805187358901803560208181028481018201909552818452989b9a998901989297509082019550935083925085019084908082843750506040805187358901803560208181028481018201909552818452989b9a998901989297509082019550935083925085019084908082843750949750610a569650505050505050565b3480156102a057600080fd5b506040805160206004803580820135601f81018490048402850184019095528484526100b29436949293602493928401919081908401838280828437509497505050833560ff1694505050602082013591604001359050610c7b565b6040805160206004803580820135601f810184900484028501840190955284845261017894369492936024939284019190819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a999881019791965091820194509250829150840183828082843750949750610f0c9650505050505050565b34801561039257600080fd5b5061039b6111a3565b60408051918252519081900360200190f35b3480156103b957600080fd5b506100b260043560ff602435166044356064356111a9565b3480156103dd57600080fd5b5061039b60043561130a565b3480156103f557600080fd5b506040805160206004803580820135601f81018490048402850184019095528484526101789436949293602493928401919081908401838280828437505060408051818801358901803560208181028481018201909552818452989b600160a060020a038b35169b8a8c01359b919a90995060609091019750929550908201935091829185019084908082843750506040805187358901803560208181028481018201909552818452989b9a998901989297509082019550935083925085019084908082843750506040805187358901803560208181028481018201909552818452989b9a99890198929750908201955093508392508501908490808284375094975061131c9650505050505050565b34801561051157600080fd5b506040805160206004803580820135601f810184900484028501840190955284845261061094369492936024939284019190819084018382808284375050604080516020808901358a01803580830284810184018652818552999c8b359c909b909a950198509296508101945090925082919085019084908082843750506040805187358901803560208181028481018201909552818452989b9a998901989297509082019550935083925085019084908082843750506040805187358901803560208181028481018201909552818452989b9a9989019892975090820195509350839250850190849080828437509497506114d49650505050505050565b604080519115158252519081900360200190f35b34801561063057600080fd5b5061039b611715565b600280548290811061064757fe5b600091825260209091200154600160a060020a0316905081565b600061066b611c85565b60606000871161067a57600080fd5b604080517f23b872dd000000000000000000000000000000000000000000000000000000008152336004820152306024820152604481018990529051600160a060020a038a16916323b872dd9160648083019260209291908290030181600087803b1580156106e857600080fd5b505af11580156106fc573d6000803e3d6000fd5b505050506040513d602081101561071257600080fd5b5051151561071f57600080fd5b87600160a060020a031663313ce5676040518163ffffffff1660e060020a028152600401602060405180830381600087803b15801561075d57600080fd5b505af1158015610771573d6000803e3d6000fd5b505050506040513d602081101561078757600080fd5b50519250600060ff84161161079b57600080fd5b6040805160c081018252600160a060020a038a168152336020820152908101869052606081018590526080810188905260ff841660a082015291506107df8261171b565b9050806040518082805190602001908083835b602083106108115780518252601f1990920191602091820191016107f2565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051809103902060008060045481526020019081526020016000208160001916905550600160046000828254019250508190555033600160a060020a03167f363ffaf80f7a934d1d3bf7fc142b4e2925afd02adc95e00fe7693a74a24baf416001600454038a8a8a8a8a88604051808060200189815260200188600160a060020a0316600160a060020a031681526020018781526020018681526020018060200180602001806020018581038552600f8152602001807f455243323043726f7373436861696e0000000000000000000000000000000000815250602001858103845288818151815260200191508051906020019080838360005b83811015610949578181015183820152602001610931565b50505050905090810190601f1680156109765780820380516001836020036101000a031916815260200191505b50858103835287518152875160209182019189019080838360005b838110156109a9578181015183820152602001610991565b50505050905090810190601f1680156109d65780820380516001836020036101000a031916815260200191505b50858103825286518152865160209182019188019080838360005b83811015610a095781810151838201526020016109f1565b50505050905090810190601f168015610a365780820380516001836020036101000a031916815260200191505b509b50505050505050505050505060405180910390a25050505050505050565b6000600160a060020a0387161515610a6d57600080fd5b60008611610a7a57600080fd5b87600160a060020a031663a9059cbb88886040518363ffffffff1660e060020a0281526004018083600160a060020a0316600160a060020a0316815260200182815260200192505050602060405180830381600087803b158015610add57600080fd5b505af1158015610af1573d6000803e3d6000fd5b505050506040513d6020811015610b0757600080fd5b50511515610b1457600080fd5b846040516020018082805190602001908083835b60208310610b475780518252601f199092019160209182019101610b28565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040526040518082805190602001908083835b60208310610baa5780518252601f199092019160209182019101610b8b565b51815160001960209485036101000a0190811690199190911617905260408051949092018490038420600081815260068352839020805460ff19166001179055600160a060020a038f811692860192909252908d1684830152606084018c90526080808552600d908501527f576974684472617745524332300000000000000000000000000000000000000060a085015290519095503394507f98ef7f0d02c5ec83cf62fda844997bad56b1e0f06cf91cedeffeb58bd08df06d93509182900360c001919050a25050505050505050565b60008060606000876040516020018082805190602001908083835b60208310610cb55780518252601f199092019160209182019101610c96565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040526040518082805190602001908083835b60208310610d185780518252601f199092019160209182019101610cf9565b51815160209384036101000a6000190180199092169116179052604080519290940182900382208285018552601c8084527f19457468657265756d205369676e6564204d6573736167653a0a3332000000008484019081529551919a509298508896508995500192839250908083835b60208310610da75780518252601f199092019160209182019101610d88565b51815160209384036101000a600019018019909216911617905292019384525060408051808503815293820190819052835193945092839250908401908083835b60208310610e075780518252601f199092019160209182019101610de8565b51815160209384036101000a600019018019909216911617905260408051929094018290038220600080845283830180875282905260ff8f1684870152606084018e9052608084018d905294519097506001965060a080840196509194601f19820194509281900390910191865af1158015610e87573d6000803e3d6000fd5b505060408051601f19810151600160a060020a0381166020830152828252600b828401527f436865636b5369676e6572000000000000000000000000000000000000000000606083015291519196507f14186b8ac9c91f14b0f16f9e886356157442bb899be26513dfe1d4d5929a5bac925081900360800190a1505050949350505050565b610f14611c85565b606060003411610f2357600080fd5b6040805160c08101825260008082523360208301529181018690526060810185905234608082015260a08101919091529150610f5e8261171b565b9050806040518082805190602001908083835b60208310610f905780518252601f199092019160209182019101610f71565b518151602093840361010090810a60001901801990931692909116919091179091526040805193909501839003832060048054600090815280855287812092909255805460018101909155848401819052958401819052346060850181905260e0808652600d908601527f45544843726f7373436861696e0000000000000000000000000000000000000092850192909252610120608085018181528d51918601919091528c513399507fc8a840d188e5175163eacef4950c691222c7c591788bdee7451a88868334a315985091955091938c938c938b93839260a084019160c0850191610140860191908a01908083838f5b8381101561109b578181015183820152602001611083565b50505050905090810190601f1680156110c85780820380516001836020036101000a031916815260200191505b50858103835287518152875160209182019189019080838360005b838110156110fb5781810151838201526020016110e3565b50505050905090810190601f1680156111285780820380516001836020036101000a031916815260200191505b50858103825286518152865160209182019188019080838360005b8381101561115b578181015183820152602001611143565b50505050905090810190601f1680156111885780820380516001836020036101000a031916815260200191505b509a505050505050505050505060405180910390a250505050565b60045481565b604080518082018252601c8082527f19457468657265756d205369676e6564204d6573736167653a0a33320000000060208084019182529351600094859385938b939092019182918083835b602083106112145780518252601f1990920191602091820191016111f5565b51815160209384036101000a600019018019909216911617905292019384525060408051808503815293820190819052835193945092839250908401908083835b602083106112745780518252601f199092019160209182019101611255565b51815160209384036101000a600019018019909216911617905260408051929094018290038220600080845283830180875282905260ff8e1684870152606084018d9052608084018c905294519097506001965060a080840196509194601f19820194509281900390910191865af11580156112f4573d6000803e3d6000fd5b5050604051601f19015198975050505050505050565b60006020819052908152604090205481565b6000600160a060020a038616151561133357600080fd5b6000851161134057600080fd5b604051600160a060020a0387169086156108fc029087906000818181858888f19350505050158015611376573d6000803e3d6000fd5b50866040516020018082805190602001908083835b602083106113aa5780518252601f19909201916020918201910161138b565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040526040518082805190602001908083835b6020831061140d5780518252601f1990920191602091820191016113ee565b51815160001960209485036101000a0190811690199190911617905260408051949092018490038420600081815260068352839020805460ff19166001179055600160a060020a038d16918501919091528382018b90526060808552600b908501527f5769746844726177455448000000000000000000000000000000000000000000608085015290519095503394507f5614f9659cb7146440d26cd1cb85f8d39eae4467d95b092bd92391e81d50227993509182900360a001919050a250505050505050565b6000806000806000895160241415156114ec57600080fd5b896040516020018082805190602001908083835b6020831061151f5780518252601f199092019160209182019101611500565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040526040518082805190602001908083835b602083106115825780518252601f199092019160209182019101611563565b51815160209384036101000a6000190180199092169116179052604080519290940182900390912060008181526006909252929020549197505060ff161591506115cd905057600080fd5b865188511480156115df575085518751145b15156115ea57600080fd5b600154885110156115fa57600080fd5b60009250600091505b6001548210156116f65761165e89898481518110151561161f57fe5b90602001906020020151898581518110151561163757fe5b90602001906020020151898681518110151561164f57fe5b906020019060200201516111a9565b600160a060020a03811660009081526003602052604090205490915060ff16151561168857600080fd5b6000848152600560209081526040808320600160a060020a038516845290915290205460ff16156116b857600080fd5b6000848152600560209081526040808320600160a060020a03851684529091529020805460ff19166001908117909155928301929190910190611603565b60015483101561170557600080fd5b5060019998505050505050505050565b60015481565b606061172a8260000151611999565b6117378360200151611999565b8360400151846060015161174e8660800151611b9f565b61175b8760a00151611b9f565b6040516020018087805190602001908083835b6020831061178d5780518252601f19909201916020918201910161176e565b6001836020036101000a0380198251168184511680821785525050505050509050018060f860020a60230281525060010186805190602001908083835b602083106117e95780518252601f1990920191602091820191016117ca565b6001836020036101000a0380198251168184511680821785525050505050509050018060f860020a60230281525060010185805190602001908083835b602083106118455780518252601f199092019160209182019101611826565b6001836020036101000a0380198251168184511680821785525050505050509050018060f860020a60230281525060010184805190602001908083835b602083106118a15780518252601f199092019160209182019101611882565b6001836020036101000a0380198251168184511680821785525050505050509050018060f860020a60230281525060010183805190602001908083835b602083106118fd5780518252601f1990920191602091820191016118de565b6001836020036101000a0380198251168184511680821785525050505050509050018060f860020a60230281525060010182805190602001908083835b602083106119595780518252601f19909201916020918201910161193a565b6001836020036101000a03801982511681845116808217855250505050505090500196505050505050506040516020818303038152906040529050919050565b604080518082018252601081527f303132333435363738396162636465660000000000000000000000000000000060208201528151602a8082526060828101909452600160a060020a03851692918491600091908160200160208202803883390190505091507f3000000000000000000000000000000000000000000000000000000000000000826000815181101515611a2f57fe5b906020010190600160f860020a031916908160001a9053507f7800000000000000000000000000000000000000000000000000000000000000826001815181101515611a7757fe5b906020010190600160f860020a031916908160001a905350600090505b6014811015611b925782600485600c840160208110611aaf57fe5b1a60f860020a02600160f860020a0319169060020a900460f860020a9004815181101515611ad957fe5b90602001015160f860020a900460f860020a028282600202600201815181101515611b0057fe5b906020010190600160f860020a031916908160001a9053508284600c830160208110611b2857fe5b1a60f860020a02600f60f860020a021660f860020a9004815181101515611b4b57fe5b90602001015160f860020a900460f860020a028282600202600301815181101515611b7257fe5b906020010190600160f860020a031916908160001a905350600101611a94565b8194505b50505050919050565b60606000808281851515611be85760408051808201909152600181527f300000000000000000000000000000000000000000000000000000000000000060208201529450611b96565b8593505b8315611c0357600190920191600a84049350611bec565b826040519080825280601f01601f191660200182016040528015611c31578160200160208202803883390190505b5091505060001982015b8515611b9257815160001982019160f860020a6030600a8a060102918491908110611c6257fe5b906020010190600160f860020a031916908160001a905350600a86049550611c3b565b60c0604051908101604052806000600160a060020a031681526020016000600160a060020a031681526020016060815260200160608152602001600081526020016000815250905600a165627a7a72305820e6b372f4825c47580bf7a5638c499814b57ff9f27ec38b558c58696e5c723c520029"

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
	

	
		// ERC20CrossChain is a paid mutator transaction binding the contract method 0x0af8890c.
		//
		// Solidity: function ERC20CrossChain(address _token, uint256 _value, uint256 _decimal, string _toChainId, string _toAddress) returns()
		func (_EthereumCrossChain *EthereumCrossChainTransactor) ERC20CrossChain(opts *bind.TransactOpts , _token common.Address , _value *big.Int , _decimal *big.Int , _toChainId string , _toAddress string ) (*types.Transaction, error) {
			return _EthereumCrossChain.contract.Transact(opts, "ERC20CrossChain" , _token, _value, _decimal, _toChainId, _toAddress)
		}

		// ERC20CrossChain is a paid mutator transaction binding the contract method 0x0af8890c.
		//
		// Solidity: function ERC20CrossChain(address _token, uint256 _value, uint256 _decimal, string _toChainId, string _toAddress) returns()
		func (_EthereumCrossChain *EthereumCrossChainSession) ERC20CrossChain( _token common.Address , _value *big.Int , _decimal *big.Int , _toChainId string , _toAddress string ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.ERC20CrossChain(&_EthereumCrossChain.TransactOpts , _token, _value, _decimal, _toChainId, _toAddress)
		}

		// ERC20CrossChain is a paid mutator transaction binding the contract method 0x0af8890c.
		//
		// Solidity: function ERC20CrossChain(address _token, uint256 _value, uint256 _decimal, string _toChainId, string _toAddress) returns()
		func (_EthereumCrossChain *EthereumCrossChainTransactorSession) ERC20CrossChain( _token common.Address , _value *big.Int , _decimal *big.Int , _toChainId string , _toAddress string ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.ERC20CrossChain(&_EthereumCrossChain.TransactOpts , _token, _value, _decimal, _toChainId, _toAddress)
		}
	
		// ETHCrossChain is a paid mutator transaction binding the contract method 0x690ad5b7.
		//
		// Solidity: function ETHCrossChain(string _toChainId, string _toAddress) returns()
		func (_EthereumCrossChain *EthereumCrossChainTransactor) ETHCrossChain(opts *bind.TransactOpts , _toChainId string , _toAddress string ) (*types.Transaction, error) {
			return _EthereumCrossChain.contract.Transact(opts, "ETHCrossChain" , _toChainId, _toAddress)
		}

		// ETHCrossChain is a paid mutator transaction binding the contract method 0x690ad5b7.
		//
		// Solidity: function ETHCrossChain(string _toChainId, string _toAddress) returns()
		func (_EthereumCrossChain *EthereumCrossChainSession) ETHCrossChain( _toChainId string , _toAddress string ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.ETHCrossChain(&_EthereumCrossChain.TransactOpts , _toChainId, _toAddress)
		}

		// ETHCrossChain is a paid mutator transaction binding the contract method 0x690ad5b7.
		//
		// Solidity: function ETHCrossChain(string _toChainId, string _toAddress) returns()
		func (_EthereumCrossChain *EthereumCrossChainTransactorSession) ETHCrossChain( _toChainId string , _toAddress string ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.ETHCrossChain(&_EthereumCrossChain.TransactOpts , _toChainId, _toAddress)
		}
	
		// VerifyWeb3ETHSigner is a paid mutator transaction binding the contract method 0x44fcddb9.
		//
		// Solidity: function VerifyWeb3ETHSigner(string msg, uint8 v, bytes32 r, bytes32 s) returns(address signer)
		func (_EthereumCrossChain *EthereumCrossChainTransactor) VerifyWeb3ETHSigner(opts *bind.TransactOpts , msg string , v uint8 , r [32]byte , s [32]byte ) (*types.Transaction, error) {
			return _EthereumCrossChain.contract.Transact(opts, "VerifyWeb3ETHSigner" , msg, v, r, s)
		}

		// VerifyWeb3ETHSigner is a paid mutator transaction binding the contract method 0x44fcddb9.
		//
		// Solidity: function VerifyWeb3ETHSigner(string msg, uint8 v, bytes32 r, bytes32 s) returns(address signer)
		func (_EthereumCrossChain *EthereumCrossChainSession) VerifyWeb3ETHSigner( msg string , v uint8 , r [32]byte , s [32]byte ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.VerifyWeb3ETHSigner(&_EthereumCrossChain.TransactOpts , msg, v, r, s)
		}

		// VerifyWeb3ETHSigner is a paid mutator transaction binding the contract method 0x44fcddb9.
		//
		// Solidity: function VerifyWeb3ETHSigner(string msg, uint8 v, bytes32 r, bytes32 s) returns(address signer)
		func (_EthereumCrossChain *EthereumCrossChainTransactorSession) VerifyWeb3ETHSigner( msg string , v uint8 , r [32]byte , s [32]byte ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.VerifyWeb3ETHSigner(&_EthereumCrossChain.TransactOpts , msg, v, r, s)
		}
	
		// WithDrawERC20 is a paid mutator transaction binding the contract method 0x1cc98f77.
		//
		// Solidity: function WithDrawERC20(address _token, address _to, uint256 _value, string _txId, uint8[] _v, bytes32[] _r, bytes32[] _s) returns()
		func (_EthereumCrossChain *EthereumCrossChainTransactor) WithDrawERC20(opts *bind.TransactOpts , _token common.Address , _to common.Address , _value *big.Int , _txId string , _v []uint8 , _r [][32]byte , _s [][32]byte ) (*types.Transaction, error) {
			return _EthereumCrossChain.contract.Transact(opts, "WithDrawERC20" , _token, _to, _value, _txId, _v, _r, _s)
		}

		// WithDrawERC20 is a paid mutator transaction binding the contract method 0x1cc98f77.
		//
		// Solidity: function WithDrawERC20(address _token, address _to, uint256 _value, string _txId, uint8[] _v, bytes32[] _r, bytes32[] _s) returns()
		func (_EthereumCrossChain *EthereumCrossChainSession) WithDrawERC20( _token common.Address , _to common.Address , _value *big.Int , _txId string , _v []uint8 , _r [][32]byte , _s [][32]byte ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.WithDrawERC20(&_EthereumCrossChain.TransactOpts , _token, _to, _value, _txId, _v, _r, _s)
		}

		// WithDrawERC20 is a paid mutator transaction binding the contract method 0x1cc98f77.
		//
		// Solidity: function WithDrawERC20(address _token, address _to, uint256 _value, string _txId, uint8[] _v, bytes32[] _r, bytes32[] _s) returns()
		func (_EthereumCrossChain *EthereumCrossChainTransactorSession) WithDrawERC20( _token common.Address , _to common.Address , _value *big.Int , _txId string , _v []uint8 , _r [][32]byte , _s [][32]byte ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.WithDrawERC20(&_EthereumCrossChain.TransactOpts , _token, _to, _value, _txId, _v, _r, _s)
		}
	
		// WithDrawETH is a paid mutator transaction binding the contract method 0xb03450fe.
		//
		// Solidity: function WithDrawETH(string _txId, address _to, uint256 _value, uint8[] _v, bytes32[] _r, bytes32[] _s) returns()
		func (_EthereumCrossChain *EthereumCrossChainTransactor) WithDrawETH(opts *bind.TransactOpts , _txId string , _to common.Address , _value *big.Int , _v []uint8 , _r [][32]byte , _s [][32]byte ) (*types.Transaction, error) {
			return _EthereumCrossChain.contract.Transact(opts, "WithDrawETH" , _txId, _to, _value, _v, _r, _s)
		}

		// WithDrawETH is a paid mutator transaction binding the contract method 0xb03450fe.
		//
		// Solidity: function WithDrawETH(string _txId, address _to, uint256 _value, uint8[] _v, bytes32[] _r, bytes32[] _s) returns()
		func (_EthereumCrossChain *EthereumCrossChainSession) WithDrawETH( _txId string , _to common.Address , _value *big.Int , _v []uint8 , _r [][32]byte , _s [][32]byte ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.WithDrawETH(&_EthereumCrossChain.TransactOpts , _txId, _to, _value, _v, _r, _s)
		}

		// WithDrawETH is a paid mutator transaction binding the contract method 0xb03450fe.
		//
		// Solidity: function WithDrawETH(string _txId, address _to, uint256 _value, uint8[] _v, bytes32[] _r, bytes32[] _s) returns()
		func (_EthereumCrossChain *EthereumCrossChainTransactorSession) WithDrawETH( _txId string , _to common.Address , _value *big.Int , _v []uint8 , _r [][32]byte , _s [][32]byte ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.WithDrawETH(&_EthereumCrossChain.TransactOpts , _txId, _to, _value, _v, _r, _s)
		}
	
		// VerifyMultiSigner is a paid mutator transaction binding the contract method 0xc281eee6.
		//
		// Solidity: function _verifyMultiSigner(string _txId, bytes32 sha3Hash, uint8[] _v, bytes32[] _r, bytes32[] _s) returns(bool)
		func (_EthereumCrossChain *EthereumCrossChainTransactor) VerifyMultiSigner(opts *bind.TransactOpts , _txId string , sha3Hash [32]byte , _v []uint8 , _r [][32]byte , _s [][32]byte ) (*types.Transaction, error) {
			return _EthereumCrossChain.contract.Transact(opts, "_verifyMultiSigner" , _txId, sha3Hash, _v, _r, _s)
		}

		// VerifyMultiSigner is a paid mutator transaction binding the contract method 0xc281eee6.
		//
		// Solidity: function _verifyMultiSigner(string _txId, bytes32 sha3Hash, uint8[] _v, bytes32[] _r, bytes32[] _s) returns(bool)
		func (_EthereumCrossChain *EthereumCrossChainSession) VerifyMultiSigner( _txId string , sha3Hash [32]byte , _v []uint8 , _r [][32]byte , _s [][32]byte ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.VerifyMultiSigner(&_EthereumCrossChain.TransactOpts , _txId, sha3Hash, _v, _r, _s)
		}

		// VerifyMultiSigner is a paid mutator transaction binding the contract method 0xc281eee6.
		//
		// Solidity: function _verifyMultiSigner(string _txId, bytes32 sha3Hash, uint8[] _v, bytes32[] _r, bytes32[] _s) returns(bool)
		func (_EthereumCrossChain *EthereumCrossChainTransactorSession) VerifyMultiSigner( _txId string , sha3Hash [32]byte , _v []uint8 , _r [][32]byte , _s [][32]byte ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.VerifyMultiSigner(&_EthereumCrossChain.TransactOpts , _txId, sha3Hash, _v, _r, _s)
		}
	
		// VerifySigner is a paid mutator transaction binding the contract method 0x94187159.
		//
		// Solidity: function _verifySigner(bytes32 sha3Hash, uint8 v, bytes32 r, bytes32 s) returns(address signer)
		func (_EthereumCrossChain *EthereumCrossChainTransactor) VerifySigner(opts *bind.TransactOpts , sha3Hash [32]byte , v uint8 , r [32]byte , s [32]byte ) (*types.Transaction, error) {
			return _EthereumCrossChain.contract.Transact(opts, "_verifySigner" , sha3Hash, v, r, s)
		}

		// VerifySigner is a paid mutator transaction binding the contract method 0x94187159.
		//
		// Solidity: function _verifySigner(bytes32 sha3Hash, uint8 v, bytes32 r, bytes32 s) returns(address signer)
		func (_EthereumCrossChain *EthereumCrossChainSession) VerifySigner( sha3Hash [32]byte , v uint8 , r [32]byte , s [32]byte ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.VerifySigner(&_EthereumCrossChain.TransactOpts , sha3Hash, v, r, s)
		}

		// VerifySigner is a paid mutator transaction binding the contract method 0x94187159.
		//
		// Solidity: function _verifySigner(bytes32 sha3Hash, uint8 v, bytes32 r, bytes32 s) returns(address signer)
		func (_EthereumCrossChain *EthereumCrossChainTransactorSession) VerifySigner( sha3Hash [32]byte , v uint8 , r [32]byte , s [32]byte ) (*types.Transaction, error) {
		  return _EthereumCrossChain.Contract.VerifySigner(&_EthereumCrossChain.TransactOpts , sha3Hash, v, r, s)
		}
	

	
		// EthereumCrossChainDebugIterator is returned from FilterDebug and is used to iterate over the raw logs and unpacked data for Debug events raised by the EthereumCrossChain contract.
		type EthereumCrossChainDebugIterator struct {
			Event *EthereumCrossChainDebug // Event containing the contract specifics and raw log

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
		func (it *EthereumCrossChainDebugIterator) Next() bool {
			// If the iterator failed, stop iterating
			if (it.fail != nil) {
				return false
			}
			// If the iterator completed, deliver directly whatever's available
			if (it.done) {
				select {
				case log := <-it.logs:
					it.Event = new(EthereumCrossChainDebug)
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
				it.Event = new(EthereumCrossChainDebug)
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
		func (it *EthereumCrossChainDebugIterator) Error() error {
			return it.fail
		}
		// Close terminates the iteration process, releasing any pending underlying
		// resources.
		func (it *EthereumCrossChainDebugIterator) Close() error {
			it.sub.Unsubscribe()
			return nil
		}

		// EthereumCrossChainDebug represents a Debug event raised by the EthereumCrossChain contract.
		type EthereumCrossChainDebug struct { 
			Topic string; 
			Addr common.Address; 
			Raw types.Log // Blockchain specific contextual infos
		}

		// FilterDebug is a free log retrieval operation binding the contract event 0x14186b8ac9c91f14b0f16f9e886356157442bb899be26513dfe1d4d5929a5bac.
		//
		// Solidity: event Debug(string topic, address addr)
 		func (_EthereumCrossChain *EthereumCrossChainFilterer) FilterDebug(opts *bind.FilterOpts) (*EthereumCrossChainDebugIterator, error) {
			
			
			

			logs, sub, err := _EthereumCrossChain.contract.FilterLogs(opts, "Debug")
			if err != nil {
				return nil, err
			}
			return &EthereumCrossChainDebugIterator{contract: _EthereumCrossChain.contract, event: "Debug", logs: logs, sub: sub}, nil
 		}

		// WatchDebug is a free log subscription operation binding the contract event 0x14186b8ac9c91f14b0f16f9e886356157442bb899be26513dfe1d4d5929a5bac.
		//
		// Solidity: event Debug(string topic, address addr)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) WatchDebug(opts *bind.WatchOpts, sink chan<- *EthereumCrossChainDebug) (event.Subscription, error) {
			
			
			

			logs, sub, err := _EthereumCrossChain.contract.WatchLogs(opts, "Debug")
			if err != nil {
				return nil, err
			}
			return event.NewSubscription(func(quit <-chan struct{}) error {
				defer sub.Unsubscribe()
				for {
					select {
					case log := <-logs:
						// New log arrived, parse the event and forward to the user
						event := new(EthereumCrossChainDebug)
						if err := _EthereumCrossChain.contract.UnpackLog(event, "Debug", log); err != nil {
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

		// ParseDebug is a log parse operation binding the contract event 0x14186b8ac9c91f14b0f16f9e886356157442bb899be26513dfe1d4d5929a5bac.
		//
		// Solidity: event Debug(string topic, address addr)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) ParseDebug(log types.Log) (*EthereumCrossChainDebug, error) {
			event := new(EthereumCrossChainDebug)
			if err := _EthereumCrossChain.contract.UnpackLog(event, "Debug", log); err != nil {
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
			 []byte; 
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

 	
		// EthereumCrossChainDebugNIterator is returned from FilterDebugN and is used to iterate over the raw logs and unpacked data for DebugN events raised by the EthereumCrossChain contract.
		type EthereumCrossChainDebugNIterator struct {
			Event *EthereumCrossChainDebugN // Event containing the contract specifics and raw log

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
		func (it *EthereumCrossChainDebugNIterator) Next() bool {
			// If the iterator failed, stop iterating
			if (it.fail != nil) {
				return false
			}
			// If the iterator completed, deliver directly whatever's available
			if (it.done) {
				select {
				case log := <-it.logs:
					it.Event = new(EthereumCrossChainDebugN)
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
				it.Event = new(EthereumCrossChainDebugN)
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
		func (it *EthereumCrossChainDebugNIterator) Error() error {
			return it.fail
		}
		// Close terminates the iteration process, releasing any pending underlying
		// resources.
		func (it *EthereumCrossChainDebugNIterator) Close() error {
			it.sub.Unsubscribe()
			return nil
		}

		// EthereumCrossChainDebugN represents a DebugN event raised by the EthereumCrossChain contract.
		type EthereumCrossChainDebugN struct { 
			 *big.Int; 
			Raw types.Log // Blockchain specific contextual infos
		}

		// FilterDebugN is a free log retrieval operation binding the contract event 0xd4f190b7c2bbb19584eb23cae55f049509c7bf1c0fa046a9c93634b3983c953a.
		//
		// Solidity: event DebugN(uint256 )
 		func (_EthereumCrossChain *EthereumCrossChainFilterer) FilterDebugN(opts *bind.FilterOpts) (*EthereumCrossChainDebugNIterator, error) {
			
			

			logs, sub, err := _EthereumCrossChain.contract.FilterLogs(opts, "DebugN")
			if err != nil {
				return nil, err
			}
			return &EthereumCrossChainDebugNIterator{contract: _EthereumCrossChain.contract, event: "DebugN", logs: logs, sub: sub}, nil
 		}

		// WatchDebugN is a free log subscription operation binding the contract event 0xd4f190b7c2bbb19584eb23cae55f049509c7bf1c0fa046a9c93634b3983c953a.
		//
		// Solidity: event DebugN(uint256 )
		func (_EthereumCrossChain *EthereumCrossChainFilterer) WatchDebugN(opts *bind.WatchOpts, sink chan<- *EthereumCrossChainDebugN) (event.Subscription, error) {
			
			

			logs, sub, err := _EthereumCrossChain.contract.WatchLogs(opts, "DebugN")
			if err != nil {
				return nil, err
			}
			return event.NewSubscription(func(quit <-chan struct{}) error {
				defer sub.Unsubscribe()
				for {
					select {
					case log := <-logs:
						// New log arrived, parse the event and forward to the user
						event := new(EthereumCrossChainDebugN)
						if err := _EthereumCrossChain.contract.UnpackLog(event, "DebugN", log); err != nil {
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

		// ParseDebugN is a log parse operation binding the contract event 0xd4f190b7c2bbb19584eb23cae55f049509c7bf1c0fa046a9c93634b3983c953a.
		//
		// Solidity: event DebugN(uint256 )
		func (_EthereumCrossChain *EthereumCrossChainFilterer) ParseDebugN(log types.Log) (*EthereumCrossChainDebugN, error) {
			event := new(EthereumCrossChainDebugN)
			if err := _EthereumCrossChain.contract.UnpackLog(event, "DebugN", log); err != nil {
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

 	
		// EthereumCrossChainERC20CrossChainEventIterator is returned from FilterERC20CrossChainEvent and is used to iterate over the raw logs and unpacked data for ERC20CrossChainEvent events raised by the EthereumCrossChain contract.
		type EthereumCrossChainERC20CrossChainEventIterator struct {
			Event *EthereumCrossChainERC20CrossChainEvent // Event containing the contract specifics and raw log

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
		func (it *EthereumCrossChainERC20CrossChainEventIterator) Next() bool {
			// If the iterator failed, stop iterating
			if (it.fail != nil) {
				return false
			}
			// If the iterator completed, deliver directly whatever's available
			if (it.done) {
				select {
				case log := <-it.logs:
					it.Event = new(EthereumCrossChainERC20CrossChainEvent)
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
				it.Event = new(EthereumCrossChainERC20CrossChainEvent)
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
		func (it *EthereumCrossChainERC20CrossChainEventIterator) Error() error {
			return it.fail
		}
		// Close terminates the iteration process, releasing any pending underlying
		// resources.
		func (it *EthereumCrossChainERC20CrossChainEventIterator) Close() error {
			it.sub.Unsubscribe()
			return nil
		}

		// EthereumCrossChainERC20CrossChainEvent represents a ERC20CrossChainEvent event raised by the EthereumCrossChain contract.
		type EthereumCrossChainERC20CrossChainEvent struct { 
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

		// FilterERC20CrossChainEvent is a free log retrieval operation binding the contract event 0x363ffaf80f7a934d1d3bf7fc142b4e2925afd02adc95e00fe7693a74a24baf41.
		//
		// Solidity: event ERC20CrossChainEvent(string topic, address indexed sender, uint256 txId, address token, uint256 value, uint256 decimal, string toChainId, string toAddress, string rawdata)
 		func (_EthereumCrossChain *EthereumCrossChainFilterer) FilterERC20CrossChainEvent(opts *bind.FilterOpts, sender []common.Address) (*EthereumCrossChainERC20CrossChainEventIterator, error) {
			
			
			var senderRule []interface{}
			for _, senderItem := range sender {
				senderRule = append(senderRule, senderItem)
			}
			
			
			
			
			
			
			

			logs, sub, err := _EthereumCrossChain.contract.FilterLogs(opts, "ERC20CrossChainEvent", senderRule)
			if err != nil {
				return nil, err
			}
			return &EthereumCrossChainERC20CrossChainEventIterator{contract: _EthereumCrossChain.contract, event: "ERC20CrossChainEvent", logs: logs, sub: sub}, nil
 		}

		// WatchERC20CrossChainEvent is a free log subscription operation binding the contract event 0x363ffaf80f7a934d1d3bf7fc142b4e2925afd02adc95e00fe7693a74a24baf41.
		//
		// Solidity: event ERC20CrossChainEvent(string topic, address indexed sender, uint256 txId, address token, uint256 value, uint256 decimal, string toChainId, string toAddress, string rawdata)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) WatchERC20CrossChainEvent(opts *bind.WatchOpts, sink chan<- *EthereumCrossChainERC20CrossChainEvent, sender []common.Address) (event.Subscription, error) {
			
			
			var senderRule []interface{}
			for _, senderItem := range sender {
				senderRule = append(senderRule, senderItem)
			}
			
			
			
			
			
			
			

			logs, sub, err := _EthereumCrossChain.contract.WatchLogs(opts, "ERC20CrossChainEvent", senderRule)
			if err != nil {
				return nil, err
			}
			return event.NewSubscription(func(quit <-chan struct{}) error {
				defer sub.Unsubscribe()
				for {
					select {
					case log := <-logs:
						// New log arrived, parse the event and forward to the user
						event := new(EthereumCrossChainERC20CrossChainEvent)
						if err := _EthereumCrossChain.contract.UnpackLog(event, "ERC20CrossChainEvent", log); err != nil {
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

		// ParseERC20CrossChainEvent is a log parse operation binding the contract event 0x363ffaf80f7a934d1d3bf7fc142b4e2925afd02adc95e00fe7693a74a24baf41.
		//
		// Solidity: event ERC20CrossChainEvent(string topic, address indexed sender, uint256 txId, address token, uint256 value, uint256 decimal, string toChainId, string toAddress, string rawdata)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) ParseERC20CrossChainEvent(log types.Log) (*EthereumCrossChainERC20CrossChainEvent, error) {
			event := new(EthereumCrossChainERC20CrossChainEvent)
			if err := _EthereumCrossChain.contract.UnpackLog(event, "ERC20CrossChainEvent", log); err != nil {
				return nil, err
			}
			return event, nil
		}

 	
		// EthereumCrossChainETHCrossChainEventIterator is returned from FilterETHCrossChainEvent and is used to iterate over the raw logs and unpacked data for ETHCrossChainEvent events raised by the EthereumCrossChain contract.
		type EthereumCrossChainETHCrossChainEventIterator struct {
			Event *EthereumCrossChainETHCrossChainEvent // Event containing the contract specifics and raw log

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
		func (it *EthereumCrossChainETHCrossChainEventIterator) Next() bool {
			// If the iterator failed, stop iterating
			if (it.fail != nil) {
				return false
			}
			// If the iterator completed, deliver directly whatever's available
			if (it.done) {
				select {
				case log := <-it.logs:
					it.Event = new(EthereumCrossChainETHCrossChainEvent)
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
				it.Event = new(EthereumCrossChainETHCrossChainEvent)
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
		func (it *EthereumCrossChainETHCrossChainEventIterator) Error() error {
			return it.fail
		}
		// Close terminates the iteration process, releasing any pending underlying
		// resources.
		func (it *EthereumCrossChainETHCrossChainEventIterator) Close() error {
			it.sub.Unsubscribe()
			return nil
		}

		// EthereumCrossChainETHCrossChainEvent represents a ETHCrossChainEvent event raised by the EthereumCrossChain contract.
		type EthereumCrossChainETHCrossChainEvent struct { 
			Topic string; 
			Sender common.Address; 
			TxId *big.Int; 
			Token common.Address; 
			Value *big.Int; 
			ToChainId string; 
			ToAddress string; 
			Rawdata string; 
			Raw types.Log // Blockchain specific contextual infos
		}

		// FilterETHCrossChainEvent is a free log retrieval operation binding the contract event 0xc8a840d188e5175163eacef4950c691222c7c591788bdee7451a88868334a315.
		//
		// Solidity: event ETHCrossChainEvent(string topic, address indexed sender, uint256 txId, address token, uint256 value, string toChainId, string toAddress, string rawdata)
 		func (_EthereumCrossChain *EthereumCrossChainFilterer) FilterETHCrossChainEvent(opts *bind.FilterOpts, sender []common.Address) (*EthereumCrossChainETHCrossChainEventIterator, error) {
			
			
			var senderRule []interface{}
			for _, senderItem := range sender {
				senderRule = append(senderRule, senderItem)
			}
			
			
			
			
			
			

			logs, sub, err := _EthereumCrossChain.contract.FilterLogs(opts, "ETHCrossChainEvent", senderRule)
			if err != nil {
				return nil, err
			}
			return &EthereumCrossChainETHCrossChainEventIterator{contract: _EthereumCrossChain.contract, event: "ETHCrossChainEvent", logs: logs, sub: sub}, nil
 		}

		// WatchETHCrossChainEvent is a free log subscription operation binding the contract event 0xc8a840d188e5175163eacef4950c691222c7c591788bdee7451a88868334a315.
		//
		// Solidity: event ETHCrossChainEvent(string topic, address indexed sender, uint256 txId, address token, uint256 value, string toChainId, string toAddress, string rawdata)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) WatchETHCrossChainEvent(opts *bind.WatchOpts, sink chan<- *EthereumCrossChainETHCrossChainEvent, sender []common.Address) (event.Subscription, error) {
			
			
			var senderRule []interface{}
			for _, senderItem := range sender {
				senderRule = append(senderRule, senderItem)
			}
			
			
			
			
			
			

			logs, sub, err := _EthereumCrossChain.contract.WatchLogs(opts, "ETHCrossChainEvent", senderRule)
			if err != nil {
				return nil, err
			}
			return event.NewSubscription(func(quit <-chan struct{}) error {
				defer sub.Unsubscribe()
				for {
					select {
					case log := <-logs:
						// New log arrived, parse the event and forward to the user
						event := new(EthereumCrossChainETHCrossChainEvent)
						if err := _EthereumCrossChain.contract.UnpackLog(event, "ETHCrossChainEvent", log); err != nil {
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

		// ParseETHCrossChainEvent is a log parse operation binding the contract event 0xc8a840d188e5175163eacef4950c691222c7c591788bdee7451a88868334a315.
		//
		// Solidity: event ETHCrossChainEvent(string topic, address indexed sender, uint256 txId, address token, uint256 value, string toChainId, string toAddress, string rawdata)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) ParseETHCrossChainEvent(log types.Log) (*EthereumCrossChainETHCrossChainEvent, error) {
			event := new(EthereumCrossChainETHCrossChainEvent)
			if err := _EthereumCrossChain.contract.UnpackLog(event, "ETHCrossChainEvent", log); err != nil {
				return nil, err
			}
			return event, nil
		}

 	
		// EthereumCrossChainWithDrawERC20EventIterator is returned from FilterWithDrawERC20Event and is used to iterate over the raw logs and unpacked data for WithDrawERC20Event events raised by the EthereumCrossChain contract.
		type EthereumCrossChainWithDrawERC20EventIterator struct {
			Event *EthereumCrossChainWithDrawERC20Event // Event containing the contract specifics and raw log

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
		func (it *EthereumCrossChainWithDrawERC20EventIterator) Next() bool {
			// If the iterator failed, stop iterating
			if (it.fail != nil) {
				return false
			}
			// If the iterator completed, deliver directly whatever's available
			if (it.done) {
				select {
				case log := <-it.logs:
					it.Event = new(EthereumCrossChainWithDrawERC20Event)
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
				it.Event = new(EthereumCrossChainWithDrawERC20Event)
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
		func (it *EthereumCrossChainWithDrawERC20EventIterator) Error() error {
			return it.fail
		}
		// Close terminates the iteration process, releasing any pending underlying
		// resources.
		func (it *EthereumCrossChainWithDrawERC20EventIterator) Close() error {
			it.sub.Unsubscribe()
			return nil
		}

		// EthereumCrossChainWithDrawERC20Event represents a WithDrawERC20Event event raised by the EthereumCrossChain contract.
		type EthereumCrossChainWithDrawERC20Event struct { 
			Topic string; 
			Sender common.Address; 
			Token common.Address; 
			ToAddress common.Address; 
			Value *big.Int; 
			Raw types.Log // Blockchain specific contextual infos
		}

		// FilterWithDrawERC20Event is a free log retrieval operation binding the contract event 0x98ef7f0d02c5ec83cf62fda844997bad56b1e0f06cf91cedeffeb58bd08df06d.
		//
		// Solidity: event WithDrawERC20Event(string topic, address indexed sender, address token, address toAddress, uint256 value)
 		func (_EthereumCrossChain *EthereumCrossChainFilterer) FilterWithDrawERC20Event(opts *bind.FilterOpts, sender []common.Address) (*EthereumCrossChainWithDrawERC20EventIterator, error) {
			
			
			var senderRule []interface{}
			for _, senderItem := range sender {
				senderRule = append(senderRule, senderItem)
			}
			
			
			

			logs, sub, err := _EthereumCrossChain.contract.FilterLogs(opts, "WithDrawERC20Event", senderRule)
			if err != nil {
				return nil, err
			}
			return &EthereumCrossChainWithDrawERC20EventIterator{contract: _EthereumCrossChain.contract, event: "WithDrawERC20Event", logs: logs, sub: sub}, nil
 		}

		// WatchWithDrawERC20Event is a free log subscription operation binding the contract event 0x98ef7f0d02c5ec83cf62fda844997bad56b1e0f06cf91cedeffeb58bd08df06d.
		//
		// Solidity: event WithDrawERC20Event(string topic, address indexed sender, address token, address toAddress, uint256 value)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) WatchWithDrawERC20Event(opts *bind.WatchOpts, sink chan<- *EthereumCrossChainWithDrawERC20Event, sender []common.Address) (event.Subscription, error) {
			
			
			var senderRule []interface{}
			for _, senderItem := range sender {
				senderRule = append(senderRule, senderItem)
			}
			
			
			

			logs, sub, err := _EthereumCrossChain.contract.WatchLogs(opts, "WithDrawERC20Event", senderRule)
			if err != nil {
				return nil, err
			}
			return event.NewSubscription(func(quit <-chan struct{}) error {
				defer sub.Unsubscribe()
				for {
					select {
					case log := <-logs:
						// New log arrived, parse the event and forward to the user
						event := new(EthereumCrossChainWithDrawERC20Event)
						if err := _EthereumCrossChain.contract.UnpackLog(event, "WithDrawERC20Event", log); err != nil {
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

		// ParseWithDrawERC20Event is a log parse operation binding the contract event 0x98ef7f0d02c5ec83cf62fda844997bad56b1e0f06cf91cedeffeb58bd08df06d.
		//
		// Solidity: event WithDrawERC20Event(string topic, address indexed sender, address token, address toAddress, uint256 value)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) ParseWithDrawERC20Event(log types.Log) (*EthereumCrossChainWithDrawERC20Event, error) {
			event := new(EthereumCrossChainWithDrawERC20Event)
			if err := _EthereumCrossChain.contract.UnpackLog(event, "WithDrawERC20Event", log); err != nil {
				return nil, err
			}
			return event, nil
		}

 	
		// EthereumCrossChainWithDrawETHEventIterator is returned from FilterWithDrawETHEvent and is used to iterate over the raw logs and unpacked data for WithDrawETHEvent events raised by the EthereumCrossChain contract.
		type EthereumCrossChainWithDrawETHEventIterator struct {
			Event *EthereumCrossChainWithDrawETHEvent // Event containing the contract specifics and raw log

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
		func (it *EthereumCrossChainWithDrawETHEventIterator) Next() bool {
			// If the iterator failed, stop iterating
			if (it.fail != nil) {
				return false
			}
			// If the iterator completed, deliver directly whatever's available
			if (it.done) {
				select {
				case log := <-it.logs:
					it.Event = new(EthereumCrossChainWithDrawETHEvent)
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
				it.Event = new(EthereumCrossChainWithDrawETHEvent)
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
		func (it *EthereumCrossChainWithDrawETHEventIterator) Error() error {
			return it.fail
		}
		// Close terminates the iteration process, releasing any pending underlying
		// resources.
		func (it *EthereumCrossChainWithDrawETHEventIterator) Close() error {
			it.sub.Unsubscribe()
			return nil
		}

		// EthereumCrossChainWithDrawETHEvent represents a WithDrawETHEvent event raised by the EthereumCrossChain contract.
		type EthereumCrossChainWithDrawETHEvent struct { 
			Topic string; 
			Sender common.Address; 
			ToAddress common.Address; 
			Value *big.Int; 
			Raw types.Log // Blockchain specific contextual infos
		}

		// FilterWithDrawETHEvent is a free log retrieval operation binding the contract event 0x5614f9659cb7146440d26cd1cb85f8d39eae4467d95b092bd92391e81d502279.
		//
		// Solidity: event WithDrawETHEvent(string topic, address indexed sender, address toAddress, uint256 value)
 		func (_EthereumCrossChain *EthereumCrossChainFilterer) FilterWithDrawETHEvent(opts *bind.FilterOpts, sender []common.Address) (*EthereumCrossChainWithDrawETHEventIterator, error) {
			
			
			var senderRule []interface{}
			for _, senderItem := range sender {
				senderRule = append(senderRule, senderItem)
			}
			
			

			logs, sub, err := _EthereumCrossChain.contract.FilterLogs(opts, "WithDrawETHEvent", senderRule)
			if err != nil {
				return nil, err
			}
			return &EthereumCrossChainWithDrawETHEventIterator{contract: _EthereumCrossChain.contract, event: "WithDrawETHEvent", logs: logs, sub: sub}, nil
 		}

		// WatchWithDrawETHEvent is a free log subscription operation binding the contract event 0x5614f9659cb7146440d26cd1cb85f8d39eae4467d95b092bd92391e81d502279.
		//
		// Solidity: event WithDrawETHEvent(string topic, address indexed sender, address toAddress, uint256 value)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) WatchWithDrawETHEvent(opts *bind.WatchOpts, sink chan<- *EthereumCrossChainWithDrawETHEvent, sender []common.Address) (event.Subscription, error) {
			
			
			var senderRule []interface{}
			for _, senderItem := range sender {
				senderRule = append(senderRule, senderItem)
			}
			
			

			logs, sub, err := _EthereumCrossChain.contract.WatchLogs(opts, "WithDrawETHEvent", senderRule)
			if err != nil {
				return nil, err
			}
			return event.NewSubscription(func(quit <-chan struct{}) error {
				defer sub.Unsubscribe()
				for {
					select {
					case log := <-logs:
						// New log arrived, parse the event and forward to the user
						event := new(EthereumCrossChainWithDrawETHEvent)
						if err := _EthereumCrossChain.contract.UnpackLog(event, "WithDrawETHEvent", log); err != nil {
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

		// ParseWithDrawETHEvent is a log parse operation binding the contract event 0x5614f9659cb7146440d26cd1cb85f8d39eae4467d95b092bd92391e81d502279.
		//
		// Solidity: event WithDrawETHEvent(string topic, address indexed sender, address toAddress, uint256 value)
		func (_EthereumCrossChain *EthereumCrossChainFilterer) ParseWithDrawETHEvent(log types.Log) (*EthereumCrossChainWithDrawETHEvent, error) {
			event := new(EthereumCrossChainWithDrawETHEvent)
			if err := _EthereumCrossChain.contract.UnpackLog(event, "WithDrawETHEvent", log); err != nil {
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

