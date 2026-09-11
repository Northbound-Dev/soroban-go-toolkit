package soroban

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/stellar/go/xdr"
)

// DecodeString decodes an ScVal containing a string value.
func DecodeString(v xdr.ScVal) (string, error) {
	switch v.Type {
	case xdr.ScValTypeScvString:
		var str string
		if err := xdr.SafeUnmarshalBase64(v.StringVal, &str); err != nil {
			return "", err
		}
		return str, nil
	default:
		return "", ErrWrongType{Expected: "string", Actual: v.Type.String()}
	}
}

// DecodeBytes decodes an ScVal containing a byte array value.
func DecodeBytes(v xdr.ScVal) ([]byte, error) {
	switch v.Type {
	case xdr.ScValTypeScvBytes:
		var data []byte
		if err := xdr.SafeUnmarshalBase64(v.BytesVal, &data); err != nil {
			return nil, err
		}
		return data, nil
	default:
		return nil, ErrWrongType{Expected: "bytes", Actual: v.Type.String()}
	}
}

// DecodeBool decodes an ScVal containing a boolean value.
func DecodeBool(v xdr.ScVal) (bool, error) {
	switch v.Type {
	case xdr.ScValTypeScvBool:
		return v.BoolVal, nil
	default:
		return false, ErrWrongType{Expected: "bool", Actual: v.Type.String()}
	}
}

// DecodeI32 decodes an ScVal containing a 32-bit integer value.
func DecodeI32(v xdr.ScVal) (int32, error) {
	switch v.Type {
	case xdr.ScValTypeScvI32:
		return v.I32Val, nil
	default:
		return 0, ErrWrongType{Expected: "i32", Actual: v.Type.String()}
	}
}

// DecodeU32 decodes an ScVal containing a 32-bit unsigned integer value.
func DecodeU32(v xdr.ScVal) (uint32, error) {
	switch v.Type {
	case xdr.ScValTypeScvU32:
		return v.U32Val, nil
	default:
		return 0, ErrWrongType{Expected: "u32", Actual: v.Type.String()}
	}
}

// DecodeI64 decodes an ScVal containing a 64-bit integer value.
func DecodeI64(v xdr.ScVal) (int64, error) {
	switch v.Type {
	case xdr.ScValTypeScvI64:
		return v.I64Val, nil
	default:
		return 0, ErrWrongType{Expected: "i64", Actual: v.Type.String()}
	}
}

// DecodeU64 decodes an ScVal containing a 64-bit unsigned integer value.
func DecodeU64(v xdr.ScVal) (uint64, error) {
	switch v.Type {
	case xdr.ScValTypeScvU64:
		return v.U64Val, nil
	default:
		return 0, ErrWrongType{Expected: "u64", Actual: v.Type.String()}
	}
}

// DecodeI128 decodes an ScVal containing a 128-bit integer value.
// Returns as *big.Int to handle the full range.
func DecodeI128(v xdr.ScVal) (*big.Int, error) {
	switch v.Type {
	case xdr.ScValTypeScvI128:
		// XDR encodes i128 as two uint64 in big-endian order
		var buf [16]byte
		binary.BigEndian.PutUint64(buf[0:8], v.U128Val.Hi)
		binary.BigEndian.PutUint64(buf[8:16], v.U128Val.Lo)
		// Interpret as signed two's complement
		if buf[0]&0x80 != 0 {
			// Negative number - sign extend
			for i := 0; i < 16; i++ {
				buf[i] = ^buf[i]
			}
			// Add 1 to make it negative
			var carry uint64 = 1
			for i := 15; i >= 0; i-- {
				val := uint64(buf[i]) + carry
				buf[i] = byte(val)
				carry = val >> 8
			}
		}
		return new(big.Int).SetBytes(buf[:]), nil
	default:
		return nil, ErrWrongType{Expected: "i128", Actual: v.Type.String()}
	}
}

// DecodeU128 decodes an ScVal containing a 128-bit unsigned integer value.
// Returns as *big.Int to handle the full range.
func DecodeU128(v xdr.ScVal) (*big.Int, error) {
	switch v.Type {
	case xdr.ScValTypeScvU128:
		var buf [16]byte
		binary.BigEndian.PutUint64(buf[0:8], v.U128Val.Hi)
		binary.BigEndian.PutUint64(buf[8:16], v.U128Val.Lo)
		return new(big.Int).SetBytes(buf[:]), nil
	default:
		return nil, ErrWrongType{Expected: "u128", Actual: v.Type.String()}
	}
}

