package soroban

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stellar/go/xdr"
)

func TestDecodeString(t *testing.T) {
	tests := []struct {
		name    string
		val     xdr.ScVal
		want    string
		wantErr bool
	}{
		{
			name: "valid string",
			val: xdr.ScVal{
				Type:   xdr.ScValTypeScvString,
				StringVal: xdr.StringVal("hello"),
			},
			want:    "hello",
			wantErr: false,
		},
		{
			name: "empty string",
			val: xdr.ScVal{
				Type:   xdr.ScValTypeScvString,
				StringVal: xdr.StringVal(""),
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "wrong type",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvBool,
				BoolVal: true,
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeString(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeBytes(t *testing.T) {
	tests := []struct {
		name    string
		val     xdr.ScVal
		want    []byte
		wantErr bool
	}{
		{
			name: "valid bytes",
			val: xdr.ScVal{
				Type:  xdr.ScValTypeScvBytes,
				BytesVal: xdr.ScVarBytes{[]byte{1, 2, 3}},
			},
			want:    []byte{1, 2, 3},
			wantErr: false,
		},
		{
			name: "empty bytes",
			val: xdr.ScVal{
				Type:  xdr.ScValTypeScvBytes,
				BytesVal: xdr.ScVarBytes{},
			},
			want:    []byte{},
			wantErr: false,
		},
		{
			name: "wrong type",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvString,
				StringVal: xdr.StringVal("hello"),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeBytes(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeI32(t *testing.T) {
	tests := []struct {
		name    string
		val     xdr.ScVal
		want    int32
		wantErr bool
	}{
		{
			name: "positive i32",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI32,
				I32Val: 42,
			},
			want:    42,
			wantErr: false,
		},
		{
			name: "negative i32",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI32,
				I32Val: -42,
			},
			want:    -42,
			wantErr: false,
		},
		{
			name: "zero i32",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI32,
				I32Val: 0,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "wrong type",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI64,
				I64Val: 42,
            },
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeI32(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeU32(t *testing.T) {
	tests := []struct {
		name    string
		val     xdr.ScVal
		want    uint32
		wantErr bool
	}{
		{
			name: "valid u32",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU32,
				U32Val: 42,
			},
			want:    42,
			wantErr: false,
		},
		{
			name: "zero u32",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU32,
				U32Val: 0,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "max u32",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU32,
				U32Val: ^uint32(0),
			},
			want:    ^uint32(0),
			wantErr: false,
		},
		{
			name: "wrong type",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU64,
				4,
				U64Val: 42,
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeU32(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeI64(t *testing.T) {
	tests := []struct {
		name    string
		val     xdr.ScVal
		want    int64
		wantErr bool
	}{
		{
			name: "valid i64",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI64,
				I64Val: 42,
			},
			want:    42,
			wantErr: false,
		},
		{
			name: "negative i64",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI64,
				I64Val: -42,
			},
			want:    -42,
			wantErr: false,
		},
		{
			name: "zero i64",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI64,
				I64Val: 0,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "wrong type",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI32,
				I32Val: 42,
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeI64(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeU64(t *testing.T) {
	tests := []struct {
		name    string
		val     xdr.ScVal
		want    uint64
		wantErr bool
	}{
		{
			name: "valid u64",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU64,
				U64Val: 42,
			},
			want:    42,
			wantErr: false,
		},
		{
			name: "zero u64",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU64,
				U64Val: 0,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "max u64",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU64,
				U64Val: ^uint64(0),
			},
			want:    ^uint64(0),
			wantErr: false,
		},
		{
			name: "wrong type",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU32,
				U32Val: 42,
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeU64(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeI128(t *testing.T) {
	tests := []struct {
		name    string
		val     xdr.ScVal
		want    *big.Int
		wantErr bool
	}{
		{
			name: "positive i128",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI128,
				U128Val: xdr.Uint128{
					Hi: 0,
					Lo: 42,
				},
			},
			want:    big.NewInt(42),
			wantErr: false,
		},
		{
			name: "zero i128",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI128,
				U128Val: xdr.Uint128{
					Hi: 0,
					Lo: 0,
				},
			},
			want:    big.NewInt(0),
			wantErr: false,
		},
		{
			name: "negative i128",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI128,
				U128Val: xdr.Uint128{
					Hi: ^uint64(0),
					Lo: ^uint64(0) - 41, // -42 in two's complement
				},
			},
			want:    big.NewInt(-42),
			wantErr: false,
		},
		{
			name: "wrong type",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI64,
				I64Val: 42,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeI128(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want.String(), got.String())
		})
	}
}

func TestDecodeU128(t *testing.T) {
	tests := []struct {
		name    string
		val     xdr.ScVal
		want    *big.Int
		wantErr bool
	}{
		{
			name: "valid u128",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU128,
				U128Val: xdr.Uint128{
					Hi: 0,
					Lo: 42,
				},
			},
			want:    big.NewInt(42),
			wantErr: false,
		},
		{
			name: "zero u128",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU128,
				U128Val: xdr.Uint128{
					Hi: 0,
					Lo: 0,
				},
			},
			want:    big.NewInt(0),
			wantErr: false,
		},
		{
			name: "max u128",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU128,
				U128Val: xdr.Uint128{
					Hi: ^uint64(0),
					Lo: ^uint64(0),
				},
			},
			want:    new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)), // 2^128 - 1
			wantErr: false,
		},
		{
			name: "wrong type",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU64,
				U64Val: 42,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeU128(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want.String(), got.String())
		})
	}
}

func TestDecodeI256(t *testing.T) {
	tests := []struct {
		name    string
		val     xdr.ScVal
		want    *big.Int
		wantErr bool
	}{
		{
			name: "positive i256",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI256,
				U256Val: xdr.Uint256{
					Hi: [2]uint64{0, 0},
					Lo: [2]uint64{0, 42},
				},
			},
			want:    big.NewInt(42),
			wantErr: false,
		},
		{
			name: "zero i256",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI256,
				U256Val: xdr.Uint256{
					Hi: [2]uint64{0, 0},
					Lo: [2]uint64{0, 0},
				},
			},
			want:    big.NewInt(0),
			wantErr: false,
		},
		{
			name: "negative i256",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI256,
				U256Val: xdr.Uint256{
					Hi: [2]uint64{^uint64(0), ^uint64(0)},
					Lo: [2]uint64{^uint64(0), ^uint64(0) - 41}, // -42 in two's complement
				},
			},
			want:    big.NewInt(-42),
			wantErr: false,
		},
		{
			name: "wrong type",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvI64,
				I64Val: 42,
            },
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeI256(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want.String(), got.String())
		})
	}
}

func TestDecodeU256(t *testing.T) {
	tests := []struct {
		name    string
		val     xdr.ScVal
		want    *big.Int
		wantErr bool
	}{
		{
			name: "valid u256",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU256,
				U256Val: xdr.Uint256{
					Hi: [2]uint64{0, 0},
					Lo: [2]uint64{0, 42},
				},
			},
			want:    big.NewInt(42),
			wantErr: false,
		},
		{
			name: "zero u256",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU256,
				U256Val: xdr.Uint256{
					Hi: [2]uint64{0, 0},
					Lo: [2]uint64{0, 0},
				},
			},
			want:    big.NewInt(0),
			wantErr: false,
		},
		{
			name: "max u256",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU256,
				U256Val: xdr.Uint256{
					Hi: [2]uint64{^uint64(0), ^uint64(0)},
					Lo: [2]uint64{^uint64(0), ^uint64(0)},
				},
			},
			want:    new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)), // 2^256 - 1
			wantErr: false,
		},
		{
			name: "wrong type",
			val: xdr.ScVal{
				Type: xdr.ScValTypeScvU64,
				U64Val: 42,
            },
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeU256(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want.String(), got.String())
		})
	}
}

