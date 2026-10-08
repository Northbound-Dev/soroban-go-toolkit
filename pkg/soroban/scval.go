package soroban

import (
	"fmt"
	"math/big"

	"github.com/stellar/go/xdr"
)

// ErrWrongType indicates that an ScVal was not of the expected type.
type ErrWrongType struct {
	Expected string
	Actual   string
}

func (e ErrWrongType) Error() string {
	return fmt.Sprintf("soroban: expected ScVal type %s, got %s", e.Expected, e.Actual)
}

// MapEntry represents a key-value pair in an ordered Soroban contract map.
type MapEntry[K any, V any] struct {
	Key   K
	Value V
}

// OrderedMap preserves the exact wire sequence of key-value pairs from an ScMap.
type OrderedMap[K any, V any] []MapEntry[K, V]

// DecodeBase64ScVal parses a base64-encoded XDR ScVal string.
func DecodeBase64ScVal(b64 string) (xdr.ScVal, error) {
	var val xdr.ScVal
	if err := xdr.SafeUnmarshalBase64(b64, &val); err != nil {
		return val, fmt.Errorf("soroban: unmarshal base64 ScVal: %w", err)
	}
	return val, nil
}

// DecodeBool decodes a boolean ScVal.
func DecodeBool(v xdr.ScVal) (bool, error) {
	if v.Type != xdr.ScValTypeScvBool || v.B == nil {
		return false, ErrWrongType{Expected: "bool", Actual: v.Type.String()}
	}
	return *v.B, nil
}

// DecodeU32 decodes an unsigned 32-bit integer ScVal.
func DecodeU32(v xdr.ScVal) (uint32, error) {
	if v.Type != xdr.ScValTypeScvU32 || v.U32 == nil {
		return 0, ErrWrongType{Expected: "u32", Actual: v.Type.String()}
	}
	return uint32(*v.U32), nil
}

// DecodeI32 decodes a signed 32-bit integer ScVal.
func DecodeI32(v xdr.ScVal) (int32, error) {
	if v.Type != xdr.ScValTypeScvI32 || v.I32 == nil {
		return 0, ErrWrongType{Expected: "i32", Actual: v.Type.String()}
	}
	return int32(*v.I32), nil
}

// DecodeU64 decodes an unsigned 64-bit integer ScVal.
func DecodeU64(v xdr.ScVal) (uint64, error) {
	if v.Type != xdr.ScValTypeScvU64 || v.U64 == nil {
		return 0, ErrWrongType{Expected: "u64", Actual: v.Type.String()}
	}
	return uint64(*v.U64), nil
}

// DecodeI64 decodes a signed 64-bit integer ScVal.
func DecodeI64(v xdr.ScVal) (int64, error) {
	if v.Type != xdr.ScValTypeScvI64 || v.I64 == nil {
		return 0, ErrWrongType{Expected: "i64", Actual: v.Type.String()}
	}
	return int64(*v.I64), nil
}

// DecodeTimepoint decodes a timepoint (unix timestamp) ScVal.
func DecodeTimepoint(v xdr.ScVal) (uint64, error) {
	if v.Type != xdr.ScValTypeScvTimepoint || v.Timepoint == nil {
		return 0, ErrWrongType{Expected: "timepoint", Actual: v.Type.String()}
	}
	return uint64(*v.Timepoint), nil
}

// DecodeDuration decodes a duration (seconds) ScVal.
func DecodeDuration(v xdr.ScVal) (uint64, error) {
	if v.Type != xdr.ScValTypeScvDuration || v.Duration == nil {
		return 0, ErrWrongType{Expected: "duration", Actual: v.Type.String()}
	}
	return uint64(*v.Duration), nil
}

// DecodeU128 decodes an unsigned 128-bit integer ScVal into an exact *big.Int.
func DecodeU128(v xdr.ScVal) (*big.Int, error) {
	if v.Type != xdr.ScValTypeScvU128 || v.U128 == nil {
		return nil, ErrWrongType{Expected: "u128", Actual: v.Type.String()}
	}
	z := new(big.Int).SetUint64(uint64(v.U128.Hi))
	z.Lsh(z, 64)
	lo := new(big.Int).SetUint64(uint64(v.U128.Lo))
	z.Or(z, lo)
	return z, nil
}

