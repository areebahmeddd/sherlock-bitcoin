package heuristics

// ApplySelfTransfer detects transactions where all outputs match the input script type.
func ApplySelfTransfer(tc *TxContext) HeuristicResult {
	if tc.IsCoinbase || len(tc.Tx.TxIn) == 0 || len(tc.Tx.TxOut) == 0 {
		return HeuristicResult{Detected: false}
	}

	inTypes := tc.inputScriptTypes()
	outTypes := tc.outputScriptTypes()

	dominant := mostCommon(inTypes)
	if dominant == "unknown" {
		return HeuristicResult{Detected: false}
	}

	for _, ot := range outTypes {
		if ot != dominant && ot != "op_return" {
			return HeuristicResult{Detected: false}
		}
	}

	// consolidations (many inputs) are classified separately
	if len(tc.Tx.TxIn) > 2 {
		return HeuristicResult{Detected: false}
	}

	return HeuristicResult{Detected: true}
}
