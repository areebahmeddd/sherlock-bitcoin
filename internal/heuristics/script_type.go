package heuristics

import (
	"bytes"
	"io"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// ScriptType returns a canonical script type name for a scriptPubKey.
func ScriptType(script []byte) string {
	if len(script) == 0 {
		return "unknown"
	}
	class := txscript.GetScriptClass(script)
	switch class {
	case txscript.PubKeyHashTy:
		return "p2pkh"
	case txscript.PubKeyTy:
		return "p2pk"
	case txscript.ScriptHashTy:
		return "p2sh"
	case txscript.WitnessV0PubKeyHashTy:
		return "p2wpkh"
	case txscript.WitnessV0ScriptHashTy:
		return "p2wsh"
	case txscript.WitnessV1TaprootTy:
		return "p2tr"
	case txscript.NullDataTy:
		return "op_return"
	case txscript.MultiSigTy:
		return "multisig"
	default:
		return "unknown"
	}
}

// inferInputScriptType infers the scriptPubKey type from scriptSig and witness data.
func inferInputScriptType(in *wire.TxIn) string {
	ss := in.SignatureScript
	wit := in.Witness

	if len(wit) > 0 && len(ss) == 0 {
		switch {
		case len(wit) == 1 && len(wit[0]) == 64:
			return "p2tr"
		case len(wit) == 2:
			if len(wit[1]) == 33 || len(wit[1]) == 65 {
				return "p2wpkh"
			}
		case len(wit) >= 2:
			last := wit[len(wit)-1]
			if len(last) > 2 && last[0] != 0x50 { // not annex
				return "p2wsh"
			}
			return "p2tr"
		}
		return "p2wpkh"
	}

	if len(ss) > 0 && len(wit) > 0 {
		return "p2sh" // P2SH-wrapped segwit
	}

	if len(ss) > 0 {
		reader := bytes.NewReader(ss)
		items, err := parseScriptPushes(reader)
		if err == nil {
			switch len(items) {
			case 1:
				return "p2pkh" // sig only or P2SH
			case 2:
				// sig + pubkey
				if len(items[1]) == 33 || len(items[1]) == 65 {
					return "p2pkh"
				}
				return "p2sh"
			default:
				return "p2sh"
			}
		}
		return "p2pkh"
	}

	return "unknown"
}

// parseScriptPushes parses a scriptSig into its push items.
func parseScriptPushes(r *bytes.Reader) ([][]byte, error) {
	var items [][]byte
	for r.Len() > 0 {
		b, err := r.ReadByte()
		if err != nil {
			break
		}
		var data []byte
		switch {
		case b == 0x00:
			data = []byte{}
		case b >= 0x01 && b <= 0x4B:
			data = make([]byte, b)
			if _, err := io.ReadFull(r, data); err != nil {
				return items, nil
			}
		case b == 0x4C:
			ln, err := r.ReadByte()
			if err != nil {
				return items, nil
			}
			data = make([]byte, ln)
			if _, err := io.ReadFull(r, data); err != nil {
				return items, nil
			}
		case b == 0x4D:
			var ln [2]byte
			if _, err := io.ReadFull(r, ln[:]); err != nil {
				return items, nil
			}
			sz := int(ln[0]) | int(ln[1])<<8
			data = make([]byte, sz)
			if _, err := io.ReadFull(r, data); err != nil {
				return items, nil
			}
		default:
			// opcode, skip
			continue
		}
		items = append(items, data)
	}
	return items, nil
}
