/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */


package types

import (
	"math/big"
)

var (
	_BIG_INT_0 = big.NewInt(0)
)

const (
	CODE_VAR_ENT = "ent_params"
	CODE_VAR_RET = "ret_params"
)

func BigInt0() *big.Int {
	return _BIG_INT_0
}

type ParamUData = map[string]interface{}

type QueryType = byte

const (
	QueryType_Contract  QueryType = 0
	QueryType_Nonce     QueryType = 1
	QueryType_Balance   QueryType = 2
	QueryType_Receipt   QueryType = 3
	QueryType_Existence QueryType = 4
)
