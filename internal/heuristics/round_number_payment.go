package heuristics

// ApplyRoundNumberPayment detects round-number output amounts.
func ApplyRoundNumberPayment(tc *TxContext) HeuristicResult {
	if tc.IsCoinbase {
		return HeuristicResult{Detected: false}
	}
	for _, out := range tc.Tx.TxOut {
		if out.Value > 0 && isRoundAmount(out.Value) {
			return HeuristicResult{Detected: true}
		}
	}
	return HeuristicResult{Detected: false}
}

// isRoundAmount reports whether sats is a multiple of 1000 (0.00001 BTC).
func isRoundAmount(sats int64) bool {
	if sats <= 0 {
		return false
	}
	return sats%1000 == 0
}
