package heuristics

import (
	"testing"

	"github.com/btcsuite/btcd/wire"

	"sherlock/internal/blockfile"
)

func TestApplySelfTransfer_SingleInSingleOut(t *testing.T) {
	tc := makeTx(1, 1)
	if got := ApplySelfTransfer(tc).Detected; !got {
		t.Error("SelfTransfer: 1-in/1-out should be flagged")
	}
}

func TestApplySelfTransfer_MultipleOutputs(t *testing.T) {
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(50000, p2wpkhScript(10)))
	tx.AddTxOut(wire.NewTxOut(40000, p2pkhScript()))

	tc := &TxContext{
		Tx:       tx,
		PrevOuts: []blockfile.PrevOut{{Value: 100000, ScriptPubKey: p2wpkhScript(0)}},
	}
	if got := ApplySelfTransfer(tc).Detected; got {
		t.Error("SelfTransfer: mixed-type outputs should NOT be flagged")
	}
}

func TestApplySelfTransfer_ManyInputs(t *testing.T) {
	tc := makeTx(3, 1)
	if got := ApplySelfTransfer(tc).Detected; got {
		t.Error("SelfTransfer: >2 inputs should NOT be flagged")
	}
}
