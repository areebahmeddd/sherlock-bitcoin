package analysis

import (
	"encoding/binary"
	"math"
	"path/filepath"
	"testing"

	"github.com/btcsuite/btcd/wire"
)

func TestComputeFeeStats_Empty(t *testing.T) {
	fs := computeFeeStats(nil)
	if fs.MinSatVB != 0 || fs.MaxSatVB != 0 || fs.MedianSatVB != 0 || fs.MeanSatVB != 0 {
		t.Errorf("empty slice should yield zero FeeStats, got %+v", fs)
	}
}

func TestComputeFeeStats_SingleValue(t *testing.T) {
	fs := computeFeeStats([]float64{4.5})
	if math.Abs(fs.MinSatVB-4.5) > 0.001 {
		t.Errorf("min = %.4f, want 4.5", fs.MinSatVB)
	}
	if math.Abs(fs.MaxSatVB-4.5) > 0.001 {
		t.Errorf("max = %.4f, want 4.5", fs.MaxSatVB)
	}
	if math.Abs(fs.MedianSatVB-4.5) > 0.001 {
		t.Errorf("median = %.4f, want 4.5", fs.MedianSatVB)
	}
	if math.Abs(fs.MeanSatVB-4.5) > 0.001 {
		t.Errorf("mean = %.4f, want 4.5", fs.MeanSatVB)
	}
}

func TestComputeFeeStats_OddCount(t *testing.T) {
	fs := computeFeeStats([]float64{8, 1, 16, 2, 4})
	if math.Abs(fs.MinSatVB-1.0) > 0.001 {
		t.Errorf("min = %.4f, want 1.0", fs.MinSatVB)
	}
	if math.Abs(fs.MaxSatVB-16.0) > 0.001 {
		t.Errorf("max = %.4f, want 16.0", fs.MaxSatVB)
	}
	if math.Abs(fs.MedianSatVB-4.0) > 0.001 {
		t.Errorf("median = %.4f, want 4.0", fs.MedianSatVB)
	}
	if math.Abs(fs.MeanSatVB-6.2) > 0.001 {
		t.Errorf("mean = %.4f, want 6.2", fs.MeanSatVB)
	}
}

func TestComputeFeeStats_EvenCount(t *testing.T) {
	fs := computeFeeStats([]float64{4, 1, 8, 2})
	if math.Abs(fs.MinSatVB-1.0) > 0.001 {
		t.Errorf("min = %.4f, want 1.0", fs.MinSatVB)
	}
	if math.Abs(fs.MaxSatVB-8.0) > 0.001 {
		t.Errorf("max = %.4f, want 8.0", fs.MaxSatVB)
	}
	if math.Abs(fs.MedianSatVB-3.0) > 0.001 {
		t.Errorf("median = %.4f, want 3.0", fs.MedianSatVB)
	}
	if math.Abs(fs.MeanSatVB-3.75) > 0.001 {
		t.Errorf("mean = %.4f, want 3.75", fs.MeanSatVB)
	}
}

func TestComputeFeeStats_OrderIndependent(t *testing.T) {
	a := computeFeeStats([]float64{1, 2, 3})
	b := computeFeeStats([]float64{3, 1, 2})
	if a != b {
		t.Errorf("order-independence violated: %+v != %+v", a, b)
	}
}

func buildBlockAtHeight(t *testing.T, height int) *wire.MsgBlock {
	t.Helper()
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(height))
	n := 4
	for n > 1 && buf[n-1] == 0 {
		n--
	}
	script := make([]byte, 1+n)
	script[0] = byte(n)
	copy(script[1:], buf[:n])

	hdr := wire.BlockHeader{}
	block := wire.NewMsgBlock(&hdr)
	cb := wire.NewMsgTx(1)
	cb.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{Index: 0xffffffff},
		SignatureScript:  script,
		Sequence:         0xffffffff,
	})
	cb.AddTxOut(&wire.TxOut{PkScript: []byte{0x51}})
	block.AddTransaction(cb)
	return block
}

func TestExtractBlockHeight_ValidBIP34(t *testing.T) {
	tests := []int{1, 100, 847493, 907139}
	for _, want := range tests {
		b := buildBlockAtHeight(t, want)
		got := extractBlockHeight(b)
		if got != want {
			t.Errorf("extractBlockHeight(%d) = %d", want, got)
		}
	}
}

