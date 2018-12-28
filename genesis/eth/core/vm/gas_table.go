package vm

import (
	"math/big"

	"github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/eth/common/math"
	"github.com/dappledger/AnnChain/genesis/eth/params"
)

func memoryGasCost(mem *Memory, newMemSize *big.Int) (*big.Int, error) {
	if newMemSize.Uint64() == 0 {
		return big.NewInt(0), nil
	}

	if newMemSize.Uint64() > 0xffffffffe0 {
		return big.NewInt(0), errGasUintOverflow
	}

	newMemSizeWords := toWordSize(newMemSize)
	newMemSize = newMemSizeWords.Mul(newMemSizeWords, big.NewInt(32))

	if newMemSize.Uint64() > uint64(mem.Len()) {
		square := newMemSizeWords.Uint64() * newMemSizeWords.Uint64()
		linCoef := newMemSizeWords.Uint64() * params.MemoryGas.Uint64()
		quadCoef := square / params.QuadCoeffDiv.Uint64()
		newTotalFee := linCoef + quadCoef

		fee := newTotalFee - mem.lastGasCost
		mem.lastGasCost = newTotalFee

		return new(big.Int).SetUint64(fee), nil
	}
	return big.NewInt(0), nil
}

func constGasFunc(gas *big.Int) gasFunc {
	return func(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
		return gas, nil
	}
}

func gasCalldataCopy(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}

	newGas := gas.Uint64()
	var overflow bool
	if newGas, overflow = math.SafeAdd(newGas, GasFastestStep.Uint64()); overflow {
		return big.NewInt(0), errGasUintOverflow
	}

	words, overflow := bigUint64(stack.Back(2))
	if overflow {
		return big.NewInt(0), errGasUintOverflow
	}

	if words, overflow = math.SafeMul(toWordSize(new(big.Int).SetUint64(words)).Uint64(), params.CopyGas.Uint64()); overflow {
		return big.NewInt(0), errGasUintOverflow
	}

	if newGas, overflow = math.SafeAdd(newGas, words); overflow {
		return big.NewInt(0), errGasUintOverflow
	}
	return new(big.Int).SetUint64(newGas), nil
}

func gasSStore(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	var (
		y, x   = stack.Back(1), stack.Back(0)
		val, _ = env.StateDB.GetState(contract.Address(), common.BigToHash(x))
	)
	// This checks for 3 scenario's and calculates gas accordingly
	// 1. From a zero-value address to a non-zero value         (NEW VALUE)
	// 2. From a non-zero value address to a zero-value address (DELETE)
	// 3. From a non-zero to a non-zero                         (CHANGE)
	if common.EmptyHash(val) && !common.EmptyHash(common.BigToHash(y)) {
		// 0 => non 0
		return new(big.Int).Set(params.SstoreSetGas), nil
	} else if !common.EmptyHash(val) && common.EmptyHash(common.BigToHash(y)) {
		env.StateDB.AddRefund(params.SstoreRefundGas)

		return new(big.Int).Set(params.SstoreClearGas), nil
	} else {
		// non 0 => non 0 (or 0 => 0)
		return new(big.Int).Set(params.SstoreResetGas), nil
	}
}

func makeGasLog(n uint) gasFunc {
	return func(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
		requestedSize, overflow := bigUint64(stack.Back(1))
		if overflow {
			return big.NewInt(0), errGasUintOverflow
		}

		gas, err := memoryGasCost(mem, memorySize)
		if err != nil {
			return big.NewInt(0), err
		}

		newGas := gas.Uint64()

		if newGas, overflow = math.SafeAdd(newGas, params.LogGas.Uint64()); overflow {
			return big.NewInt(0), errGasUintOverflow
		}
		if newGas, overflow = math.SafeAdd(newGas, uint64(n)*params.LogTopicGas.Uint64()); overflow {
			return big.NewInt(0), errGasUintOverflow
		}

		var memorySizeGas uint64
		if memorySizeGas, overflow = math.SafeMul(requestedSize, params.LogDataGas.Uint64()); overflow {
			return big.NewInt(0), errGasUintOverflow
		}
		if newGas, overflow = math.SafeAdd(newGas, memorySizeGas); overflow {
			return big.NewInt(0), errGasUintOverflow
		}
		return new(big.Int).SetUint64(newGas), nil
	}
}