func TestDecodeVecOfString(t *testing.T) {
	val := xdr.ScVal{
		Type: xdr.ScValTypeScvVec,
		VecVal: xdr.ScVec{
			Elements: []*xdr.ScVal{
				{
					Type:  xdr.ScValTypeScvString,
					StringVal: xdr.StringVal("hello"),
				},
				{
					Type:  xdr.ScValTypeScvString,
					StringVal: xdr.StringVal("world"),
				},
			},
		},
	}

	decodeElem := func(v xdr.ScVal) (string, error) {
		return DecodeString(v)
	}

	got, err := DecodeVec(val, decodeElem)
	require.NoError(t, err)
	require.Equal(t, []string{"hello", "world"}, got)
}

func TestDecodeVecWrongType(t *testing.T) {
	val := xdr.ScVal{
		Type: xdr.ScValTypeScvString,
		StringVal: xdr.StringVal("not a vec"),
	}

	decodeElem := func(v xdr.ScVal) (string, error) {
		return DecodeString(v)
	}

	_, err := DecodeVec(val, decodeElem)
	require.Error(t, err)
}

func TestDecodeMapStringToString(t *testing.T) {
	val := xdr.ScVal{
		Type: xdr.ScValTypeScvMap,
		MapVal: xdr.ScMap{
			Entries: &xdr.ScMapEntry{
				Key: &xdr.ScVal{
					Type:  xdr.ScValTypeScvString,
					StringVal: xdr.StringVal("key1"),
				},
				Value: &xdr.ScVal{
					Type:  xdr.ScValTypeScvString,
					StringVal: xdr.StringVal("value1"),
				},
				Next: &xdr.ScMapEntry{
					Key: &xdr.ScVal{
						Type:  xdr.ScValTypeScvString,
						StringVal: xdr.StringVal("key2"),
					},
					Value: &xdr.ScVal{
						Type:  xdr.ScValTypeScvString,
						StringVal: xdr.StringVal("value2"),
					},
					Next: nil,
				},
			},
		},
	}

	decodeKey := func(v xdr.ScVal) (string, error) {
		return DecodeString(v)
	}
	decodeValue := func(v xdr.ScVal) (string, error) {
		return DecodeString(v)
	}

	got, err := DecodeMap(val, decodeKey, decodeValue)
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"key1": "value1",
		"key2": "value2",
	}, got)
}

func TestDecodeMapWrongType(t *testing.T) {
	val := xdr.ScVal{
		Type: xdr.ScValTypeScvString,
		StringVal: xdr.StringVal("not a map"),
	}

	decodeKey := func(v xdr.ScVal) (string, error) {
		return DecodeString(v)
	}
	decodeValue := func(v xdr.ScVal) (string, error) {
		return DecodeString(v)
	}

	_, err := DecodeMap(val, decodeKey, decodeValue)
	require.Error(t, err)
}