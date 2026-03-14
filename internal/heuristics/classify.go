package heuristics

// Classify assigns a transaction classification based on heuristic results.
func Classify(tc *TxContext, h *TxHeuristics) string {
	if tc.IsCoinbase {
		return "unknown"
	}
	if h.Coinjoin.Detected {
		return "coinjoin"
	}
	if h.Consolidation.Detected {
		return "consolidation"
	}
	if h.SelfTransfer.Detected {
		return "self_transfer"
	}
	nOut := len(tc.Tx.TxOut)
	nIn := len(tc.Tx.TxIn)
	if nOut > 3 && nIn >= 1 {
		return "batch_payment"
	}
	if nOut <= 2 && nIn >= 1 {
		return "simple_payment"
	}
	return "unknown"
}

// mostCommon returns the most frequent non-"unknown" item in items.
func mostCommon(items []string) string {
	counts := make(map[string]int)
	for _, s := range items {
		if s != "unknown" {
			counts[s]++
		}
	}
	best := "unknown"
	bestN := 0
	for k, v := range counts {
		if v > bestN {
			bestN = v
			best = k
		}
	}
	return best
}
