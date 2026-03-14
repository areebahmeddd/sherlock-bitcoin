package report

import (
	"strings"
	"testing"

	"sherlock/internal/analysis"
)

func minimalResult() *analysis.Result {
	heuristics := map[string]interface{}{
		"cioh":          map[string]interface{}{"detected": true},
		"coinjoin":      map[string]interface{}{"detected": false},
		"consolidation": map[string]interface{}{"detected": true},
		"change_detection": map[string]interface{}{
			"detected":            true,
			"likely_change_index": 1,
			"method":              "script_type_match",
			"confidence":          "high",
		},
		"self_transfer":        map[string]interface{}{"detected": false},
		"address_reuse":        map[string]interface{}{"detected": false},
		"op_return":            map[string]interface{}{"detected": false},
		"round_number_payment": map[string]interface{}{"detected": false},
		"peeling_chain":        map[string]interface{}{"detected": false},
	}

	txs := []analysis.TxResult{
		{TXID: "aabb" + strings.Repeat("00", 30), Heuristics: heuristics, Classification: "consolidation"},
		{TXID: "ccdd" + strings.Repeat("ff", 30), Heuristics: heuristics, Classification: "simple_payment"},
	}

	block := analysis.BlockResult{
		BlockHash:   strings.Repeat("ab", 32),
		BlockHeight: 800000,
		TxCount:     2,
		AnalysisSummary: analysis.Summary{
			TotalTxsAnalyzed:    2,
			HeuristicsApplied:   []string{"cioh", "change_detection", "coinjoin", "consolidation", "self_transfer", "address_reuse", "op_return", "round_number_payment", "peeling_chain"},
			FlaggedTransactions: 1,
			ScriptTypeDist:      map[string]int{"p2wpkh": 3, "p2tr": 1},
			FeeRateStats: analysis.FeeStats{
				MinSatVB: 1.0, MaxSatVB: 100.0, MedianSatVB: 28.0, MeanSatVB: 45.2,
			},
		},
		Transactions: txs,
	}

	return &analysis.Result{
		Ok:         true,
		Mode:       "chain_analysis",
		File:       "blk04330.dat",
		BlockCount: 1,
		AnalysisSummary: analysis.Summary{
			TotalTxsAnalyzed:    2,
			HeuristicsApplied:   []string{"cioh", "change_detection", "coinjoin", "consolidation", "self_transfer", "address_reuse", "op_return", "round_number_payment", "peeling_chain"},
			FlaggedTransactions: 1,
			ScriptTypeDist:      map[string]int{"p2wpkh": 3, "p2tr": 1},
			FeeRateStats: analysis.FeeStats{
				MinSatVB: 1.0, MaxSatVB: 100.0, MedianSatVB: 28.0, MeanSatVB: 45.2,
			},
		},
		Blocks: []analysis.BlockResult{block},
	}
}

func TestGenerate_NotEmpty(t *testing.T) {
	r := minimalResult()
	out := Generate(r)
	if out == "" {
		t.Fatal("Generate returned empty string")
	}
}

func TestGenerate_ContainsFilename(t *testing.T) {
	r := minimalResult()
	out := Generate(r)
	if !strings.Contains(out, "blk04330.dat") {
		t.Error("report should contain the filename")
	}
}

func TestGenerate_ContainsMarkdownHeaders(t *testing.T) {
	r := minimalResult()
	out := Generate(r)
	for _, hdr := range []string{"# Chain Analysis Report", "## Summary Statistics", "## Block 1"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("expected header %q in report", hdr)
		}
	}
}

func TestGenerate_AtLeast1KB(t *testing.T) {
	r := minimalResult()
	out := Generate(r)
	if len(out) < 1024 {
		t.Errorf("report length %d is less than 1 KB (1024 bytes)", len(out))
	}
}

func TestGenerate_ContainsHeuristicCatalogue(t *testing.T) {
	r := minimalResult()
	out := Generate(r)
	if !strings.Contains(out, "cioh") {
		t.Error("report should contain heuristic ids")
	}
	if !strings.Contains(out, "change_detection") {
		t.Error("report should contain change_detection")
	}
}

func TestGenerate_ContainsFeeStats(t *testing.T) {
	r := minimalResult()
	out := Generate(r)
	if !strings.Contains(out, "sat/vB") {
		t.Error("report should mention sat/vB units")
	}
}

func TestGenerate_NotableTransactions(t *testing.T) {
	r := minimalResult()
	out := Generate(r)
	if !strings.Contains(out, "consolidation") {
		t.Error("report should list consolidation in notable transactions")
	}
}

func TestGetNotableTransactions_Empty(t *testing.T) {
	result := getNotableTransactions(nil, 10)
	if len(result) != 0 {
		t.Errorf("expected empty result for nil input, got %d", len(result))
	}
}

func TestGetNotableTransactions_Coinjoin(t *testing.T) {
	txs := []analysis.TxResult{
		{TXID: "aabb", Classification: "coinjoin", Heuristics: map[string]interface{}{
			"coinjoin": map[string]interface{}{"detected": true},
		}},
	}
	result := getNotableTransactions(txs, 10)
	if len(result) != 1 || result[0].TXID != "aabb" {
		t.Error("coinjoin tx should be included in notable")
	}
}

func TestGetNotableTransactions_MaxCount(t *testing.T) {
	var txs []analysis.TxResult
	for i := 0; i < 20; i++ {
		txs = append(txs, analysis.TxResult{
			TXID:           "tx" + string(rune('a'+i)),
			Classification: "coinjoin",
			Heuristics: map[string]interface{}{
				"coinjoin": map[string]interface{}{"detected": true},
			},
		})
	}
	result := getNotableTransactions(txs, 5)
	if len(result) > 5 {
		t.Errorf("expected at most 5 results, got %d", len(result))
	}
}

func TestIsInteresting_TwoHeuristics(t *testing.T) {
	tx := analysis.TxResult{
		TXID: "aabb",
		Heuristics: map[string]interface{}{
			"cioh":          map[string]interface{}{"detected": true},
			"consolidation": map[string]interface{}{"detected": true},
		},
	}
	if !isInteresting(tx) {
		t.Error("tx with 2 detected heuristics should be interesting")
	}
}

func TestIsInteresting_OneHeuristic(t *testing.T) {
	tx := analysis.TxResult{
		TXID: "aabb",
		Heuristics: map[string]interface{}{
			"cioh": map[string]interface{}{"detected": true},
		},
	}
	if isInteresting(tx) {
		t.Error("tx with only 1 detected heuristic should NOT be interesting")
	}
}

func TestAlreadyIncluded_Found(t *testing.T) {
	txs := []analysis.TxResult{{TXID: "aabb"}, {TXID: "ccdd"}}
	if !alreadyIncluded(txs, "ccdd") {
		t.Error("alreadyIncluded should return true for ccdd")
	}
}

func TestAlreadyIncluded_NotFound(t *testing.T) {
	txs := []analysis.TxResult{{TXID: "aabb"}}
	if alreadyIncluded(txs, "zzzz") {
		t.Error("alreadyIncluded should return false for unknown txid")
	}
}
