package analysis

import (
	"math"
	"sort"

	"github.com/btcsuite/btcd/wire"
)

// computeFeeStats returns min/max/median/mean fee rates (sat/vB) for the given rates slice.
func computeFeeStats(rates []float64) FeeStats {
	if len(rates) == 0 {
		return FeeStats{}
	}

	sorted := make([]float64, len(rates))
	copy(sorted, rates)
	sort.Float64s(sorted)

	min := sorted[0]
	max := sorted[len(sorted)-1]

	var sum float64
	for _, r := range sorted {
		sum += r
	}
	mean := sum / float64(len(sorted))

	var median float64
	n := len(sorted)
	if n%2 == 0 {
		median = (sorted[n/2-1] + sorted[n/2]) / 2
	} else {
		median = sorted[n/2]
	}

	return FeeStats{
		MinSatVB:    roundTo2(min),
		MaxSatVB:    roundTo2(max),
		MedianSatVB: roundTo2(median),
		MeanSatVB:   roundTo2(mean),
	}
}

func roundTo2(v float64) float64 {
	return math.Round(v*100) / 100
}

// extractBlockHeight reads the block height from the coinbase scriptSig (BIP34).
func extractBlockHeight(block *wire.MsgBlock) int {
	if len(block.Transactions) == 0 {
		return 0
	}
	cb := block.Transactions[0]
	if len(cb.TxIn) == 0 {
		return 0
	}
	ss := cb.TxIn[0].SignatureScript
	if len(ss) < 2 {
		return 0
	}
	pushLen := int(ss[0])
	if pushLen == 0 || pushLen > 4 || len(ss) < 1+pushLen {
		return 0
	}
	heightBytes := ss[1 : 1+pushLen]
	var h int
	for i, b := range heightBytes {
		h |= int(b) << (i * 8)
	}
	return h
}
