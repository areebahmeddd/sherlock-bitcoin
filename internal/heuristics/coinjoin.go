package heuristics

// ApplyCoinjoin detects CoinJoin-like transactions: many inputs, equal-value outputs.
func ApplyCoinjoin(tc *TxContext) HeuristicResult {
	if tc.IsCoinbase || len(tc.Tx.TxIn) < 3 || len(tc.Tx.TxOut) < 3 {
		return HeuristicResult{Detected: false}
	}

	valueCounts := make(map[int64]int)
	for _, out := range tc.Tx.TxOut {
		if out.Value > 0 {
			valueCounts[out.Value]++
		}
	}

	maxSameValue := 0
	for _, cnt := range valueCounts {
		if cnt > maxSameValue {
			maxSameValue = cnt
		}
	}

	if maxSameValue < 2 {
		return HeuristicResult{Detected: false}
	}

	return HeuristicResult{Detected: maxSameValue >= 2}
}
