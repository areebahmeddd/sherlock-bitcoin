package heuristics

import (
	"github.com/btcsuite/btcd/wire"

	"sherlock/internal/blockfile"
)

// TxContext bundles a transaction with its prevout data.
type TxContext struct {
	Tx         *wire.MsgTx
	PrevOuts   []blockfile.PrevOut
	IsCoinbase bool
}

// inputScriptTypes returns the script type for each input.
func (tc *TxContext) inputScriptTypes() []string {
	types := make([]string, len(tc.Tx.TxIn))
	for i, in := range tc.Tx.TxIn {
		if i < len(tc.PrevOuts) && len(tc.PrevOuts[i].ScriptPubKey) > 0 {
			types[i] = ScriptType(tc.PrevOuts[i].ScriptPubKey)
		} else {
			types[i] = inferInputScriptType(in)
		}
	}
	return types
}

// outputScriptTypes returns the script type for each output.
func (tc *TxContext) outputScriptTypes() []string {
	types := make([]string, len(tc.Tx.TxOut))
	for i, out := range tc.Tx.TxOut {
		types[i] = ScriptType(out.PkScript)
	}
	return types
}

// Fee returns the transaction fee in satoshis (or -1 if unknown).
func (tc *TxContext) Fee() int64 {
	if tc.IsCoinbase {
		return 0
	}
	var inSum int64
	for i := range tc.Tx.TxIn {
		if i < len(tc.PrevOuts) && len(tc.PrevOuts[i].ScriptPubKey) > 0 {
			inSum += tc.PrevOuts[i].Value
		} else {
			return -1 // unknown — prevout not available
		}
	}
	var outSum int64
	for _, out := range tc.Tx.TxOut {
		outSum += out.Value
	}
	if inSum < outSum {
		return -1
	}
	return inSum - outSum
}

// VSize returns the virtual size of the transaction in vbytes.
func (tc *TxContext) VSize() int {
	baseSize := tc.Tx.SerializeSizeStripped()
	totalSize := tc.Tx.SerializeSize()
	weight := baseSize*3 + totalSize
	return (weight + 3) / 4
}

// HeuristicResult holds the result of one heuristic applied to a transaction.
type HeuristicResult struct {
	Detected bool `json:"detected"`
	Extra    map[string]interface{}
}

// TxHeuristics holds all heuristic results for a single transaction.
type TxHeuristics struct {
	CIOH            HeuristicResult
	ChangeDetection HeuristicResult
	Coinjoin        HeuristicResult
	Consolidation   HeuristicResult
	SelfTransfer    HeuristicResult
	AddressReuse    HeuristicResult
	OPReturn        HeuristicResult
	RoundPayment    HeuristicResult
	PeelingChain    HeuristicResult
	Classification  string
}
