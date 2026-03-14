package heuristics

import "testing"

func TestApplyConsolidation_ManyInputsSingleOutput(t *testing.T) {
	tc := makeTx(5, 1)
	if got := ApplyConsolidation(tc).Detected; !got {
		t.Error("Consolidation: many-input single-output should be flagged")
	}
}

func TestApplyConsolidation_SingleInputManyOutputs(t *testing.T) {
	tc := makeTx(1, 5)
	if got := ApplyConsolidation(tc).Detected; got {
		t.Error("Consolidation: single-input many-output should NOT be flagged")
	}
}

func TestApplyConsolidation_Coinbase(t *testing.T) {
	tc := makeTx(5, 1)
	tc.IsCoinbase = true
	if got := ApplyConsolidation(tc).Detected; got {
		t.Error("Consolidation: coinbase should NOT be flagged")
	}
}
