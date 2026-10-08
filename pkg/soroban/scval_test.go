package soroban

import (
	"errors"
	"math/big"
	"testing"

	"github.com/stellar/go/xdr"
)

func TestDecodePrimitives(t *testing.T) {
	bTrue := true
	scBool := xdr.ScVal{Type: xdr.ScValTypeScvBool, B: &bTrue}
	gotBool, err := DecodeBool(scBool)
	if err != nil || !gotBool {
		t.Errorf("DecodeBool = %v, %v, want true, nil", gotBool, err)
	}

	u32Val := xdr.Uint32(42)
	scU32 := xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: &u32Val}
	gotU32, err := DecodeU32(scU32)
	if err != nil || gotU32 != 42 {
		t.Errorf("DecodeU32 = %v, %v, want 42, nil", gotU32, err)
	}

	i32Val := xdr.Int32(-42)
	scI32 := xdr.ScVal{Type: xdr.ScValTypeScvI32, I32: &i32Val}
	gotI32, err := DecodeI32(scI32)
	if err != nil || gotI32 != -42 {
		t.Errorf("DecodeI32 = %v, %v, want -42, nil", gotI32, err)
	}

	u64Val := xdr.Uint64(18446744073709551615)
	scU64 := xdr.ScVal{Type: xdr.ScValTypeScvU64, U64: &u64Val}
	gotU64, err := DecodeU64(scU64)
	if err != nil || gotU64 != 18446744073709551615 {
		t.Errorf("DecodeU64 = %v, %v", gotU64, err)
	}

	i64Val := xdr.Int64(-9223372036854775808)
	scI64 := xdr.ScVal{Type: xdr.ScValTypeScvI64, I64: &i64Val}
	gotI64, err := DecodeI64(scI64)
	if err != nil || gotI64 != -9223372036854775808 {
		t.Errorf("DecodeI64 = %v, %v", gotI64, err)
	}

	sym := xdr.ScSymbol("TRANSFER")
	scSym := xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: &sym}
	gotSym, err := DecodeSymbol(scSym)
	if err != nil || gotSym != "TRANSFER" {
		t.Errorf("DecodeSymbol = %v, %v, want TRANSFER", gotSym, err)
	}

	str := xdr.ScString("hello world")
	scStr := xdr.ScVal{Type: xdr.ScValTypeScvString, Str: &str}
	gotStr, err := DecodeString(scStr)
	if err != nil || gotStr != "hello world" {
		t.Errorf("DecodeString = %v, %v, want hello world", gotStr, err)
	}

	bytes := xdr.ScBytes{1, 2, 3, 4}
	scBytes := xdr.ScVal{Type: xdr.ScValTypeScvBytes, Bytes: &bytes}
	gotBytes, err := DecodeBytes(scBytes)
	if err != nil || len(gotBytes) != 4 || gotBytes[0] != 1 {
		t.Errorf("DecodeBytes = %v, %v", gotBytes, err)
	}
}

func TestDecodeExactBigIntegers(t *testing.T) {
	// U128
	u128Parts := xdr.UInt128Parts{Hi: 1, Lo: 2}
	scU128 := xdr.ScVal{Type: xdr.ScValTypeScvU128, U128: &u128Parts}
	gotU128, err := DecodeU128(scU128)
	if err != nil {
		t.Fatalf("DecodeU128 err: %v", err)
	}
	wantU128 := new(big.Int).Lsh(big.NewInt(1), 64)
	wantU128.Add(wantU128, big.NewInt(2))
	if gotU128.Cmp(wantU128) != 0 {
		t.Errorf("DecodeU128 = %s, want %s", gotU128.String(), wantU128.String())
	}

	// I128 positive
	i128Pos := xdr.Int128Parts{Hi: 1, Lo: 5}
	scI128Pos := xdr.ScVal{Type: xdr.ScValTypeScvI128, I128: &i128Pos}
	gotI128Pos, err := DecodeI128(scI128Pos)
	if err != nil {
		t.Fatalf("DecodeI128 positive err: %v", err)
	}
	wantI128Pos := new(big.Int).Lsh(big.NewInt(1), 64)
	wantI128Pos.Add(wantI128Pos, big.NewInt(5))
	if gotI128Pos.Cmp(wantI128Pos) != 0 {
		t.Errorf("DecodeI128 pos = %s, want %s", gotI128Pos.String(), wantI128Pos.String())
	}

	// I128 negative: -1 is Hi = -1, Lo = 0xFFFFFFFFFFFFFFFF
	i128Neg1 := xdr.Int128Parts{Hi: xdr.Int64(-1), Lo: xdr.Uint64(^uint64(0))}
	scI128Neg1 := xdr.ScVal{Type: xdr.ScValTypeScvI128, I128: &i128Neg1}
	gotI128Neg1, err := DecodeI128(scI128Neg1)
	if err != nil {
		t.Fatalf("DecodeI128 -1 err: %v", err)
	}
	if gotI128Neg1.Cmp(big.NewInt(-1)) != 0 {
		t.Errorf("DecodeI128 -1 = %s, want -1", gotI128Neg1.String())
	}

	// I256 negative: -1 is HiHi = -1, HiLo = -1, LoHi = -1, LoLo = -1
	i256Neg1 := xdr.Int256Parts{
		HiHi: xdr.Int64(-1),
		HiLo: xdr.Uint64(^uint64(0)),
		LoHi: xdr.Uint64(^uint64(0)),
		LoLo: xdr.Uint64(^uint64(0)),
	}
	scI256Neg1 := xdr.ScVal{Type: xdr.ScValTypeScvI256, I256: &i256Neg1}
	gotI256Neg1, err := DecodeI256(scI256Neg1)
	if err != nil {
		t.Fatalf("DecodeI256 -1 err: %v", err)
	}
	if gotI256Neg1.Cmp(big.NewInt(-1)) != 0 {
		t.Errorf("DecodeI256 -1 = %s, want -1", gotI256Neg1.String())
	}
}

