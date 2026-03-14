package heuristics

// ApplyConsolidation detects transactions with many inputs and few outputs of the same type.
func ApplyConsolidation(tc *TxContext) HeuristicResult {
	if tc.IsCoinbase {
		return HeuristicResult{Detected: false}
	}

	nIn := len(tc.Tx.TxIn)
	nOut := len(tc.Tx.TxOut)

	if nIn < 3 || nOut > 2 {
		return HeuristicResult{Detected: false}
	}

	inTypes := tc.inputScriptTypes()
	outTypes := tc.outputScriptTypes()

	dominant := mostCommon(inTypes)
	if dominant == "unknown" {
		return HeuristicResult{Detected: false}
	}

	inputMatch := 0
	for _, t := range inTypes {
		if t == dominant {
			inputMatch++
		}
	}
	outputMatch := 0
	for _, t := range outTypes {
		if t == dominant || t == "op_return" {
			outputMatch++
		}
	}

	detected := inputMatch >= nIn*2/3 && outputMatch == nOut
	return HeuristicResult{Detected: detected}
}