func TestExtractBlockHeight_EmptyBlock(t *testing.T) {
	hdr := wire.BlockHeader{}
	block := wire.NewMsgBlock(&hdr)
	if got := extractBlockHeight(block); got != 0 {
		t.Errorf("extractBlockHeight empty block = %d, want 0", got)
	}
}

func TestExtractBlockHeight_EmptyScriptSig(t *testing.T) {
	hdr := wire.BlockHeader{}
	block := wire.NewMsgBlock(&hdr)
	cb := wire.NewMsgTx(1)
	cb.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{Index: 0xffffffff},
		SignatureScript:  []byte{},
	})
	cb.AddTxOut(&wire.TxOut{PkScript: []byte{0x51}})
	block.AddTransaction(cb)
	if got := extractBlockHeight(block); got != 0 {
		t.Errorf("extractBlockHeight with empty scriptSig = %d, want 0", got)
	}
}

func TestToLowerHex_Uppercase(t *testing.T) {
	if got := toLowerHex("ABCDEF"); got != "abcdef" {
		t.Errorf("toLowerHex(ABCDEF) = %q, want abcdef", got)
	}
}

func TestToLowerHex_AlreadyLowercase(t *testing.T) {
	s := "deadbeef"
	if got := toLowerHex(s); got != s {
		t.Errorf("toLowerHex(%q) = %q, want identity", s, got)
	}
}

func fixturesDir(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "fixtures")
}

func TestAnalyzeFile_blk04330(t *testing.T) {
	dir := fixturesDir(t)
	blk := filepath.Join(dir, "blk04330.dat")
	rev := filepath.Join(dir, "rev04330.dat")
	xor := filepath.Join(dir, "xor.dat")

	result, err := AnalyzeFile(blk, rev, xor)
	if err != nil {
		t.Fatalf("AnalyzeFile: %v", err)
	}
	if !result.Ok {
		t.Fatal("expected ok=true")
	}
	if result.Mode != "chain_analysis" {
		t.Errorf("mode = %q, want chain_analysis", result.Mode)
	}
	if result.BlockCount != len(result.Blocks) {
		t.Errorf("block_count %d != len(blocks) %d", result.BlockCount, len(result.Blocks))
	}
	if result.BlockCount == 0 {
		t.Fatal("expected > 0 blocks")
	}
	if len(result.Blocks[0].Transactions) == 0 {
		t.Error("first block must have non-empty transactions slice")
	}
	if result.Blocks[0].TxCount != len(result.Blocks[0].Transactions) {
		t.Errorf("tx_count %d != len(transactions) %d",
			result.Blocks[0].TxCount, len(result.Blocks[0].Transactions))
	}
	var sumTxs int
	for _, b := range result.Blocks {
		sumTxs += b.TxCount
	}
	if result.AnalysisSummary.TotalTxsAnalyzed != sumTxs {
		t.Errorf("total_transactions_analyzed %d != sum of tx_counts %d",
			result.AnalysisSummary.TotalTxsAnalyzed, sumTxs)
	}
	fs := result.AnalysisSummary.FeeRateStats
	if fs.MinSatVB > fs.MedianSatVB || fs.MedianSatVB > fs.MaxSatVB {
		t.Errorf("fee stats order violated: min=%.2f median=%.2f max=%.2f",
			fs.MinSatVB, fs.MedianSatVB, fs.MaxSatVB)
	}
	if len(result.AnalysisSummary.HeuristicsApplied) < 5 {
		t.Errorf("need >= 5 heuristics_applied, got %d", len(result.AnalysisSummary.HeuristicsApplied))
	}
}

func TestAnalyzeFile_blk05051(t *testing.T) {
	dir := fixturesDir(t)
	blk := filepath.Join(dir, "blk05051.dat")
	rev := filepath.Join(dir, "rev05051.dat")
	xor := filepath.Join(dir, "xor.dat")

	result, err := AnalyzeFile(blk, rev, xor)
	if err != nil {
		t.Fatalf("AnalyzeFile: %v", err)
	}
	if !result.Ok {
		t.Fatal("expected ok=true")
	}
	if result.BlockCount != len(result.Blocks) {
		t.Errorf("block_count %d != len(blocks) %d", result.BlockCount, len(result.Blocks))
	}
}

func TestAnalyzeFile_NonexistentFile(t *testing.T) {
	_, err := AnalyzeFile("/nonexistent/blk.dat", "/nonexistent/rev.dat", "/nonexistent/xor.dat")
	if err == nil {
		t.Fatal("expected error for nonexistent files")
	}
}
