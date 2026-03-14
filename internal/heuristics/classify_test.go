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

func TestClassify_SimplePayment(t *testing.T) {
	tc := makeTx(1, 2)
	h := &TxHeuristics{}
	result := Classify(tc, h)
	if result != "simple_payment" {
		t.Errorf("Classify simple = %q, want simple_payment", result)
	}
}

func TestClassify_Coinbase(t *testing.T) {
	tc := makeTx(1, 1)
	tc.IsCoinbase = true
	h := &TxHeuristics{}
	if got := Classify(tc, h); got != "unknown" {
		t.Errorf("Classify coinbase = %q, want unknown", got)
	}
}

func TestClassify_SelfTransfer(t *testing.T) {
	tc := makeTx(1, 1)
	h := &TxHeuristics{SelfTransfer: HeuristicResult{Detected: true}}
	if got := Classify(tc, h); got != "self_transfer" {
		t.Errorf("Classify self_transfer = %q, want self_transfer", got)
	}
}

func TestClassify_Unknown(t *testing.T) {
	tc := makeTx(1, 3)
	h := &TxHeuristics{}
	if got := Classify(tc, h); got != "unknown" {
		t.Errorf("Classify 3-output = %q, want unknown", got)
	}
}
