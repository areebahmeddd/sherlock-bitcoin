package heuristics

import "testing"

func TestApplyCIOH_SingleInput(t *testing.T) {
	tc := makeTx(1, 2)
	if got := ApplyCIOH(tc).Detected; got {
		t.Error("CIOH: single-input tx should NOT be flagged")
	}
}

func TestApplyCIOH_MultiInput(t *testing.T) {
	tc := makeTx(3, 2)
	if got := ApplyCIOH(tc).Detected; !got {
		t.Error("CIOH: multi-input tx should be flagged")
	}
}

func TestApplyCIOH_Coinbase(t *testing.T) {
	tc := makeTx(2, 1)
	tc.IsCoinbase = true
	if got := ApplyCIOH(tc).Detected; got {
		t.Error("CIOH: coinbase should never be flagged")
	}
}
