package heuristics

// ApplyPeelingChain detects 1-in/2-out transactions where one output is ≥10× the other.
func ApplyPeelingChain(tc *TxContext) HeuristicResult {
	if tc.IsCoinbase || len(tc.Tx.TxIn) != 1 || len(tc.Tx.TxOut) != 2 {
		return HeuristicResult{Detected: false}
	}

	v0 := tc.Tx.TxOut[0].Value
	v1 := tc.Tx.TxOut[1].Value

	if v0 <= 0 || v1 <= 0 {
		return HeuristicResult{Detected: false}
	}

	larger := v0
	smaller := v1
	if v1 > v0 {
		larger = v1
		smaller = v0
	}

	detected := larger >= smaller*10
	return HeuristicResult{Detected: detected}
}
