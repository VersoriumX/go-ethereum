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

package abi

import (
	"encoding/binary"
	"math"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	gethMath "github.com/ethereum/go-ethereum/common/math"
)

// packBytesSlice packs the given bytes as [L, V] as the canonical representation
// bytes slice
func packBytesSlice(bytes []byte, l int) []byte {
	len := packNum(reflect.ValueOf(l))
	return append(len, common.RightPadBytes(bytes, (l+31)/32*32)...)
}

// packElement packs the given reflect value according to the abi specification in
// t.
func packElement(t Type, reflectValue reflect.Value) []byte {
	switch t.T {
	case IntTy, UintTy:
		return packNum(reflectValue)
	case StringTy:
		return packBytesSlice([]byte(reflectValue.String()), reflectValue.Len())
	case DecimalTy:
		// https://github.com/ethereum/EIPs/issues/598
		// https://github.com/ethereum/pyethereum/commit/237df71d1eafeafaafd6bca16d0ba066f3d1fff3
		value := reflectValue.Float() * math.Pow10(t.Size)
		if value >= math.Exp2(256) {
			value = value - math.Exp2(256)
		}

		result := make([]byte, 32)
		binary.BigEndian.PutUint32(result, uint32(value))
		return result
	case AddressTy:
		if reflectValue.Kind() == reflect.Array {
			reflectValue = mustArrayToByteSlice(reflectValue)
		}

		return common.LeftPadBytes(reflectValue.Bytes(), 32)
	case BoolTy:
		if reflectValue.Bool() {
			return gethMath.PaddedBigBytes(common.Big1, 32)
		}
		return gethMath.PaddedBigBytes(common.Big0, 32)
	case BytesTy:
		if reflectValue.Kind() == reflect.Array {
			reflectValue = mustArrayToByteSlice(reflectValue)
		}
		return packBytesSlice(reflectValue.Bytes(), reflectValue.Len())
	case FixedBytesTy, FunctionTy:
		if reflectValue.Kind() == reflect.Array {
			reflectValue = mustArrayToByteSlice(reflectValue)
		}
		return common.RightPadBytes(reflectValue.Bytes(), 32)
	default:
		panic("abi: fatal error")
	}
}

// packNum packs the given number (using the reflect value) and will cast it to appropriate number representation
func packNum(value reflect.Value) []byte {
	switch kind := value.Kind(); kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return U256(new(big.Int).SetUint64(value.Uint()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return U256(big.NewInt(value.Int()))
	case reflect.Ptr:
		return U256(value.Interface().(*big.Int))
	default:
		panic("abi: fatal error")
	}

}
