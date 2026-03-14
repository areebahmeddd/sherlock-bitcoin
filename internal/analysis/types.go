package analysis

// Result is the top-level output of AnalyzeFile.
type Result struct {
	Ok              bool          `json:"ok"`
	Mode            string        `json:"mode"`
	File            string        `json:"file"`
	BlockCount      int           `json:"block_count"`
	AnalysisSummary Summary       `json:"analysis_summary"`
	Blocks          []BlockResult `json:"blocks"`
}

// Summary aggregates heuristic and fee statistics for a result or block.
type Summary struct {
	TotalTxsAnalyzed    int            `json:"total_transactions_analyzed"`
	HeuristicsApplied   []string       `json:"heuristics_applied"`
	FlaggedTransactions int            `json:"flagged_transactions"`
	ScriptTypeDist      map[string]int `json:"script_type_distribution"`
	FeeRateStats        FeeStats       `json:"fee_rate_stats"`
}

// FeeStats holds min/max/median/mean fee rates in sat/vB.
type FeeStats struct {
	MinSatVB    float64 `json:"min_sat_vb"`
	MaxSatVB    float64 `json:"max_sat_vb"`
	MedianSatVB float64 `json:"median_sat_vb"`
	MeanSatVB   float64 `json:"mean_sat_vb"`
}

// BlockResult holds analysis output for a single block.
type BlockResult struct {
	BlockHash       string     `json:"block_hash"`
	BlockHeight     int        `json:"block_height"`
	TxCount         int        `json:"tx_count"`
	AnalysisSummary Summary    `json:"analysis_summary"`
	Transactions    []TxResult `json:"transactions,omitempty"`
}

// TxResult holds the per-transaction heuristic output.
type TxResult struct {
	TXID           string                 `json:"txid"`
	Heuristics     map[string]interface{} `json:"heuristics"`
	Classification string                 `json:"classification"`
}

var allHeuristicIDs = []string{
	"cioh",
	"change_detection",
	"coinjoin",
	"consolidation",
	"self_transfer",
	"address_reuse",
	"op_return",
	"round_number_payment",
	"peeling_chain",
}
