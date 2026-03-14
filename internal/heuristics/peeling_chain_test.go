package heuristics

import (
	"testing"

	"github.com/btcsuite/btcd/wire"

	"sherlock/internal/blockfile"
)

func TestApplyPeelingChain_Detected(t *testing.T) {
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(1_000, p2wpkhScript(10)))
	tx.AddTxOut(wire.NewTxOut(90_000, p2wpkhScript(11)))

	tc := &TxContext{
		Tx:       tx,
		PrevOuts: []blockfile.PrevOut{{Value: 100_000, ScriptPubKey: p2wpkhScript(0)}},
	}
	if got := ApplyPeelingChain(tc).Detected; !got {
		t.Error("PeelingChain: 1-in 2-out with 90x ratio should be detected")
	}
}
