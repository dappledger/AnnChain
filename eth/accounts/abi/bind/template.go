// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package bind

import "github.com/dappledger/AnnChain/eth/accounts/abi"

// tmplData is the data structure required to fill the binding template.
type tmplData struct {
	Package   string                   // Name of the package to place the generated file in
	Contracts map[string]*tmplContract // List of contracts to generate into this file
}

// tmplContract contains the data needed to generate an individual contract binding.
type tmplContract struct {
	Type        string                 // Type name of the main contract binding
	InputABI    string                 // JSON ABI used as the input to generate the binding from
	InputBin    string                 // Optional EVM bytecode used to denetare deploy code from
	Constructor abi.Method             // Contract constructor for deploy parametrization
	Calls       map[string]*tmplMethod // Contract calls that only read state data
	Transacts   map[string]*tmplMethod // Contract calls that write state data
	Events      map[string]*tmplEvent  // Contract events accessors
}

// tmplMethod is a wrapper around an abi.Method that contains a few preprocessed
// and cached data fields.
type tmplMethod struct {
	Original   abi.Method // Original method as parsed by the abi package
	Normalized abi.Method // Normalized version of the parsed method (capitalized names, non-anonymous args/returns)
	Structured bool       // Whether the returns should be accumulated into a struct
}

// tmplEvent is a wrapper around an a
type tmplEvent struct {
	Original   abi.Event // Original event as parsed by the abi package
	Normalized abi.Event // Normalized version of the parsed fields
}

// tmplSource is language to template mapping containing all the supported
// programming languages the package can generate to.
var tmplSource = map[Lang]string{
	LangGo:   tmplSourceGo,
	LangJava: tmplSourceJava,
}

