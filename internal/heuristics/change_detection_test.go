package heuristics

import (
	"testing"

	"github.com/btcsuite/btcd/wire"

	"sherlock/internal/blockfile"
)

func TestApplyChangeDetection_ScriptTypeMatch(t *testing.T) {
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(40_000, p2wpkhScript(10)))
	tx.AddTxOut(wire.NewTxOut(50_000, p2pkhScript()))

	tc := &TxContext{
		Tx:       tx,
		PrevOuts: []blockfile.PrevOut{{Value: 100_000, ScriptPubKey: p2wpkhScript(0)}},
	}
	r := ApplyChangeDetection(tc)
	if !r.Detected {
		t.Error("ChangeDetection: should be detected via script_type_match")
	}
	if r.Extra["method"] != "script_type_match" {
		t.Errorf("method = %v, want script_type_match", r.Extra["method"])
	}
	if r.Extra["confidence"] != "high" {
		t.Errorf("confidence = %v, want high", r.Extra["confidence"])
	}
}

func TestApplyChangeDetection_SingleOutput(t *testing.T) {
	tc := makeTx(1, 1)
	if got := ApplyChangeDetection(tc).Detected; got {
		t.Error("ChangeDetection: single-output tx should NOT be flagged")
	}
}

func TestApplyChangeDetection_Coinbase(t *testing.T) {
	tc := makeTx(1, 2)
	tc.IsCoinbase = true
	if got := ApplyChangeDetection(tc).Detected; got {
		t.Error("ChangeDetection: coinbase should NOT be flagged")
	}
}

func TestApplyChangeDetection_RoundNumber(t *testing.T) {
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(1_000_000, p2wpkhScript(10)))
	tx.AddTxOut(wire.NewTxOut(49_999, p2wpkhScript(11)))

	tc := &TxContext{
		Tx:       tx,
		PrevOuts: []blockfile.PrevOut{{Value: 1_060_000, ScriptPubKey: p2wpkhScript(0)}},
	}
	r := ApplyChangeDetection(tc)
	if !r.Detected {
		t.Error("ChangeDetection: single non-round output should be detected via round_number")
	}
	if r.Extra["method"] != "round_number" {
		t.Errorf("method = %v, want round_number", r.Extra["method"])
	}
	if r.Extra["confidence"] != "medium" {
		t.Errorf("confidence = %v, want medium", r.Extra["confidence"])
	}
}

func TestApplyChangeDetection_SmallestOutput(t *testing.T) {
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(60_001, p2wpkhScript(10)))
	tx.AddTxOut(wire.NewTxOut(30_003, p2wpkhScript(11)))

	tc := &TxContext{
		Tx:       tx,
		PrevOuts: []blockfile.PrevOut{{Value: 100_000, ScriptPubKey: p2wpkhScript(0)}},
	}
	r := ApplyChangeDetection(tc)
	if !r.Detected {
		t.Error("ChangeDetection: 2-output tx should fall back to smallest_output")
	}
	if r.Extra["method"] != "smallest_output" {
		t.Errorf("method = %v, want smallest_output", r.Extra["method"])
	}
	if r.Extra["likely_change_index"] != 1 {
		t.Errorf("likely_change_index = %v, want 1", r.Extra["likely_change_index"])
	}
}
