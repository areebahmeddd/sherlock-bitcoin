package heuristics

// ApplyCIOH detects multi-input transactions (common input ownership heuristic).
func ApplyCIOH(tc *TxContext) HeuristicResult {
	if tc.IsCoinbase {
		return HeuristicResult{Detected: false}
	}
	return HeuristicResult{Detected: len(tc.Tx.TxIn) > 1}
}
