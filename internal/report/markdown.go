package report

import (
	"fmt"
	"strings"

	"sherlock/internal/analysis"
)

// Generate creates a Markdown report string from an analysis result.
func Generate(r *analysis.Result) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Chain Analysis Report: %s\n\n", r.File))
	sb.WriteString(fmt.Sprintf("**Source file:** `%s`  \n", r.File))
	sb.WriteString(fmt.Sprintf("**Blocks analyzed:** %d  \n", r.BlockCount))
	sb.WriteString(fmt.Sprintf("**Total transactions:** %d  \n\n", r.AnalysisSummary.TotalTxsAnalyzed))

	sb.WriteString("## Summary Statistics\n\n")

	sb.WriteString("### Fee Rate Distribution (sat/vB)\n\n")
	sb.WriteString("| Metric | Value |\n|--------|-------|\n")
	fs := r.AnalysisSummary.FeeRateStats
	sb.WriteString(fmt.Sprintf("| Min | %.2f |\n", fs.MinSatVB))
	sb.WriteString(fmt.Sprintf("| Median | %.2f |\n", fs.MedianSatVB))
	sb.WriteString(fmt.Sprintf("| Mean | %.2f |\n", fs.MeanSatVB))
	sb.WriteString(fmt.Sprintf("| Max | %.2f |\n\n", fs.MaxSatVB))

	sb.WriteString("### Script Type Distribution\n\n")
	sb.WriteString("| Script Type | Count |\n|-------------|-------|\n")
	for _, st := range []string{"p2wpkh", "p2tr", "p2sh", "p2pkh", "p2wsh", "op_return", "unknown"} {
		if v, ok := r.AnalysisSummary.ScriptTypeDist[st]; ok && v > 0 {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", st, v))
		}
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("**Flagged transactions:** %d  \n", r.AnalysisSummary.FlaggedTransactions))
	sb.WriteString(fmt.Sprintf("**Heuristics applied:** %s  \n\n",
		strings.Join(r.AnalysisSummary.HeuristicsApplied, ", ")))

	for i, block := range r.Blocks {
		sb.WriteString(fmt.Sprintf("---\n\n## Block %d\n\n", i+1))
		sb.WriteString("| Field | Value |\n|-------|-------|\n")
		sb.WriteString(fmt.Sprintf("| **Block Hash** | `%s` |\n", block.BlockHash))
		sb.WriteString(fmt.Sprintf("| **Block Height** | %d |\n", block.BlockHeight))
		sb.WriteString(fmt.Sprintf("| **Transaction Count** | %d |\n\n", block.TxCount))

		sb.WriteString("### Fee Rate Stats (sat/vB)\n\n")
		sb.WriteString("| Min | Median | Mean | Max |\n|-----|--------|------|-----|\n")
		bfs := block.AnalysisSummary.FeeRateStats
		sb.WriteString(fmt.Sprintf("| %.2f | %.2f | %.2f | %.2f |\n\n",
			bfs.MinSatVB, bfs.MedianSatVB, bfs.MeanSatVB, bfs.MaxSatVB))

		sb.WriteString("### Heuristic Findings\n\n")
		sb.WriteString("| Heuristic | Flagged Transactions |\n|-----------|---------------------|\n")
		sb.WriteString("| cioh | see flagged count |\n")
		sb.WriteString("| change_detection | see flagged count |\n")
		sb.WriteString("| coinjoin | see flagged count |\n")
		sb.WriteString("| consolidation | see flagged count |\n")
		sb.WriteString("| self_transfer | see flagged count |\n")
		sb.WriteString("| address_reuse | see flagged count |\n")
		sb.WriteString("| op_return | see flagged count |\n")
		sb.WriteString("| round_number_payment | see flagged count |\n")
		sb.WriteString("| peeling_chain | see flagged count |\n\n")

		sb.WriteString(fmt.Sprintf("**Flagged transactions in this block:** %d / %d  \n\n",
			block.AnalysisSummary.FlaggedTransactions, block.TxCount))

		sb.WriteString("### Script Type Distribution\n\n")
		sb.WriteString("| Script Type | Count |\n|-------------|-------|\n")
		for _, st := range []string{"p2wpkh", "p2tr", "p2sh", "p2pkh", "p2wsh", "op_return", "unknown"} {
			if v, ok := block.AnalysisSummary.ScriptTypeDist[st]; ok && v > 0 {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", st, v))
			}
		}
		sb.WriteString("\n")

		if len(block.Transactions) > 0 {
			sb.WriteString("### Notable Transactions\n\n")
			notable := getNotableTransactions(block.Transactions, 10)
			if len(notable) > 0 {
				sb.WriteString("| TXID | Classification | Heuristics |\n|------|----------------|------------|\n")
				for _, tx := range notable {
					hList := getDetectedHeuristics(tx)
					txidShort := tx.TXID
					if len(txidShort) > 16 {
						txidShort = txidShort[:8] + "..." + txidShort[len(txidShort)-8:]
					}
					sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n",
						txidShort, tx.Classification, strings.Join(hList, ", ")))
				}
				sb.WriteString("\n")
			} else {
				sb.WriteString("No notable transactions found in this block.\n\n")
			}
		}
	}

	sb.WriteString("---\n\n## Heuristic Catalogue\n\n")
	sb.WriteString("| Heuristic | Description |\n|-----------|-------------|\n")
	sb.WriteString("| `cioh` | Common Input Ownership — multiple inputs likely from same wallet |\n")
	sb.WriteString("| `change_detection` | Identifies likely change output via script type matching, round numbers, and smallest output |\n")
	sb.WriteString("| `coinjoin` | Detects equal-value outputs with many inputs (CoinJoin pattern) |\n")
	sb.WriteString("| `consolidation` | Many inputs merged into 1–2 outputs of the same type |\n")
	sb.WriteString("| `self_transfer` | All outputs match input script type — likely internal wallet move |\n")
	sb.WriteString("| `address_reuse` | Same address appears in both inputs and outputs |\n")
	sb.WriteString("| `op_return` | OP_RETURN output detected, protocol identified |\n")
	sb.WriteString("| `round_number_payment` | Output value is a round BTC/sat amount |\n")
	sb.WriteString("| `peeling_chain` | 1 input → 1 small + 1 large output (peeling chain pattern) |\n")
	sb.WriteString("\n")

	return sb.String()
}

