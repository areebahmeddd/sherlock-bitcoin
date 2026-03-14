package heuristics

import (
	"testing"

	"github.com/btcsuite/btcd/wire"

	"sherlock/internal/blockfile"
)

func TestApplyAddressReuse_SameScript(t *testing.T) {
	script := []byte{0x00, 0x14, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(50000, script))

	tc := &TxContext{
		Tx:         tx,
		PrevOuts:   []blockfile.PrevOut{{Value: 60000, ScriptPubKey: script}},
		IsCoinbase: false,
	}

	if got := ApplyAddressReuse(tc).Detected; !got {
		t.Error("AddressReuse: prevout script matching output should be flagged")
	}
}

func TestApplyAddressReuse_DifferentScripts(t *testing.T) {
	outScript := []byte{0x00, 0x14, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	inScript := []byte{0x00, 0x14, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40}

	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(50000, outScript))

	tc := &TxContext{
		Tx:         tx,
		PrevOuts:   []blockfile.PrevOut{{Value: 60000, ScriptPubKey: inScript}},
		IsCoinbase: false,
	}

	if got := ApplyAddressReuse(tc).Detected; got {
		t.Error("AddressReuse: different scripts should NOT be flagged")
	}
}

func TestApplyAddressReuse_Coinbase(t *testing.T) {
	tc := makeTx(1, 1)
	tc.IsCoinbase = true
	if got := ApplyAddressReuse(tc).Detected; got {
		t.Error("AddressReuse: coinbase should never be flagged")
	}
}
