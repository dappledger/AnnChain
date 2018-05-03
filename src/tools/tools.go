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


package tools

import (
	"github.com/gogo/protobuf/proto"
)

func PbMarshal(msg proto.Message) []byte {
	ret, err := proto.Marshal(msg)
	if err != nil {
		return nil
	}
	return ret
}

func PbUnmarshal(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}

func CopyBytes(byts []byte) []byte {
	cp := make([]byte, len(byts))
	copy(cp, byts)
	return cp
}
