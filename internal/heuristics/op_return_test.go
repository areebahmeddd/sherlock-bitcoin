package heuristics

import (
	"testing"

	"github.com/btcsuite/btcd/wire"
)

func TestApplyOPReturn_Detected(t *testing.T) {
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(0, []byte{0x6a, 0x04, 0xde, 0xad, 0xbe, 0xef}))
	tc := &TxContext{Tx: tx, IsCoinbase: false}
	if got := ApplyOPReturn(tc).Detected; !got {
		t.Error("OPReturn: OP_RETURN output should be flagged")
	}
}

func TestApplyOPReturn_NoOPReturn(t *testing.T) {
	tc := makeTx(1, 2)
	if got := ApplyOPReturn(tc).Detected; got {
		t.Error("OPReturn: no OP_RETURN output should NOT be flagged")
	}
}

func TestApplyOPReturn_Coinbase(t *testing.T) {
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(0, []byte{0x6a, 0x24, 0xaa, 0x21, 0xa9, 0xed}))
	tc := &TxContext{Tx: tx, IsCoinbase: true}
	if got := ApplyOPReturn(tc).Detected; got {
		t.Error("OPReturn: coinbase segwit commitment should NOT be flagged")
	}
}

func TestApplyOPReturn_OmniProtocol(t *testing.T) {
	script := []byte{0x6a, 0x0c, 0x6f, 0x6d, 0x6e, 0x69, 0x00, 0x00, 0x00, 0x1f, 0x00, 0x00, 0x00, 0x00}
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(0, script))
	tc := &TxContext{Tx: tx, IsCoinbase: false}
	r := ApplyOPReturn(tc)
	if !r.Detected {
		t.Error("OPReturn: Omni script should be detected")
	}
	if r.Extra["protocol"] != "omni" {
		t.Errorf("protocol = %v, want omni", r.Extra["protocol"])
	}
}

func TestApplyOPReturn_RunesProtocol(t *testing.T) {
	script := []byte{0x6a, 0x04, 0x52, 0x55, 0x4e, 0x45}
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(0, script))
	tc := &TxContext{Tx: tx, IsCoinbase: false}
	r := ApplyOPReturn(tc)
	if !r.Detected {
		t.Error("OPReturn: Runes script should be detected")
	}
	if r.Extra["protocol"] != "runes" {
		t.Errorf("protocol = %v, want runes", r.Extra["protocol"])
	}
}
