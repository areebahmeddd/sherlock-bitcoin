package heuristics

// ApplyChangeDetection tries to identify the change output.
func ApplyChangeDetection(tc *TxContext) HeuristicResult {
	if tc.IsCoinbase || len(tc.Tx.TxOut) < 2 {
		return HeuristicResult{Detected: false}
	}

	inTypes := tc.inputScriptTypes()
	outTypes := tc.outputScriptTypes()

	dominantInput := mostCommon(inTypes)
	if dominantInput == "unknown" || dominantInput == "op_return" {
		return HeuristicResult{Detected: false}
	}

	// method 1: change output matches the dominant input script type
	changeIdx := -1
	method := ""
	confidence := "low"

	for i, ot := range outTypes {
		if ot == dominantInput {
			differentCount := 0
			for j, ot2 := range outTypes {
				if j != i && ot2 != dominantInput && ot2 != "op_return" {
					differentCount++
				}
			}
			if differentCount > 0 {
				changeIdx = i
				method = "script_type_match"
				confidence = "high"
				break
			}
		}
	}

	// method 2: single non-round output is likely change
	if changeIdx < 0 {
		nonRoundIdx := -1
		nonRoundCount := 0
		for i, out := range tc.Tx.TxOut {
			if outTypes[i] != "op_return" && !isRoundAmount(out.Value) {
				nonRoundCount++
				nonRoundIdx = i
			}
		}
		if nonRoundCount == 1 {
			changeIdx = nonRoundIdx
			method = "round_number"
			confidence = "medium"
		}
	}

	// method 3: smaller of two outputs is likely change
	if changeIdx < 0 && len(tc.Tx.TxOut) == 2 {
		idx := 0
		if tc.Tx.TxOut[1].Value < tc.Tx.TxOut[0].Value {
			idx = 1
		}
		changeIdx = idx
		method = "smallest_output"
		confidence = "low"
	}

	if changeIdx < 0 {
		return HeuristicResult{Detected: false}
	}

	return HeuristicResult{
		Detected: true,
		Extra: map[string]interface{}{
			"likely_change_index": changeIdx,
			"method":              method,
			"confidence":          confidence,
		},
	}
}
