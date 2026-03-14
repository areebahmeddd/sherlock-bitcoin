package heuristics

import (
	"github.com/btcsuite/btcd/wire"

	"sherlock/internal/blockfile"
)

func p2wpkhScript(seed byte) []byte {
	s := make([]byte, 22)
	s[0] = 0x00
	s[1] = 0x14
	for i := 2; i < 22; i++ {
		s[i] = seed + byte(i)
	}
	return s
}

func p2pkhScript() []byte {
	s := []byte{0x76, 0xa9, 0x14}
	s = append(s, make([]byte, 20)...)
	s = append(s, 0x88, 0xac)
	return s
}

func makeTx(inputs, outputs int) *TxContext {
	tx := wire.NewMsgTx(2)
	var prevOuts []blockfile.PrevOut
	for i := 0; i < inputs; i++ {
		in := wire.NewTxIn(&wire.OutPoint{}, nil, nil)
		tx.AddTxIn(in)
		prevOuts = append(prevOuts, blockfile.PrevOut{
			Value:        50_000,
			ScriptPubKey: p2wpkhScript(byte(i)),
		})
	}
	for i := 0; i < outputs; i++ {
		tx.AddTxOut(wire.NewTxOut(10_000+int64(i)*1_000, p2wpkhScript(byte(100+i))))
	}
	return &TxContext{Tx: tx, PrevOuts: prevOuts, IsCoinbase: false}
}