func gasSha3(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}

	newGas := gas.Uint64()
	var overflow bool
	if newGas, overflow = math.SafeAdd(newGas, params.Sha3Gas.Uint64()); overflow {
		return big.NewInt(0), errGasUintOverflow
	}

	wordGas, overflow := bigUint64(stack.Back(1))
	if overflow {
		return big.NewInt(0), errGasUintOverflow
	}
	if wordGas, overflow = math.SafeMul(toWordSize(new(big.Int).SetUint64(wordGas)).Uint64(), params.Sha3WordGas.Uint64()); overflow {
		return big.NewInt(0), errGasUintOverflow
	}
	if newGas, overflow = math.SafeAdd(newGas, wordGas); overflow {
		return big.NewInt(0), errGasUintOverflow
	}
	return gas, nil
}

func gasCodeCopy(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}
	gas.Add(gas, GasFastestStep)
	words := toWordSize(stack.Back(2))

	return gas.Add(gas, words.Mul(words, params.CopyGas)), nil
}

func gasExtCodeCopy(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}
	gas.Add(gas, gt.ExtcodeCopy)
	words := toWordSize(stack.Back(3))

	return gas.Add(gas, words.Mul(words, params.CopyGas)), nil
}

func gasExtCodeHash(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	return gt.ExtcodeHash, nil
}

func gasCreate2(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}
	gas.Add(gas, params.Create2Gas)

	wordGas, overflow := bigUint64(stack.Back(2))
	if overflow {
		return big.NewInt(0), errGasUintOverflow
	}

	if wordGas, overflow = math.SafeMul(toWordSize(new(big.Int).SetUint64(wordGas)).Uint64(), params.Sha3WordGas.Uint64()); overflow {
		return big.NewInt(0), errGasUintOverflow
	}

	gas.Add(gas, new(big.Int).SetUint64(wordGas))
	return gas, nil
}

func gasMLoad(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}
	return new(big.Int).Add(GasFastestStep, gas), nil
}

func gasMStore8(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}
	return new(big.Int).Add(GasFastestStep, gas), nil
}

func gasMStore(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}
	return new(big.Int).Add(GasFastestStep, gas), nil
}

func gasCreate(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}
	return new(big.Int).Add(params.CreateGas, gas), nil
}

func gasBalance(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	return gt.Balance, nil
}

func gasExtCodeSize(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	return gt.ExtcodeSize, nil
}

func gasSLoad(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	return gt.SLoad, nil
}

func gasExp(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	expByteLen := int64((stack.data[stack.len()-2].BitLen() + 7) / 8)
	gas := big.NewInt(expByteLen)
	gas.Mul(gas, gt.ExpByte)
	return gas.Add(gas, GasSlowStep), nil
}

func gasStaticCall(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}
	gas.Add(gas, gt.Calls)

	evm.callGasTemp = callGas(gt, contract.Gas, gas, stack.Back(0))
	if err != nil {
		return big.NewInt(0), err
	}
	gas.Add(gas, evm.callGasTemp)

	return gas, nil
}

func gasRevert(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	return memoryGasCost(mem, memorySize)
}

