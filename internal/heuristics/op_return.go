package heuristics

import "bytes"

// ApplyOPReturn detects OP_RETURN outputs and classifies the embedded protocol.
// Coinbase transactions are skipped (their OP_RETURN is the segwit commitment).
func ApplyOPReturn(tc *TxContext) HeuristicResult {
	if tc.IsCoinbase {
		return HeuristicResult{Detected: false}
	}
	for _, out := range tc.Tx.TxOut {
		if len(out.PkScript) > 0 && out.PkScript[0] == 0x6a { // OP_RETURN
			protocol := classifyOPReturn(out.PkScript)
			return HeuristicResult{
				Detected: true,
				Extra: map[string]interface{}{
					"protocol": protocol,
				},
			}
		}
	}
	return HeuristicResult{Detected: false}
}

func classifyOPReturn(script []byte) string {
	if len(script) < 2 {
		return "unknown"
	}
	data := script[1:]
	if len(data) > 0 && data[0] <= 0x4b {
		data = data[1:]
	}

	if len(data) >= 4 {
		prefix := data[:4]
		switch {
		case bytes.Equal(prefix, []byte{0x6f, 0x6d, 0x6e, 0x69}): // "omni"
			return "omni"
		case bytes.Equal(prefix, []byte{0x52, 0x55, 0x4e, 0x45}): // "RUNE"
			return "runes"
		case bytes.Equal(prefix, []byte{0x53, 0x50, 0x4b, 0x42}): // "SPKB"
			return "spos"
		}
	}
	if len(data) >= 2 {
		switch {
		case data[0] == 0x45 && data[1] == 0x54: // "ET"
			return "opentimestamps"
		}
	}
	return "data"
}
