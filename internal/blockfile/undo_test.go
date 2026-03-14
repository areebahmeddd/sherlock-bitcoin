package blockfile

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/btcsuite/btcd/wire"
)

func TestDecompressAmount(t *testing.T) {
	tests := []struct {
		compressed uint64
		want       int64
	}{
		{0, 0},
		{1, 1},
		{9, 100_000_000},
		{10, 1_000_000_000},
		{11, 2},
	}
	for _, tt := range tests {
		got := decompressAmount(tt.compressed)
		if got != tt.want {
			t.Errorf("decompressAmount(%d) = %d, want %d", tt.compressed, got, tt.want)
		}
	}
}

func TestDecompressAmount_NonNegative(t *testing.T) {
	for x := uint64(0); x < 200; x++ {
		if v := decompressAmount(x); v < 0 {
			t.Errorf("decompressAmount(%d) = %d, want >= 0", x, v)
		}
	}
}

func TestReadVarInt(t *testing.T) {
	tests := []struct {
		data []byte
		want uint64
	}{
		{[]byte{0x00}, 0},
		{[]byte{0x01}, 1},
		{[]byte{0x7F}, 127},
		{[]byte{0x09}, 9},
		{[]byte{0xe6, 0xb6, 0x4c}, 1694668},
	}
	for _, tt := range tests {
		r := bytes.NewReader(tt.data)
		got, err := readVarInt(r)
		if err != nil {
			t.Errorf("readVarInt(%x): unexpected error: %v", tt.data, err)
			continue
		}
		if got != tt.want {
			t.Errorf("readVarInt(%x) = %d, want %d", tt.data, got, tt.want)
		}
	}
}

func TestReadCompactSize(t *testing.T) {
	tests := []struct {
		data []byte
		want uint64
	}{
		{[]byte{0x00}, 0},
		{[]byte{0x02}, 2},
		{[]byte{0xfc}, 252},
		{[]byte{0xfd, 0xf3, 0x0d}, 3571},
	}
	for _, tt := range tests {
		r := bytes.NewReader(tt.data)
		got, err := readCompactSize(r)
		if err != nil {
			t.Errorf("readCompactSize(%x): unexpected error: %v", tt.data, err)
			continue
		}
		if got != tt.want {
			t.Errorf("readCompactSize(%x) = %d, want %d", tt.data, got, tt.want)
		}
	}
}

func buildRevBlock(t *testing.T) []byte {
	t.Helper()

	var coinBuf bytes.Buffer
	coinBuf.Write([]byte{0x80, 0x48})
	coinBuf.WriteByte(0x00)
	coinBuf.WriteByte(0x09)
	coinBuf.WriteByte(0x1C)
	wpkh := make([]byte, 22)
	wpkh[0] = 0x00
	wpkh[1] = 0x14
	for i := 2; i < 22; i++ {
		wpkh[i] = byte(i)
	}
	coinBuf.Write(wpkh)

	var txUndoBuf bytes.Buffer
	txUndoBuf.WriteByte(0x01)
	txUndoBuf.Write(coinBuf.Bytes())

	var blockUndoBuf bytes.Buffer
	blockUndoBuf.WriteByte(0x01)
	blockUndoBuf.Write(txUndoBuf.Bytes())

	blockData := blockUndoBuf.Bytes()
	hashBlock := make([]byte, 32)

	var out bytes.Buffer
	binary.Write(&out, binary.LittleEndian, uint32(0xD9B4BEF9))
	binary.Write(&out, binary.LittleEndian, uint32(len(blockData)))
	out.Write(blockData)
	out.Write(hashBlock)
	return out.Bytes()
}