func gasCall(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas := new(big.Int).Set(gt.Calls)
	transfersValue := stack.Back(2).Sign() != 0
	var (
		address = common.BigToAddress(stack.Back(1))
	)
	if env.StateDB.Empty(address) && transfersValue {
		gas.Add(gas, params.CallNewAccountGas)
	}

	if transfersValue {
		gas.Add(gas, params.CallValueTransferGas)
	}

	memGas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}

	gas.Add(gas, memGas)
	env.callGasTemp = callGas(gt, contract.Gas, gas, stack.data[stack.len()-1])

	// Replace the stack item with the new gas calculation. This means that
	// either the original item is left on the stack or the item is replaced by:
	// (availableGas - gas) * 63 / 64
	// We replace the stack item so that it's available when the opCall instruction is
	// called. This information is otherwise lost due to the dependency on *current*
	// available gas.
	//	stack.data[stack.len()-1] = cg
	return gas.Add(gas, env.callGasTemp), nil
}

func gasCallCode(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	var (
		transfersValue = stack.Back(2).Sign() != 0
		address        = common.BigToAddress(stack.Back(1))
	)
	gas := new(big.Int).Set(gt.Calls)
	if transfersValue && env.StateDB.Empty(address) {
		gas.Add(gas, params.CallNewAccountGas)
	}

	if transfersValue {
		gas.Add(gas, params.CallValueTransferGas)
	}

	memGas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}
	gas.Add(gas, memGas)

	env.callGasTemp = callGas(gt, contract.Gas, gas, stack.data[stack.len()-1])
	// Replace the stack item with the new gas calculation. This means that
	// either the original item is left on the stack or the item is replaced by:
	// (availableGas - gas) * 63 / 64
	// We replace the stack item so that it's available when the opCall instruction is
	// called. This information is otherwise lost due to the dependency on *current*
	// available gas.
	//	stack.data[stack.len()-1] = cg

	return gas.Add(gas, env.callGasTemp), nil
}

func gasReturn(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}
	return gas, nil
}

func gasReturnDataCopy(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}

	gas.Add(gas, GasFastestStep)

	words, overflow := bigUint64(stack.Back(2))
	if overflow {
		return big.NewInt(0), errGasUintOverflow
	}

	if words, overflow = math.SafeMul(toWordSize(new(big.Int).SetUint64(words)).Uint64(), params.CopyGas.Uint64()); overflow {
		return big.NewInt(0), errGasUintOverflow
	}

	gas.Add(gas, new(big.Int).SetUint64(words))
	return gas, nil
}

func gasSuicide(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas := new(big.Int)
	// EIP150 homestead gas reprice fork:
	if env.ChainConfig().IsEIP150(env.BlockNumber) {
		gas.Set(gt.Suicide)
		var (
			address = common.BigToAddress(stack.Back(0))
			eip158  = env.ChainConfig().IsEIP158(env.BlockNumber)
		)

		if eip158 {
			// if empty and transfers value
			if env.StateDB.Empty(address) && env.StateDB.GetBalance(contract.Address()).BitLen() > 0 {
				gas.Add(gas, gt.CreateBySuicide)
			}
		} else if !env.StateDB.Exist(address) {
			gas.Add(gas, gt.CreateBySuicide)
		}
	}

	if !env.StateDB.HasSuicided(contract.Address()) {
		env.StateDB.AddRefund(params.SuicideRefundGas)
	}
	return gas, nil
}

func gasDelegateCall(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return big.NewInt(0), err
	}
	gas = new(big.Int).Add(gt.Calls, gas)

	env.callGasTemp = callGas(gt, contract.Gas, gas, stack.data[stack.len()-1])
	// Replace the stack item with the new gas calculation. This means that
	// either the original item is left on the stack or the item is replaced by:
	// (availableGas - gas) * 63 / 64
	// We replace the stack item so that it's available when the opCall instruction is
	// called.
	//	stack.data[stack.len()-1] = cg

	return gas.Add(gas, env.callGasTemp), nil
}

func gasPush(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	return GasFastestStep, nil
}

func gasSwap(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	return GasFastestStep, nil
}

func gasDup(gt params.GasTable, env *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize *big.Int) (*big.Int, error) {
	return GasFastestStep, nil
}