// tmplSourceGo is the Go source template use to generate the contract binding
// based on.
const tmplSourceGo = `
// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package {{.Package}}

{{range $contract := .Contracts}}
	// {{.Type}}ABI is the input ABI used to generate the binding from.
	const {{.Type}}ABI = "{{.InputABI}}"

	{{if .InputBin}}
		// {{.Type}}Bin is the compiled bytecode used for deploying new contracts.
		const {{.Type}}Bin = ` + "`" + `{{.InputBin}}` + "`" + `

		// Deploy{{.Type}} deploys a new Ethereum contract, binding an instance of {{.Type}} to it.
		func Deploy{{.Type}}(goSDK *rpcclient.GoSDK, auth rpcclient.AccountBase {{range .Constructor.Inputs}}, {{.Name}} {{bindtype .Type}}{{end}}) (common.Hash, *{{.Type}}, error) {

		  cc := rpcclient.ContractCreate{
			AccountBase: auth,
			Code:        {{.Type}}Bin,
			ABI:         {{.Type}}ABI,
			Params:      []interface{}{ {{range $i, $_ := .Constructor.Inputs}}{{if ne $i 0}}, {{end}}{{.Name}}{{end}} },
		  }
		  ret, err := goSDK.ContractCreate(&cc)
		  if err != nil {
			return common.Hash{}, nil, err
		  }
		  txHash := common.HexToHash(ret["tx"].(string))
		  address := common.HexToAddress(ret["contract"].(string))
		  return txHash, &{{.Type}}{address: address, cli: goSDK}, nil
		}
	{{end}}

	// {{.Type}} is an auto generated Go binding around an Ethereum contract.
	type {{.Type}} struct {
	  address common.Address
	  cli     *rpcclient.GoSDK
	}

	// New{{.Type}} creates a new instance of {{.Type}}, bound to a specific deployed contract.
	func New{{.Type}}(goSdk *rpcclient.GoSDK, address common.Address) *{{.Type}} {
	  return &{{.Type}}{address: address, cli: goSdk}
	}

	func (_{{.Type}} *{{.Type}}) GetAddress()(common.Address){
		return _{{.Type}}.address
	}

	{{range .Calls}}
		
		func (_{{$contract.Type}} *{{$contract.Type}}) {{.Normalized.Name}}(auth rpcclient.AccountBase {{range .Normalized.Inputs}}, {{.Name}} {{bindtype .Type}} {{end}}) ( {{range .Normalized.Outputs}}{{.Name}} {{bindtype .Type}},{{end}} error) {
			
			var (
				{{range $i, $_ := .Normalized.Outputs}}ret{{$i}}  {{bindtype .Type}}
				{{end}}
			)
			m := rpcclient.ContractMethod{
				AccountBase: auth,
				Contract:    _{{$contract.Type}}.address.Hex(),
				ABI:         {{$contract.Type}}ABI,
				Method:      "{{.Original.Name}}",
				Params:      []interface{}{ {{range $i, $_ := .Normalized.Inputs}}{{if ne $i 0}}, {{end}} {{.Name}} {{end}} },
			}

			ret, err := _{{$contract.Type}}.cli.ContractRead(&m)
			arr := ret.([]interface{})
			{{range $i, $_ := .Normalized.Outputs}}ret{{$i}} = arr[{{$i}}].({{bindtype .Type}})
			{{end}}
			return {{range $i, $_ := .Normalized.Outputs}}ret{{$i}}, {{end}}err
		}

		func (_{{$contract.Type}} *{{$contract.Type}}) {{.Normalized.Name}}ByHeight(auth rpcclient.AccountBase {{range .Normalized.Inputs}}, {{.Name}} {{bindtype .Type}} {{end}}, height uint64) ( {{range .Normalized.Outputs}}{{.Name}} {{bindtype .Type}},{{end}} error) {	
			var (
				{{range $i, $_ := .Normalized.Outputs}}ret{{$i}}  {{bindtype .Type}}
				{{end}}
			)
			m := rpcclient.ContractMethod{
				AccountBase: auth,
				Contract:    _{{$contract.Type}}.address.Hex(),
				ABI:         {{$contract.Type}}ABI,
				Method:      "{{.Original.Name}}",
				Params:      []interface{}{ {{range $i, $_ := .Normalized.Inputs}}{{if ne $i 0}}, {{end}} {{.Name}} {{end}} },
			}

			ret, err := _{{$contract.Type}}.cli.ContractReadByHeight(&m, height)
			arr := ret.([]interface{})
			{{range $i, $_ := .Normalized.Outputs}}ret{{$i}} = arr[{{$i}}].({{bindtype .Type}})
			{{end}}
			return {{range $i, $_ := .Normalized.Outputs}}ret{{$i}}, {{end}}err
		}

	{{end}}

	{{range .Transacts}}
		func (_{{$contract.Type}} *{{$contract.Type}}) {{.Normalized.Name}}(auth rpcclient.AccountBase {{range .Normalized.Inputs}}, {{.Name}} {{bindtype .Type}} {{end}}) ( txHash string, err error) {
			
			m := rpcclient.ContractMethod{
				AccountBase: auth,
				Contract:    _{{$contract.Type}}.address.Hex(),
				ABI:         {{$contract.Type}}ABI,
				Method:      "{{.Original.Name}}",
				Params:      []interface{}{ {{range $i, $_ := .Normalized.Inputs}}{{if ne $i 0}}, {{end}} {{.Name}} {{end}} },
			}

			return _{{$contract.Type}}.cli.ContractAsync(&m)
		}

		func (_{{$contract.Type}} *{{$contract.Type}}) {{.Normalized.Name}}ByHeight(auth rpcclient.AccountBase {{range .Normalized.Inputs}}, {{.Name}} {{bindtype .Type}} {{end}}, height uint64) ( {{range .Normalized.Outputs}}{{.Name}} {{bindtype .Type}},{{end}} error) {
			var (
				{{range $i, $_ := .Normalized.Outputs}}ret{{$i}}  {{bindtype .Type}}
				{{end}}
			)
			m := rpcclient.ContractMethod{
				AccountBase: auth,
				Contract:    _{{$contract.Type}}.address.Hex(),
				ABI:         {{$contract.Type}}ABI,
				Method:      "{{.Original.Name}}",
				Params:      []interface{}{ {{range $i, $_ := .Normalized.Inputs}}{{if ne $i 0}}, {{end}} {{.Name}} {{end}} },
			}

			ret, err := _{{$contract.Type}}.cli.ContractReadByHeight(&m, height)
			arr := ret.([]interface{})
			{{range $i, $_ := .Normalized.Outputs}}ret{{$i}} = arr[{{$i}}].({{bindtype .Type}})
			{{end}}
			_ = arr
			return {{range $i, $_ := .Normalized.Outputs}}ret{{$i}}, {{end}}err
		}
	{{end}}
{{end}}
`

