package analysis

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/btcsuite/btcd/wire"

	"sherlock/internal/blockfile"
	"sherlock/internal/heuristics"
)

// AnalyzeFile runs chain analysis on all blocks from a blk*.dat file.
func AnalyzeFile(blkPath, revPath, xorPath string) (*Result, error) {
	xorKey, err := blockfile.LoadXORKey(xorPath)
	if err != nil {
		return nil, fmt.Errorf("load xor key: %w", err)
	}

	blocks, err := blockfile.ReadBlocks(blkPath, xorKey)
	if err != nil {
		return nil, fmt.Errorf("read blocks: %w", err)
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no blocks found in %s", blkPath)
	}

	undos, err := blockfile.ReadBlockUndos(revPath, xorKey)
	if err != nil {
		undos = nil // continue without undo data
	}

	var orderedUndos []*blockfile.BlockUndo
	if undos != nil {
		orderedUndos = blockfile.MatchUndosByHeight(undos, blocks)
	}

	fileName := filepath.Base(blkPath)

	result := &Result{
		Ok:   true,
		Mode: "chain_analysis",
		File: fileName,
	}

	var allFeeRates []float64
	var totalTxs, totalFlagged int
	scriptTypeTotals := make(map[string]int)

	for blockIdx, block := range blocks {
		var undo *blockfile.BlockUndo
		if orderedUndos != nil && blockIdx < len(orderedUndos) {
			undo = orderedUndos[blockIdx]
		}

		br, feeRates := analyzeBlock(block, undo, blockIdx == 0)

		allFeeRates = append(allFeeRates, feeRates...)
		totalTxs += br.TxCount
		totalFlagged += br.AnalysisSummary.FlaggedTransactions
		for k, v := range br.AnalysisSummary.ScriptTypeDist {
			scriptTypeTotals[k] += v
		}

		result.Blocks = append(result.Blocks, br)
	}

	result.BlockCount = len(blocks)
	result.AnalysisSummary = Summary{
		TotalTxsAnalyzed:    totalTxs,
		HeuristicsApplied:   allHeuristicIDs,
		FlaggedTransactions: totalFlagged,
		ScriptTypeDist:      scriptTypeTotals,
		FeeRateStats:        computeFeeStats(allFeeRates),
	}

	return result, nil
}

// analyzeBlock runs all heuristics on a single block.
// Transactions is omitted from the result unless includeTransactions is true.
func analyzeBlock(block *wire.MsgBlock, undo *blockfile.BlockUndo, includeTransactions bool) (BlockResult, []float64) {
	blockHash := toLowerHex(block.BlockHash().String())
	height := extractBlockHeight(block)

	var txResults []TxResult
	var feeRates []float64
	var flagged int
	scriptTypeDist := make(map[string]int)

	for txIdx, tx := range block.Transactions {
		isCoinbase := txIdx == 0

		// Gather prevouts from undo data (skip coinbase)
		var prevOuts []blockfile.PrevOut
		if !isCoinbase && undo != nil {
			undoTxIdx := txIdx - 1 // undo skips coinbase
			if undoTxIdx < len(undo.TxUndos) {
				prevOuts = undo.TxUndos[undoTxIdx].PrevOuts
			}
		}

		tc := &heuristics.TxContext{
			Tx:         tx,
			PrevOuts:   prevOuts,
			IsCoinbase: isCoinbase,
		}

		h := runHeuristics(tc)

		for _, out := range tx.TxOut {
			t := heuristics.ScriptType(out.PkScript)
			scriptTypeDist[t]++
		}

		if !isCoinbase {
			fee := tc.Fee()
			if fee >= 0 {
				vs := tc.VSize()
				if vs > 0 {
					feeRates = append(feeRates, float64(fee)/float64(vs))
				}
			}
		}

		txFlagged := h.CIOH.Detected || h.ChangeDetection.Detected ||
			h.Coinjoin.Detected || h.Consolidation.Detected ||
			h.SelfTransfer.Detected || h.AddressReuse.Detected ||
			h.OPReturn.Detected || h.RoundPayment.Detected ||
			h.PeelingChain.Detected

		if txFlagged {
			flagged++
		}

		if includeTransactions {
			txResults = append(txResults, buildTxResult(tx, h))
		}
	}

	blockSummary := Summary{
		TotalTxsAnalyzed:    len(block.Transactions),
		HeuristicsApplied:   allHeuristicIDs,
		FlaggedTransactions: flagged,
		ScriptTypeDist:      scriptTypeDist,
		FeeRateStats:        computeFeeStats(feeRates),
	}

	br := BlockResult{
		BlockHash:       blockHash,
		BlockHeight:     height,
		TxCount:         len(block.Transactions),
		AnalysisSummary: blockSummary,
	}
	if includeTransactions {
		br.Transactions = txResults
	}

	return br, feeRates
}

// runHeuristics applies all heuristics to tc and returns the combined results.
func runHeuristics(tc *heuristics.TxContext) *heuristics.TxHeuristics {
	h := &heuristics.TxHeuristics{}
	h.CIOH = heuristics.ApplyCIOH(tc)
	h.ChangeDetection = heuristics.ApplyChangeDetection(tc)
	h.Coinjoin = heuristics.ApplyCoinjoin(tc)
	h.Consolidation = heuristics.ApplyConsolidation(tc)
	h.SelfTransfer = heuristics.ApplySelfTransfer(tc)
	h.AddressReuse = heuristics.ApplyAddressReuse(tc)
	h.OPReturn = heuristics.ApplyOPReturn(tc)
	h.RoundPayment = heuristics.ApplyRoundNumberPayment(tc)
	h.PeelingChain = heuristics.ApplyPeelingChain(tc)
	h.Classification = heuristics.Classify(tc, h)
	return h
}

func buildTxResult(tx *wire.MsgTx, h *heuristics.TxHeuristics) TxResult {
	txid := toLowerHex(tx.TxHash().String())

	hMap := map[string]interface{}{
		"cioh": map[string]interface{}{
			"detected": h.CIOH.Detected,
		},
		"change_detection": buildHeuristicMap(h.ChangeDetection),
		"coinjoin": map[string]interface{}{
			"detected": h.Coinjoin.Detected,
		},
		"consolidation": map[string]interface{}{
			"detected": h.Consolidation.Detected,
		},
		"self_transfer": map[string]interface{}{
			"detected": h.SelfTransfer.Detected,
		},
		"address_reuse": map[string]interface{}{
			"detected": h.AddressReuse.Detected,
		},
		"op_return": buildHeuristicMap(h.OPReturn),
		"round_number_payment": map[string]interface{}{
			"detected": h.RoundPayment.Detected,
		},
		"peeling_chain": map[string]interface{}{
			"detected": h.PeelingChain.Detected,
		},
	}

	return TxResult{
		TXID:           txid,
		Heuristics:     hMap,
		Classification: h.Classification,
	}
}

func buildHeuristicMap(h heuristics.HeuristicResult) map[string]interface{} {
	m := map[string]interface{}{
		"detected": h.Detected,
	}
	for k, v := range h.Extra {
		m[k] = v
	}
	return m
}

func toLowerHex(s string) string {
	return strings.ToLower(s)
}