// DecodeI128 decodes a signed 128-bit integer ScVal into an exact *big.Int
// with proper two's complement sign handling.
func DecodeI128(v xdr.ScVal) (*big.Int, error) {
	if v.Type != xdr.ScValTypeScvI128 || v.I128 == nil {
		return nil, ErrWrongType{Expected: "i128", Actual: v.Type.String()}
	}
	u := new(big.Int).SetUint64(uint64(v.I128.Hi))
	u.Lsh(u, 64)
	lo := new(big.Int).SetUint64(uint64(v.I128.Lo))
	u.Or(u, lo)

	if v.I128.Hi < 0 {
		two128 := new(big.Int).Lsh(big.NewInt(1), 128)
		u.Sub(u, two128)
	}
	return u, nil
}

// DecodeU256 decodes an unsigned 256-bit integer ScVal into an exact *big.Int.
func DecodeU256(v xdr.ScVal) (*big.Int, error) {
	if v.Type != xdr.ScValTypeScvU256 || v.U256 == nil {
		return nil, ErrWrongType{Expected: "u256", Actual: v.Type.String()}
	}
	z := new(big.Int).SetUint64(uint64(v.U256.HiHi))
	z.Lsh(z, 64).Or(z, new(big.Int).SetUint64(uint64(v.U256.HiLo)))
	z.Lsh(z, 64).Or(z, new(big.Int).SetUint64(uint64(v.U256.LoHi)))
	z.Lsh(z, 64).Or(z, new(big.Int).SetUint64(uint64(v.U256.LoLo)))
	return z, nil
}

// DecodeI256 decodes a signed 256-bit integer ScVal into an exact *big.Int
// with proper two's complement sign handling.
func DecodeI256(v xdr.ScVal) (*big.Int, error) {
	if v.Type != xdr.ScValTypeScvI256 || v.I256 == nil {
		return nil, ErrWrongType{Expected: "i256", Actual: v.Type.String()}
	}
	u := new(big.Int).SetUint64(uint64(v.I256.HiHi))
	u.Lsh(u, 64).Or(u, new(big.Int).SetUint64(uint64(v.I256.HiLo)))
	u.Lsh(u, 64).Or(u, new(big.Int).SetUint64(uint64(v.I256.LoHi)))
	u.Lsh(u, 64).Or(u, new(big.Int).SetUint64(uint64(v.I256.LoLo)))

	if v.I256.HiHi < 0 {
		two256 := new(big.Int).Lsh(big.NewInt(1), 256)
		u.Sub(u, two256)
	}
	return u, nil
}

// DecodeString decodes a string ScVal.
func DecodeString(v xdr.ScVal) (string, error) {
	if v.Type != xdr.ScValTypeScvString || v.Str == nil {
		return "", ErrWrongType{Expected: "string", Actual: v.Type.String()}
	}
	return string(*v.Str), nil
}

// DecodeSymbol decodes a symbol ScVal.
func DecodeSymbol(v xdr.ScVal) (string, error) {
	if v.Type != xdr.ScValTypeScvSymbol || v.Sym == nil {
		return "", ErrWrongType{Expected: "symbol", Actual: v.Type.String()}
	}
	return string(*v.Sym), nil
}

// DecodeBytes decodes a byte slice ScVal.
func DecodeBytes(v xdr.ScVal) ([]byte, error) {
	if v.Type != xdr.ScValTypeScvBytes || v.Bytes == nil {
		return nil, ErrWrongType{Expected: "bytes", Actual: v.Type.String()}
	}
	res := make([]byte, len(*v.Bytes))
	copy(res, *v.Bytes)
	return res, nil
}

// DecodeAddress decodes a contract or account address ScVal into a strkey string (e.g. C... or G...).
func DecodeAddress(v xdr.ScVal) (string, error) {
	if v.Type != xdr.ScValTypeScvAddress || v.Address == nil {
		return "", ErrWrongType{Expected: "address", Actual: v.Type.String()}
	}
	str, err := v.Address.String()
	if err != nil {
		return "", fmt.Errorf("soroban: format address: %w", err)
	}
	return str, nil
}

// DecodeVec decodes a vector ScVal into a slice of type T using the supplied element decoder.
func DecodeVec[T any](v xdr.ScVal, decodeElem func(xdr.ScVal) (T, error)) ([]T, error) {
	if v.Type != xdr.ScValTypeScvVec || v.Vec == nil || *v.Vec == nil {
		return nil, ErrWrongType{Expected: "vec", Actual: v.Type.String()}
	}
	rawVec := **v.Vec
	res := make([]T, len(rawVec))
	for i, elem := range rawVec {
		val, err := decodeElem(elem)
		if err != nil {
			return nil, fmt.Errorf("soroban: decode vec element [%d]: %w", i, err)
		}
		res[i] = val
	}
	return res, nil
}