func TestReadBlockUndos_ParsesP2WPKH(t *testing.T) {
	data := buildRevBlock(t)
	xor := XORKey(make([]byte, 8))

	tmpPath := filepath.Join(t.TempDir(), "rev_test.dat")
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	undos, err := ReadBlockUndos(tmpPath, xor)
	if err != nil {
		t.Fatalf("ReadBlockUndos: %v", err)
	}
	if len(undos) != 1 {
		t.Fatalf("got %d block undos, want 1", len(undos))
	}
	tu := undos[0].TxUndos
	if len(tu) != 1 {
		t.Fatalf("got %d tx undos, want 1", len(tu))
	}
	po := tu[0].PrevOuts
	if len(po) != 1 {
		t.Fatalf("got %d prevouts, want 1", len(po))
	}
	if po[0].Height != 100 {
		t.Errorf("height = %d, want 100", po[0].Height)
	}
	if po[0].IsCoinbase {
		t.Error("should not be coinbase")
	}
	if po[0].Value != 100_000_000 {
		t.Errorf("value = %d, want 100_000_000", po[0].Value)
	}
	if len(po[0].ScriptPubKey) == 0 {
		t.Error("script should not be empty")
	}
}

func makeBlockAtHeight(height int32) *wire.MsgBlock {
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

func TestCoinbaseHeight_ValidBIP34(t *testing.T) {
	tests := []int32{1, 100, 847493, 907139}
	for _, want := range tests {
		b := makeBlockAtHeight(want)
		got := coinbaseHeight(b)
		if got != want {
			t.Errorf("coinbaseHeight at %d = %d", want, got)
		}
	}
}

func TestCoinbaseHeight_EmptyBlock(t *testing.T) {
	hdr := wire.BlockHeader{}
	block := wire.NewMsgBlock(&hdr)
	if got := coinbaseHeight(block); got != -1 {
		t.Errorf("coinbaseHeight empty block = %d, want -1", got)
	}
}

func TestMatchUndosByHeight_ReordersToHeightOrder(t *testing.T) {
	blockA := makeBlockAtHeight(200)
	blockB := makeBlockAtHeight(100)
	blocks := []*wire.MsgBlock{blockA, blockB}

	undoFor100 := &BlockUndo{TxUndos: []TxUndo{{PrevOuts: []PrevOut{{Value: 1}}}}}
	undoFor200 := &BlockUndo{TxUndos: []TxUndo{{PrevOuts: []PrevOut{{Value: 2}}}}}
	undos := []*BlockUndo{undoFor100, undoFor200}

	result := MatchUndosByHeight(undos, blocks)

	if len(result) != 2 {
		t.Fatalf("got %d results, want 2", len(result))
	}
	if result[0] == nil || result[0].TxUndos[0].PrevOuts[0].Value != 2 {
		t.Error("block[0] (h=200) should be matched to undoFor200")
	}
	if result[1] == nil || result[1].TxUndos[0].PrevOuts[0].Value != 1 {
		t.Error("block[1] (h=100) should be matched to undoFor100")
	}
}

func TestMatchUndosByHeight_AlreadySorted(t *testing.T) {
	blockA := makeBlockAtHeight(100)
	blockB := makeBlockAtHeight(200)
	blocks := []*wire.MsgBlock{blockA, blockB}

	undoFor100 := &BlockUndo{TxUndos: []TxUndo{{}}}
	undoFor200 := &BlockUndo{TxUndos: []TxUndo{{}, {}}}
	undos := []*BlockUndo{undoFor100, undoFor200}

	result := MatchUndosByHeight(undos, blocks)

	if result[0] != undoFor100 {
		t.Error("block[0] (h=100) should keep undoFor100")
	}
	if result[1] != undoFor200 {
		t.Error("block[1] (h=200) should keep undoFor200")
	}
}

func TestBlockHeight_ExportedWrapper(t *testing.T) {
	block := makeBlockAtHeight(847493)
	got := BlockHeight(block)
	if got != 847493 {
		t.Errorf("BlockHeight = %d, want 847493", got)
	}
}

func TestBlockHeight_EmptyBlock(t *testing.T) {
	hdr := wire.BlockHeader{}
	block := wire.NewMsgBlock(&hdr)
	if got := BlockHeight(block); got != -1 {
		t.Errorf("BlockHeight empty block = %d, want -1", got)
	}
}

func TestBuildP2PKH_Length(t *testing.T) {
	hash := make([]byte, 20)
	s := buildP2PKH(hash)
	if len(s) != 25 {
		t.Errorf("buildP2PKH length = %d, want 25", len(s))
	}
	if s[0] != 0x76 || s[1] != 0xa9 || s[2] != 0x14 {
		t.Error("buildP2PKH: unexpected prefix")
	}
	if s[23] != 0x88 || s[24] != 0xac {
		t.Error("buildP2PKH: unexpected suffix")
	}
}

func TestBuildP2SH_Length(t *testing.T) {
	hash := make([]byte, 20)
	s := buildP2SH(hash)
	if len(s) != 23 {
		t.Errorf("buildP2SH length = %d, want 23", len(s))
	}
	if s[0] != 0xa9 || s[1] != 0x14 {
		t.Error("buildP2SH: unexpected prefix")
	}
	if s[22] != 0x87 {
		t.Error("buildP2SH: missing OP_EQUAL")
	}
}

func TestBuildP2PK_Compressed(t *testing.T) {
	pubkey := make([]byte, 33)
	pubkey[0] = 0x02
	s := buildP2PK(pubkey)
	if len(s) != 35 {
		t.Errorf("buildP2PK(33) length = %d, want 35", len(s))
	}
	if s[0] != 33 {
		t.Errorf("buildP2PK: first byte = %d, want 33 (push len)", s[0])
	}
	if s[34] != 0xac {
		t.Error("buildP2PK: missing OP_CHECKSIG")
	}
}

func TestBuildP2PK_Uncompressed(t *testing.T) {
	pubkey := make([]byte, 65)
	pubkey[0] = 0x04
	s := buildP2PK(pubkey)
	if len(s) != 67 {
		t.Errorf("buildP2PK(65) length = %d, want 67", len(s))
	}
}

func TestLoadXORKey_ZeroKey(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "xor.dat")
	data := make([]byte, 8)
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		t.Fatalf("write xor file: %v", err)
	}
	key, err := LoadXORKey(tmp)
	if err != nil {
		t.Fatalf("LoadXORKey: %v", err)
	}
	if len(key) != 8 {
		t.Errorf("key length = %d, want 8", len(key))
	}
	for _, b := range key {
		if b != 0 {
			t.Error("expected all-zero XOR key")
		}
	}
}

