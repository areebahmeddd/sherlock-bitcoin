package heuristics

import (
	"testing"

	"github.com/btcsuite/btcd/wire"
)

func TestApplyRoundNumberPayment_RoundOutput(t *testing.T) {
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	script1 := make([]byte, 22)
	script1[0] = 0x00
	script1[1] = 0x14
	tx.AddTxOut(wire.NewTxOut(100_000_000, script1))
	script2 := make([]byte, 22)
	script2[0] = 0x00
	script2[1] = 0x14
	script2[2] = 0x01
	tx.AddTxOut(wire.NewTxOut(89_997_312, script2))

	tc := &TxContext{Tx: tx, IsCoinbase: false}
	if got := ApplyRoundNumberPayment(tc).Detected; !got {
		t.Error("RoundPayment: 1 BTC output should be flagged")
	}
}

func TestApplyRoundNumberPayment_NonRoundOutput(t *testing.T) {
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(99_999_999, p2wpkhScript(0)))
	tc := &TxContext{Tx: tx, IsCoinbase: false}
	if got := ApplyRoundNumberPayment(tc).Detected; got {
		t.Error("RoundPayment: non-round output should NOT be flagged")
	}
}

func TestApplyRoundNumberPayment_Coinbase(t *testing.T) {
	tc := makeTx(1, 1)
	tc.IsCoinbase = true
	if got := ApplyRoundNumberPayment(tc).Detected; got {
		t.Error("RoundPayment: coinbase should NOT be flagged")
	}
}

func TestIsRoundAmount_Zero(t *testing.T) {
	if isRoundAmount(0) {
		t.Error("isRoundAmount(0) should be false")
	}
}

func TestIsRoundAmount_Negative(t *testing.T) {
	if isRoundAmount(-1000) {
		t.Error("isRoundAmount(-1000) should be false")
	}
}