func TestDecodeCollections(t *testing.T) {
	// Vector of strings
	s1 := xdr.ScString("alpha")
	s2 := xdr.ScString("beta")
	vec := xdr.ScVec{
		xdr.ScVal{Type: xdr.ScValTypeScvString, Str: &s1},
		xdr.ScVal{Type: xdr.ScValTypeScvString, Str: &s2},
	}
	pVec := &vec
	scVec := xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &pVec}

	strings, err := DecodeVec(scVec, DecodeString)
	if err != nil {
		t.Fatalf("DecodeVec err: %v", err)
	}
	if len(strings) != 2 || strings[0] != "alpha" || strings[1] != "beta" {
		t.Errorf("DecodeVec = %v", strings)
	}

	// Map of Symbol -> U32
	symKey := xdr.ScSymbol("COUNTER")
	valU32 := xdr.Uint32(100)
	mapData := xdr.ScMap{
		{
			Key: xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: &symKey},
			Val: xdr.ScVal{Type: xdr.ScValTypeScvU32, U32: &valU32},
		},
	}
	pMap := &mapData
	scMap := xdr.ScVal{Type: xdr.ScValTypeScvMap, Map: &pMap}

	decodedMap, err := DecodeMap(scMap, DecodeSymbol, DecodeU32)
	if err != nil {
		t.Fatalf("DecodeMap err: %v", err)
	}
	if decodedMap["COUNTER"] != 100 {
		t.Errorf("decodedMap['COUNTER'] = %d, want 100", decodedMap["COUNTER"])
	}

	// OrderedMap
	ordered, err := DecodeOrderedMap(scMap, DecodeSymbol, DecodeU32)
	if err != nil {
		t.Fatalf("DecodeOrderedMap err: %v", err)
	}
	if len(ordered) != 1 || ordered[0].Key != "COUNTER" || ordered[0].Value != 100 {
		t.Errorf("DecodeOrderedMap = %v", ordered)
	}
}

func TestDecodeNative(t *testing.T) {
	sym := xdr.ScSymbol("MY_KEY")
	val := xdr.Int64(999)
	m := xdr.ScMap{
		{
			Key: xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: &sym},
			Val: xdr.ScVal{Type: xdr.ScValTypeScvI64, I64: &val},
		},
	}
	pm := &m
	sc := xdr.ScVal{Type: xdr.ScValTypeScvMap, Map: &pm}

	native, err := DecodeNative(sc)
	if err != nil {
		t.Fatalf("DecodeNative err: %v", err)
	}
	asMap, ok := native.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", native)
	}
	if asMap["MY_KEY"] != int64(999) {
		t.Errorf("MY_KEY = %v, want 999", asMap["MY_KEY"])
	}
}

func TestDecodeWrongType(t *testing.T) {
	sc := xdr.ScVal{Type: xdr.ScValTypeScvVoid}
	_, err := DecodeBool(sc)
	if err == nil {
		t.Fatal("expected error decoding void as bool")
	}
	var wrongType ErrWrongType
	if !errors.As(err, &wrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
	if wrongType.Expected != "bool" || wrongType.Actual != "ScValTypeScvVoid" {
		t.Errorf("wrongType = %+v, want Expected: bool, Actual: ScValTypeScvVoid", wrongType)
	}
}
