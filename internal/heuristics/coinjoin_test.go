package heuristics

import (
	"testing"

	"github.com/btcsuite/btcd/wire"

	"sherlock/internal/blockfile"
)

func TestApplyCoinjoin_Detected(t *testing.T) {
	tx := wire.NewMsgTx(2)
	var prevOuts []blockfile.PrevOut
	for i := 0; i < 3; i++ {
		tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
		prevOuts = append(prevOuts, blockfile.PrevOut{Value: 50_000, ScriptPubKey: p2wpkhScript(byte(i))})
	}
	tx.AddTxOut(wire.NewTxOut(40_000, p2wpkhScript(10)))
	tx.AddTxOut(wire.NewTxOut(40_000, p2wpkhScript(11)))
	tx.AddTxOut(wire.NewTxOut(20_000, p2wpkhScript(12)))

	tc := &TxContext{Tx: tx, PrevOuts: prevOuts}
	if got := ApplyCoinjoin(tc).Detected; !got {
		t.Error("Coinjoin: 3 inputs with equal-value outputs should be detected")
	}
}

func TestApplyCoinjoin_AllDifferentValues(t *testing.T) {
	tx := wire.NewMsgTx(2)
	var prevOuts []blockfile.PrevOut
	for i := 0; i < 3; i++ {
		tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
		prevOuts = append(prevOuts, blockfile.PrevOut{Value: 50_000, ScriptPubKey: p2wpkhScript(byte(i))})
	}
	tx.AddTxOut(wire.NewTxOut(10_000, p2wpkhScript(10)))
	tx.AddTxOut(wire.NewTxOut(20_000, p2wpkhScript(11)))
	tx.AddTxOut(wire.NewTxOut(30_000, p2wpkhScript(12)))

	tc := &TxContext{Tx: tx, PrevOuts: prevOuts}
	if got := ApplyCoinjoin(tc).Detected; got {
		t.Error("Coinjoin: all-different output values should NOT be detected")
	}
}
