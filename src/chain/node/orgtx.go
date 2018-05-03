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


package node

import (
	"bytes"
	"fmt"
	"time"

	"github.com/dappledger/AnnChain/angine/types"
	civiltypes "github.com/dappledger/AnnChain/src/types"
)

const (
	_ byte = iota
	OrgRegister
	OrgCreate
	OrgJoin
	OrgLeave
	OrgDelete
)

var (
	OrgTag        = []byte{'o', 'r', 'g', 0x01}
	OrgCancelTag  = []byte{'o', 'r', 'g', 0x02}
	OrgConfirmTag = []byte{'o', 'r', 'g', 0x03}
)

type (
	// OrgTx contains action and information about a change in the org ledger that might happen after confirmation
	OrgTx struct {
		civiltypes.CivilTx

		App     string                 `json:"app"`
		Act     byte                   `json:"act"`
		ChainID string                 `json:"chainid"`
		Genesis types.GenesisDoc       `json:"genesis"`
		Config  map[string]interface{} `json:"config"`
		Time    time.Time              `json:"time"`
	}

	// OrgConfirmTx comes from a former OrgTx target node and on receiving every node write their org ledger down about the change caused by the OrgTx confirmed
	OrgConfirmTx struct {
		civiltypes.CivilTx

		ChainID    string            `json:"chainid"`
		Act        byte              `json:"act"`
		Validators [][]byte          `json:"validators"` // unused
		Time       time.Time         `json:"time"`
		TxHash     []byte            `json:"txhash"`
		Attributes map[string]string `json:"attribute"` // some additionals which we wanna inform others
	}

	// OrgCancelTx cancels pending OrgTx
	OrgCancelTx struct {
		civiltypes.CivilTx

		TxHash []byte    `json:"txhash"`
		Time   time.Time `json:"time"`
	}
)

var (
	ErrOrgAlreadyIn     = fmt.Errorf("already part the organization")
	ErrOrgFailToStop    = fmt.Errorf("fail to stop organization")
	ErrOrgNotExists     = fmt.Errorf("organization doesn't exist")
	ErrOrgExistsAlready = fmt.Errorf("organization already exists")
)

func IsOrgRelatedTx(tx []byte) bool {
	return bytes.HasPrefix(tx, []byte{'o', 'r', 'g'})
}

func IsOrgTx(tx []byte) bool {
	return bytes.Equal(tx[:4], OrgTag)
}

func IsOrgCancelTx(tx []byte) bool {
	return bytes.Equal(tx[:4], OrgCancelTag)
}

func IsOrgConfirmTx(tx []byte) bool {
	return bytes.Equal(tx[:4], OrgConfirmTag)
}