func TestLoadXORKey_NonZeroKey(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "xor.dat")
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		t.Fatalf("write xor file: %v", err)
	}
	key, err := LoadXORKey(tmp)
	if err != nil {
		t.Fatalf("LoadXORKey: %v", err)
	}
	if key[0] != 0x01 || key[7] != 0x08 {
		t.Errorf("unexpected key bytes: %v", []byte(key))
	}
}

func TestLoadXORKey_NonexistentFile(t *testing.T) {
	_, err := LoadXORKey("/nonexistent/xor.dat")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestReadBlocks_NonexistentFile(t *testing.T) {
	xor := XORKey(make([]byte, 8))
	_, err := ReadBlocks("/nonexistent/blk.dat", xor)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestReadCompactSize_FE(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteByte(0xFE)
	binary.Write(&buf, binary.LittleEndian, uint32(100_000))
	got, err := readCompactSize(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("readCompactSize 0xFE: %v", err)
	}
	if got != 100_000 {
		t.Errorf("readCompactSize 0xFE = %d, want 100000", got)
	}
}

func TestReadCompactSize_FF(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteByte(0xFF)
	binary.Write(&buf, binary.LittleEndian, uint64(5_000_000_000))
	got, err := readCompactSize(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("readCompactSize 0xFF: %v", err)
	}
	if got != 5_000_000_000 {
		t.Errorf("readCompactSize 0xFF = %d, want 5000000000", got)
	}
}

func TestReadCompressedScript_P2PKH(t *testing.T) {
	data := make([]byte, 21)
	data[0] = 0x00
	script, err := readCompressedScript(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("readCompressedScript P2PKH: %v", err)
	}
	if len(script) != 25 {
		t.Errorf("P2PKH script length = %d, want 25", len(script))
	}
	if script[0] != 0x76 {
		t.Errorf("P2PKH: unexpected first opcode 0x%02x", script[0])
	}
}

func TestReadCompressedScript_P2SH(t *testing.T) {
	data := make([]byte, 21)
	data[0] = 0x01
	script, err := readCompressedScript(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("readCompressedScript P2SH: %v", err)
	}
	if len(script) != 23 {
		t.Errorf("P2SH script length = %d, want 23", len(script))
	}
	if script[0] != 0xa9 {
		t.Errorf("P2SH: unexpected first opcode 0x%02x", script[0])
	}
}

func TestReadCompressedScript_P2PK_02(t *testing.T) {
	data := make([]byte, 33)
	data[0] = 0x02
	script, err := readCompressedScript(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("readCompressedScript P2PK 02: %v", err)
	}
	if len(script) != 35 {
		t.Errorf("P2PK-02 script length = %d, want 35", len(script))
	}
	if script[1] != 0x02 {
		t.Errorf("P2PK-02: unexpected pubkey prefix 0x%02x", script[1])
	}
}

func TestReadCompressedScript_P2PK_04(t *testing.T) {
	data := make([]byte, 33)
	data[0] = 0x04
	script, err := readCompressedScript(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("readCompressedScript P2PK 04: %v", err)
	}
	if script[1] != 0x04 {
		t.Errorf("P2PK-04: unexpected pubkey prefix 0x%02x", script[1])
	}
}

func TestReadCompressedScript_Raw(t *testing.T) {
	data := []byte{0x0A, 0x01, 0x02, 0x03, 0x04}
	script, err := readCompressedScript(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("readCompressedScript raw: %v", err)
	}
	if len(script) != 4 {
		t.Errorf("raw script length = %d, want 4", len(script))
	}
	if script[0] != 0x01 || script[3] != 0x04 {
		t.Errorf("raw script bytes = %v, unexpected", script)
	}
}

func TestMatchUndosByHeight_FallbackTxCount(t *testing.T) {
	hdr := wire.BlockHeader{}
	block := wire.NewMsgBlock(&hdr)

	cb := wire.NewMsgTx(1)
	cb.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{Index: 0xffffffff},
		SignatureScript:  []byte{0x00, 0x00},
		Sequence:         0xffffffff,
	})
	cb.AddTxOut(&wire.TxOut{PkScript: []byte{0x51}})
	block.AddTransaction(cb)

	otherTx := wire.NewMsgTx(1)
	otherTx.AddTxIn(&wire.TxIn{})
	block.AddTransaction(otherTx)

	undo := &BlockUndo{TxCount: 1}
	result := MatchUndosByHeight([]*BlockUndo{undo}, []*wire.MsgBlock{block})
	if len(result) != 1 || result[0] != undo {
		t.Errorf("fallback tx-count match should return undo[0], got %v", result[0])
	}
}

