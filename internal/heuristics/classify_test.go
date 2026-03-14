package heuristics

import "testing"

func TestClassify_Coinjoin(t *testing.T) {
	tc := makeTx(3, 3)
	h := &TxHeuristics{Coinjoin: HeuristicResult{Detected: true}}
	h.Classification = Classify(tc, h)
	if h.Classification != "coinjoin" {
		t.Errorf("Classify coinjoin = %q, want coinjoin", h.Classification)
	}
}

func TestClassify_Consolidation(t *testing.T) {
	tc := makeTx(5, 1)
	h := &TxHeuristics{Consolidation: HeuristicResult{Detected: true}}
	h.Classification = Classify(tc, h)
	if h.Classification != "consolidation" {
		t.Errorf("Classify consolidation = %q, want consolidation", h.Classification)
	}
}

func TestClassify_BatchPayment(t *testing.T) {
	tc := makeTx(1, 5)
	h := &TxHeuristics{}
	result := Classify(tc, h)
	if result != "batch_payment" {
		t.Errorf("Classify batch = %q, want batch_payment", result)
	}
}