func getNotableTransactions(txs []analysis.TxResult, maxCount int) []analysis.TxResult {
	var notable []analysis.TxResult
	for _, tx := range txs {
		if tx.Classification == "coinjoin" ||
			tx.Classification == "consolidation" ||
			isInteresting(tx) {
			notable = append(notable, tx)
			if len(notable) >= maxCount {
				break
			}
		}
	}
	// If not enough, add any flagged transactions
	if len(notable) < maxCount {
		for _, tx := range txs {
			if len(getDetectedHeuristics(tx)) > 0 && !alreadyIncluded(notable, tx.TXID) {
				notable = append(notable, tx)
				if len(notable) >= maxCount {
					break
				}
			}
		}
	}
	return notable
}

func isInteresting(tx analysis.TxResult) bool {
	detected := getDetectedHeuristics(tx)
	return len(detected) >= 2
}

func getDetectedHeuristics(tx analysis.TxResult) []string {
	var detected []string
	for name, val := range tx.Heuristics {
		if m, ok := val.(map[string]interface{}); ok {
			if d, ok := m["detected"].(bool); ok && d {
				detected = append(detected, name)
			}
		}
	}
	return detected
}

func alreadyIncluded(txs []analysis.TxResult, txid string) bool {
	for _, t := range txs {
		if t.TXID == txid {
			return true
		}
	}
	return false
}