func TestCoinbaseHeight_InvalidNBytes(t *testing.T) {
	hdr := wire.BlockHeader{}
	block := wire.NewMsgBlock(&hdr)
	cb := wire.NewMsgTx(1)
	cb.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{Index: 0xffffffff},
		SignatureScript:  []byte{0x00, 0x00},
		Sequence:         0xffffffff,
	})
	cb.AddTxOut(&wire.TxOut{PkScript: []byte{0x51}})
	block.AddTransaction(cb)
	if got := coinbaseHeight(block); got != -1 {
		t.Errorf("coinbaseHeight with nBytes=0 = %d, want -1", got)
	}
}

func TestReadBlocks_EmptyFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "blk_empty.dat")
	if err := os.WriteFile(tmp, []byte{}, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	xor := XORKey(make([]byte, 8))
	blocks, err := ReadBlocks(tmp, xor)
	if err != nil {
		t.Fatalf("ReadBlocks empty file: %v", err)
	}
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks from empty file, got %d", len(blocks))
	}
}

func TestXORKey_Decode_NoOp(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	orig := make([]byte, len(data))
	copy(orig, data)

	var nilKey XORKey
	nilKey.Decode(data, 0)
	for i, b := range data {
		if b != orig[i] {
			t.Errorf("nil XORKey modified data[%d]", i)
		}
	}
}

func TestXORKey_Decode_Roundtrip(t *testing.T) {
	key := XORKey{0xAB, 0xCD, 0xEF, 0x01, 0x23, 0x45, 0x67, 0x89}
	original := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
	data := make([]byte, len(original))
	copy(data, original)

	key.Decode(data, 0)
	key.Decode(data, 0)
	for i, b := range data {
		if b != original[i] {
			t.Errorf("roundtrip failed at [%d]: got %02x, want %02x", i, b, original[i])
		}
	}
}
