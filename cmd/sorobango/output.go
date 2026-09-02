package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/stellar/go/xdr"
)

// emit writes v as indented JSON when --json is set, and otherwise delegates to
// human. Output goes through the command so that tests can capture it.
func emit(cmd *cobra.Command, asJSON bool, v any, human func(w io.Writer)) error {
	w := cmd.OutOrStdout()

	if asJSON {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(v)
	}

	human(w)
	return nil
}

// formatScVal renders a contract value for human output. Scalars are printed
// directly; anything structured falls back to its base64 XDR, which stays
// useful input for other tooling.
func formatScVal(val xdr.ScVal) string {
	switch val.Type {
	case xdr.ScValTypeScvBool:
		if val.B != nil {
			return fmt.Sprintf("%t", *val.B)
		}
	case xdr.ScValTypeScvVoid:
		return "void"
	case xdr.ScValTypeScvU32:
		if val.U32 != nil {
			return fmt.Sprintf("%d", *val.U32)
		}
	case xdr.ScValTypeScvI32:
		if val.I32 != nil {
			return fmt.Sprintf("%d", *val.I32)
		}
	case xdr.ScValTypeScvU64:
		if val.U64 != nil {
			return fmt.Sprintf("%d", *val.U64)
		}
	case xdr.ScValTypeScvI64:
		if val.I64 != nil {
			return fmt.Sprintf("%d", *val.I64)
		}
	case xdr.ScValTypeScvSymbol:
		if val.Sym != nil {
			return string(*val.Sym)
		}
	case xdr.ScValTypeScvString:
		if val.Str != nil {
			return string(*val.Str)
		}
	}

	encoded, err := xdr.MarshalBase64(val)
	if err != nil {
		return fmt.Sprintf("<%s: could not encode: %v>", val.Type, err)
	}
	return encoded
}
