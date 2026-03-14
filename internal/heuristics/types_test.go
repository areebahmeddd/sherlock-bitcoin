package heuristics

import (
	"testing"

	"github.com/btcsuite/btcd/wire"

	"sherlock/internal/blockfile"
)

func TestTxContextFee_KnownValues(t *testing.T) {
	tc := makeTx(2, 1)
	tc.Tx.TxOut[0].Value = 90_000
	fee := tc.Fee()
	if fee != 10_000 {
		t.Errorf("Fee() = %d, want 10_000", fee)
	}
}

func TestTxContextFee_UnknownPrevout(t *testing.T) {
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(50_000, p2wpkhScript(0)))
	tc := &TxContext{
		Tx:       tx,
		PrevOuts: []blockfile.PrevOut{{Value: 60_000, ScriptPubKey: nil}},
	}
	if got := tc.Fee(); got != -1 {
		t.Errorf("Fee() with empty ScriptPubKey = %d, want -1", got)
	}
}

func TestTxContextFee_Coinbase(t *testing.T) {
	tc := makeTx(1, 1)
	tc.IsCoinbase = true
	if got := tc.Fee(); got != 0 {
		t.Errorf("Fee() coinbase = %d, want 0", got)
	}
}

func TestTxContextVSize_NonSegwit(t *testing.T) {
	tc := makeTx(1, 1)
	vs := tc.VSize()
	if vs <= 0 {
		t.Errorf("VSize() = %d, want > 0", vs)
	}
}