// DecodeI256 decodes an ScVal containing a 256-bit integer value.
// Returns as *big.Int to handle the full range.
func DecodeI256(v xdr.ScVal) (*big.Int, error) {
	switch v.Type {
	case xdr.ScValTypeScvI256:
		// XDR encodes i256 as four uint64 in big-endian order
		var buf [32]byte
		binary.BigEndian.PutUint64(buf[0:8], v.U256Val.Hi[0])
		binary.BigEndian.PutUint64(buf[8:16], v.U256Val.Hi[1])
		binary.BigEndian.PutUint64(buf[16:24], v.U256Val.Lo[0])
		binary.BigEndian.PutUint64(buf[24:32], v.U256Val.Lo[1])
		// Interpret as signed two's complement
		if buf[0]&0x80 != 0 {
			// Negative number - sign extend
			for i := 0; i < 32; i++ {
				buf[i] = ^buf[i]
			}
			// Add 1 to make it negative
			var carry uint64 = 1
			for i := 31; i >= 0; i-- {
				val := uint64(buf[i]) + carry
				buf[i] = byte(val)
				carry = val >> 8
			}
		}
		return new(big.Int).SetBytes(buf[:]), nil
	default:
		return nil, ErrWrongType{Expected: "i256", Actual: v.Type.String()}
	}
}

// DecodeU256 decodes an ScVal containing a 256-bit unsigned integer value.
// Returns as *big.Int to handle the full range.
func DecodeU256(v xdr.ScVal) (*big.Int, error) {
	switch v.Type {
	case xdr.ScValTypeScvU256:
		var buf [32]byte
		binary.BigEndian.PutUint64(buf[0:8], v.U256Val.Hi[0])
		binary.BigEndian.PutUint64(buf[8:16], v.U256Val.Hi[1])
		binary.BigEndian.PutUint64(buf[16:24], v.U256Val.Lo[0])
		binary.BigEndian.PutUint64(buf[24:32], v.U256Val.Lo[1])
		return new(big.Int).SetBytes(buf[:]), nil
	default:
		return nil, ErrWrongType{Expected: "u256", Actual: v.Type.String()}
	}
}

// DecodeVec decodes an ScVal containing a vector (array) value using the
// provided element decoder function.
func DecodeVec[T any](v xdr.ScVal, decodeElem func(xdr.ScVal) (T, error)) ([]T, error) {
	switch v.Type {
	case xdr.ScValTypeScvVec:
		var result []T
		for _, elem := v.VecVal.Elements; ; elem = elem.Next {
			if elem == nil {
				break
			}
			val, err := decodeElem(elem.Value)
			if err != nil {
				return nil, err
			}
			result = append(result, val)
		}
		return result, nil
	default:
		return nil, ErrWrongType{Expected: "vec", Actual: v.Type.String()}
	}
}

// DecodeMap decodes an ScVal containing a map value using the provided key
// and value decoder functions.
func DecodeMap[K comparable, V any](v xdr.ScVal, decodeKey func(xdr.ScVal) (K, error), decodeValue func(xdr.ScVal) (V, error)) (map[K]V, error) {
	switch v.Type {
	case xdr.ScValTypeScvMap:
		result := make(map[K]V)
		for _, pair := v.MapVal.Entries; ; pair = pair.Next {
			if pair == nil {
				break
			}
			key, err := decodeKey(pair.Key)
			if err != nil {
				return nil, err
			}
			val, err := decodeValue(pair.Value)
			if err != nil {
				return nil, err
			}
			result[key] = val
		}
		return result, nil
	default:
		return nil, ErrWrongType{Expected: "map", Actual: v.Type.String()}
	}
}

// DecodeAddress decodes an ScVal containing an address value.
func DecodeAddress(v xdr.ScVal) (string, error) {
	switch v.Type {
	case xdr.ScValTypeScvAddress:
		var addr xdr.Address
		if err := xdr.SafeUnmarshalBase64(v.AddressVal, &addr); err != nil {
			return "", err
		}
		return addr.String(), nil
	default:
		return "", ErrWrongType{Expected: "address", Actual: v.Type.String()}
	}
}

// DecodeTimepoint decodes an ScVal containing a timestamp value.
func DecodeTimepoint(v xdr.ScVal) (uint64, error) {
	switch v.Type {
	case xdr.ScValTypeScvTimepoint:
		return v.TimepointVal, nil
	default:
		return 0, ErrWrongType{Expected: "timepoint", Actual: v.Type.String()}
	}
}

// DecodeDuration decodes an ScVal containing a duration value.
func DecodeDuration(v xdr.ScVal) (uint64, error) {
	switch v.Type {
	case xdr.ScValTypeScvDuration:
		return v.DurationVal, nil
	default:
		return 0, ErrWrongType{Expected: "duration", Actual: v.Type.String()}
	}
}

// ErrWrongType indicates that an ScVal was not the expected type.
type ErrWrongType struct {
	Expected string
	Actual   string
}

func (e ErrWrongType) Error() string {
	return "soroban: expected ScVal type " + e.Expected + ", got " + e.Actual
}