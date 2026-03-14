package heuristics

import "encoding/hex"

// ApplyAddressReuse detects when a prevout script matches an output script in the same tx.
func ApplyAddressReuse(tc *TxContext) HeuristicResult {
	if tc.IsCoinbase {
		return HeuristicResult{Detected: false}
	}

	outScripts := make(map[string]struct{})
	for _, out := range tc.Tx.TxOut {
		if len(out.PkScript) > 0 {
			outScripts[hex.EncodeToString(out.PkScript)] = struct{}{}
		}
	}

	for _, po := range tc.PrevOuts {
		if len(po.ScriptPubKey) == 0 {
			continue
		}
		k := hex.EncodeToString(po.ScriptPubKey)
		if _, ok := outScripts[k]; ok {
			return HeuristicResult{Detected: true}
		}
	}

	return HeuristicResult{Detected: false}
}