// tmplSourceJava is the Java source template use to generate the contract binding
// based on.
const tmplSourceJava = `
// This file is an automatically generated Java binding. Do not modify as any
// change will likely be lost upon the next re-generation!

package {{.Package}};

import org.ethereum.geth.*;
import org.ethereum.geth.internal.*;

{{range $contract := .Contracts}}
	public class {{.Type}} {
		// ABI is the input ABI used to generate the binding from.
		public final static String ABI = "{{.InputABI}}";

		{{if .InputBin}}
			// BYTECODE is the compiled bytecode used for deploying new contracts.
			public final static byte[] BYTECODE = "{{.InputBin}}".getBytes();

			// deploy deploys a new Ethereum contract, binding an instance of {{.Type}} to it.
			public static {{.Type}} deploy(TransactOpts auth, EthereumClient rpcclient{{range .Constructor.Inputs}}, {{bindtype .Type}} {{.Name}}{{end}}) throws Exception {
				Interfaces args = Geth.newInterfaces({{(len .Constructor.Inputs)}});
				{{range $index, $element := .Constructor.Inputs}}
				  args.set({{$index}}, Geth.newInterface()); args.get({{$index}}).set{{namedtype (bindtype .Type) .Type}}({{.Name}});
				{{end}}
				return new {{.Type}}(Geth.deployContract(auth, ABI, BYTECODE, rpcclient, args));
			}

			// Internal constructor used by contract deployment.
			private {{.Type}}(BoundContract deployment) {
				this.Address  = deployment.getAddress();
				this.Deployer = deployment.getDeployer();
				this.Contract = deployment;
			}
		{{end}}

		// Ethereum address where this contract is located at.
		public final Address Address;

		// Ethereum transaction in which this contract was deployed (if known!).
		public final Transaction Deployer;

		// Contract instance bound to a blockchain address.
		private final BoundContract Contract;

		// Creates a new instance of {{.Type}}, bound to a specific deployed contract.
		public {{.Type}}(Address address, EthereumClient rpcclient) throws Exception {
			this(Geth.bindContract(address, ABI, rpcclient));
		}

		{{range .Calls}}
			{{if gt (len .Normalized.Outputs) 1}}
			// {{capitalise .Normalized.Name}}Results is the output of a call to {{.Normalized.Name}}.
			public class {{capitalise .Normalized.Name}}Results {
				{{range $index, $item := .Normalized.Outputs}}public {{bindtype .Type}} {{if ne .Name ""}}{{.Name}}{{else}}Return{{$index}}{{end}};
				{{end}}
			}
			{{end}}

			// {{.Normalized.Name}} is a free data retrieval call binding the contract method 0x{{printf "%x" .Original.Id}}.
			//
			// Solidity: {{.Original.String}}
			public {{if gt (len .Normalized.Outputs) 1}}{{capitalise .Normalized.Name}}Results{{else}}{{range .Normalized.Outputs}}{{bindtype .Type}}{{end}}{{end}} {{.Normalized.Name}}(CallOpts opts{{range .Normalized.Inputs}}, {{bindtype .Type}} {{.Name}}{{end}}) throws Exception {
				Interfaces args = Geth.newInterfaces({{(len .Normalized.Inputs)}});
				{{range $index, $item := .Normalized.Inputs}}args.set({{$index}}, Geth.newInterface()); args.get({{$index}}).set{{namedtype (bindtype .Type) .Type}}({{.Name}});
				{{end}}

				Interfaces results = Geth.newInterfaces({{(len .Normalized.Outputs)}});
				{{range $index, $item := .Normalized.Outputs}}Interface result{{$index}} = Geth.newInterface(); result{{$index}}.setDefault{{namedtype (bindtype .Type) .Type}}(); results.set({{$index}}, result{{$index}});
				{{end}}

				if (opts == null) {
					opts = Geth.newCallOpts();
				}
				this.Contract.call(opts, results, "{{.Original.Name}}", args);
				{{if gt (len .Normalized.Outputs) 1}}
					{{capitalise .Normalized.Name}}Results result = new {{capitalise .Normalized.Name}}Results();
					{{range $index, $item := .Normalized.Outputs}}result.{{if ne .Name ""}}{{.Name}}{{else}}Return{{$index}}{{end}} = results.get({{$index}}).get{{namedtype (bindtype .Type) .Type}}();
					{{end}}
					return result;
				{{else}}{{range .Normalized.Outputs}}return results.get(0).get{{namedtype (bindtype .Type) .Type}}();{{end}}
				{{end}}
			}
		{{end}}

		{{range .Transacts}}
			// {{.Normalized.Name}} is a paid mutator transaction binding the contract method 0x{{printf "%x" .Original.Id}}.
			//
			// Solidity: {{.Original.String}}
			public Transaction {{.Normalized.Name}}(TransactOpts opts{{range .Normalized.Inputs}}, {{bindtype .Type}} {{.Name}}{{end}}) throws Exception {
				Interfaces args = Geth.newInterfaces({{(len .Normalized.Inputs)}});
				{{range $index, $item := .Normalized.Inputs}}args.set({{$index}}, Geth.newInterface()); args.get({{$index}}).set{{namedtype (bindtype .Type) .Type}}({{.Name}});
				{{end}}

				return this.Contract.transact(opts, "{{.Original.Name}}"	, args);
			}
		{{end}}
	}
{{end}}
`