// DecodeMap decodes a map ScVal into a Go map using key and value decoders.
// If the wire representation contains duplicate keys, later values overwrite earlier ones.
func DecodeMap[K comparable, V any](v xdr.ScVal, decodeKey func(xdr.ScVal) (K, error), decodeVal func(xdr.ScVal) (V, error)) (map[K]V, error) {
	if v.Type != xdr.ScValTypeScvMap || v.Map == nil || *v.Map == nil {
		return nil, ErrWrongType{Expected: "map", Actual: v.Type.String()}
	}
	rawMap := **v.Map
	res := make(map[K]V, len(rawMap))
	for i, entry := range rawMap {
		k, err := decodeKey(entry.Key)
		if err != nil {
			return nil, fmt.Errorf("soroban: decode map key [%d]: %w", i, err)
		}
		val, err := decodeVal(entry.Val)
		if err != nil {
			return nil, fmt.Errorf("soroban: decode map value for key [%d]: %w", i, err)
		}
		res[k] = val
	}
	return res, nil
}

// DecodeOrderedMap decodes a map ScVal preserving the exact sequence of wire entries.
func DecodeOrderedMap[K any, V any](v xdr.ScVal, decodeKey func(xdr.ScVal) (K, error), decodeVal func(xdr.ScVal) (V, error)) (OrderedMap[K, V], error) {
	if v.Type != xdr.ScValTypeScvMap || v.Map == nil || *v.Map == nil {
		return nil, ErrWrongType{Expected: "map", Actual: v.Type.String()}
	}
	rawMap := **v.Map
	res := make(OrderedMap[K, V], len(rawMap))
	for i, entry := range rawMap {
		k, err := decodeKey(entry.Key)
		if err != nil {
			return nil, fmt.Errorf("soroban: decode ordered map key [%d]: %w", i, err)
		}
		val, err := decodeVal(entry.Val)
		if err != nil {
			return nil, fmt.Errorf("soroban: decode ordered map val [%d]: %w", i, err)
		}
		res[i] = MapEntry[K, V]{Key: k, Value: val}
	}
	return res, nil
}

// DecodeNative translates any ScVal into its natural Go representation recursively:
//   - bool -> bool
//   - void -> nil
//   - u32 -> uint32
//   - i32 -> int32
//   - u64, timepoint, duration -> uint64
//   - i64 -> int64
//   - u128, i128, u256, i256 -> *big.Int
//   - string, symbol -> string
//   - bytes -> []byte
//   - address -> string (strkey)
//   - vec -> []any
//   - map -> map[string]any (if keys decode to string) or OrderedMap[any, any]
func DecodeNative(v xdr.ScVal) (any, error) {
	switch v.Type {
	case xdr.ScValTypeScvBool:
		return DecodeBool(v)
	case xdr.ScValTypeScvVoid:
		return nil, nil
	case xdr.ScValTypeScvU32:
		return DecodeU32(v)
	case xdr.ScValTypeScvI32:
		return DecodeI32(v)
	case xdr.ScValTypeScvU64:
		return DecodeU64(v)
	case xdr.ScValTypeScvI64:
		return DecodeI64(v)
	case xdr.ScValTypeScvTimepoint:
		return DecodeTimepoint(v)
	case xdr.ScValTypeScvDuration:
		return DecodeDuration(v)
	case xdr.ScValTypeScvU128:
		return DecodeU128(v)
	case xdr.ScValTypeScvI128:
		return DecodeI128(v)
	case xdr.ScValTypeScvU256:
		return DecodeU256(v)
	case xdr.ScValTypeScvI256:
		return DecodeI256(v)
	case xdr.ScValTypeScvString:
		return DecodeString(v)
	case xdr.ScValTypeScvSymbol:
		return DecodeSymbol(v)
	case xdr.ScValTypeScvBytes:
		return DecodeBytes(v)
	case xdr.ScValTypeScvAddress:
		return DecodeAddress(v)
	case xdr.ScValTypeScvVec:
		return DecodeVec(v, DecodeNative)
	case xdr.ScValTypeScvMap:
		// Attempt string-keyed map first
		strMap, err := DecodeMap(v, func(k xdr.ScVal) (string, error) {
			if k.Type == xdr.ScValTypeScvSymbol {
				return DecodeSymbol(k)
			}
			if k.Type == xdr.ScValTypeScvString {
				return DecodeString(k)
			}
			return "", ErrWrongType{Expected: "symbol or string", Actual: k.Type.String()}
		}, DecodeNative)
		if err == nil {
			return strMap, nil
		}
		// Fallback to ordered map of any
		return DecodeOrderedMap(v, DecodeNative, DecodeNative)
	default:
		return nil, fmt.Errorf("soroban: unsupported ScVal type for native decoding: %s", v.Type.String())
	}
}
