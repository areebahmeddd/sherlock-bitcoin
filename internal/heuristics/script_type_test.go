package heuristics

import (
	"bytes"
	"testing"

	"github.com/btcsuite/btcd/wire"
)

func TestScriptType_P2WPKH(t *testing.T) {
	script := make([]byte, 22)
	script[0] = 0x00
	script[1] = 0x14
	if got := ScriptType(script); got != "p2wpkh" {
		t.Errorf("ScriptType P2WPKH = %q, want p2wpkh", got)
	}
}

func TestScriptType_P2PKH(t *testing.T) {
	script := []byte{0x76, 0xa9, 0x14}
	script = append(script, make([]byte, 20)...)
	script = append(script, 0x88, 0xac)
	if got := ScriptType(script); got != "p2pkh" {
		t.Errorf("ScriptType P2PKH = %q, want p2pkh", got)
	}
}

func TestScriptType_OPReturn(t *testing.T) {
	script := []byte{0x6a, 0x04, 0xde, 0xad, 0xbe, 0xef}
	if got := ScriptType(script); got != "op_return" {
		t.Errorf("ScriptType OP_RETURN = %q, want op_return", got)
	}
}

func TestScriptType_Empty(t *testing.T) {
	if got := ScriptType(nil); got != "unknown" {
		t.Errorf("ScriptType nil = %q, want unknown", got)
	}
}

func TestScriptType_P2SH(t *testing.T) {
	script := []byte{0xa9, 0x14}
	script = append(script, make([]byte, 20)...)
	script = append(script, 0x87)
	if got := ScriptType(script); got != "p2sh" {
		t.Errorf("ScriptType P2SH = %q, want p2sh", got)
	}
}

func TestScriptType_P2WSH(t *testing.T) {
	script := make([]byte, 34)
	script[0] = 0x00
	script[1] = 0x20
	if got := ScriptType(script); got != "p2wsh" {
		t.Errorf("ScriptType P2WSH = %q, want p2wsh", got)
	}
}

func TestScriptType_P2TR(t *testing.T) {
	script := make([]byte, 34)
	script[0] = 0x51
	script[1] = 0x20
	if got := ScriptType(script); got != "p2tr" {
		t.Errorf("ScriptType P2TR = %q, want p2tr", got)
	}
}

func TestInferInputScriptType_P2WPKH(t *testing.T) {
	in := wire.NewTxIn(&wire.OutPoint{}, nil, wire.TxWitness{
		make([]byte, 72),
		make([]byte, 33),
	})
	got := inferInputScriptType(in)
	if got != "p2wpkh" {
		t.Errorf("inferInputScriptType P2WPKH = %q, want p2wpkh", got)
	}
}

func TestInferInputScriptType_P2TR(t *testing.T) {
	in := wire.NewTxIn(&wire.OutPoint{}, nil, wire.TxWitness{
		make([]byte, 64),
	})
	got := inferInputScriptType(in)
	if got != "p2tr" {
		t.Errorf("inferInputScriptType P2TR = %q, want p2tr", got)
	}
}

func TestInferInputScriptType_P2WSH(t *testing.T) {
	witnessScript := make([]byte, 35)
	in := wire.NewTxIn(&wire.OutPoint{}, nil, wire.TxWitness{
		make([]byte, 72),
		make([]byte, 33),
		witnessScript,
	})
	got := inferInputScriptType(in)
	if got != "p2wsh" && got != "p2tr" {
		t.Errorf("inferInputScriptType P2WSH = %q, want p2wsh or p2tr", got)
	}
}

func TestInferInputScriptType_P2SH_WrappedSegwit(t *testing.T) {
	scriptSig := []byte{0x16, 0x00, 0x14, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	in := &wire.TxIn{
		SignatureScript: scriptSig,
		Witness:         wire.TxWitness{make([]byte, 72), make([]byte, 33)},
	}
	got := inferInputScriptType(in)
	if got != "p2sh" {
		t.Errorf("inferInputScriptType P2SH-wrapped = %q, want p2sh", got)
	}
}

func TestInferInputScriptType_LegacyP2PKH_TwoItems(t *testing.T) {
	var scriptSig []byte
	sig := make([]byte, 71)
	sig[0] = 0x30
	scriptSig = append(scriptSig, byte(len(sig)))
	scriptSig = append(scriptSig, sig...)
	pubkey := make([]byte, 33)
	pubkey[0] = 0x02
	scriptSig = append(scriptSig, byte(len(pubkey)))
	scriptSig = append(scriptSig, pubkey...)

	in := &wire.TxIn{SignatureScript: scriptSig}
	got := inferInputScriptType(in)
	if got != "p2pkh" {
		t.Errorf("inferInputScriptType legacy 2-item = %q, want p2pkh", got)
	}
}

func TestInferInputScriptType_EmptyInput(t *testing.T) {
	in := &wire.TxIn{}
	got := inferInputScriptType(in)
	if got != "unknown" {
		t.Errorf("inferInputScriptType empty = %q, want unknown", got)
	}
}

func TestParseScriptPushes_SinglePush(t *testing.T) {
	data := []byte{0x04, 0xde, 0xad, 0xbe, 0xef}
	r := bytes.NewReader(data)
	items, err := parseScriptPushes(r)
	if err != nil {
		t.Fatalf("parseScriptPushes: %v", err)
	}
	if len(items) != 1 || len(items[0]) != 4 {
		t.Errorf("expected 1 item of len 4, got %d items", len(items))
	}
}

func TestScriptType_P2PK(t *testing.T) {
	script := make([]byte, 35)
	script[0] = 0x21
	script[1] = 0x02
	script[34] = 0xac
	if got := ScriptType(script); got != "p2pk" {
		t.Errorf("ScriptType P2PK = %q, want p2pk", got)
	}
}

func TestParseScriptPushes_PUSHDATA1(t *testing.T) {
	data := []byte{0x4c, 0x03, 0xaa, 0xbb, 0xcc}
	r := bytes.NewReader(data)
	items, err := parseScriptPushes(r)
	if err != nil {
		t.Fatalf("parseScriptPushes PUSHDATA1: %v", err)
	}
	if len(items) != 1 || len(items[0]) != 3 {
		t.Errorf("expected 1 item of len 3, got %d items", len(items))
	}
}

func TestParseScriptPushes_PUSHDATA2(t *testing.T) {
	data := []byte{0x4d, 0x02, 0x00, 0xde, 0xad}
	r := bytes.NewReader(data)
	items, err := parseScriptPushes(r)
	if err != nil {
		t.Fatalf("parseScriptPushes PUSHDATA2: %v", err)
	}
	if len(items) != 1 || len(items[0]) != 2 {
		t.Errorf("expected 1 item of len 2, got %d items", len(items))
	}
	if items[0][0] != 0xde || items[0][1] != 0xad {
		t.Errorf("unexpected push data: %v", items[0])
	}
}

func TestParseScriptPushes_OpcodeSkipped(t *testing.T) {
	data := []byte{0x51, 0x01, 0xff}
	r := bytes.NewReader(data)
	items, err := parseScriptPushes(r)
	if err != nil {
		t.Fatalf("parseScriptPushes opcode skip: %v", err)
	}
	if len(items) != 1 || items[0][0] != 0xff {
		t.Errorf("expected 1 push item [0xff], got %v", items)
	}
}

func TestParseScriptPushes_OP0(t *testing.T) {
	data := []byte{0x00, 0x01, 0xff}
	r := bytes.NewReader(data)
	items, err := parseScriptPushes(r)
	if err != nil {
		t.Fatalf("parseScriptPushes OP0: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if len(items[0]) != 0 {
		t.Errorf("item[0] should be empty (OP_0), got %v", items[0])
	}
	if items[1][0] != 0xff {
		t.Errorf("item[1] should be [0xff], got %v", items[1])
	}
}

func TestInferInputScriptType_SingleItemScriptSig(t *testing.T) {
	scriptSig := make([]byte, 73)
	scriptSig[0] = 72
	scriptSig[1] = 0x30
	in := &wire.TxIn{SignatureScript: scriptSig}
	if got := inferInputScriptType(in); got != "p2pkh" {
		t.Errorf("inferInputScriptType 1-item sig = %q, want p2pkh", got)
	}
}

func TestParseScriptPushes_OP_PUSHDATA1(t *testing.T) {
	payload := make([]byte, 10)
	data := append([]byte{0x4C, 0x0a}, payload...)
	r := bytes.NewReader(data)
	items, err := parseScriptPushes(r)
	if err != nil {
		t.Fatalf("parseScriptPushes: %v", err)
	}
	if len(items) != 1 || len(items[0]) != 10 {
		t.Errorf("expected 1 item of len 10, got items=%v", items)
	}
}

func TestParseScriptPushes_OP_PUSHDATA2(t *testing.T) {
	payload := make([]byte, 5)
	data := []byte{0x4D, 0x05, 0x00}
	data = append(data, payload...)
	r := bytes.NewReader(data)
	items, err := parseScriptPushes(r)
	if err != nil {
		t.Fatalf("parseScriptPushes: %v", err)
	}
	if len(items) != 1 || len(items[0]) != 5 {
		t.Errorf("expected 1 item of len 5, got items=%v", items)
	}
}

func TestParseScriptPushes_OP_0(t *testing.T) {
	r := bytes.NewReader([]byte{0x00})
	items, err := parseScriptPushes(r)
	if err != nil {
		t.Fatalf("parseScriptPushes: %v", err)
	}
	if len(items) != 1 || len(items[0]) != 0 {
		t.Errorf("expected 1 empty item, got %v", items)
	}
}

func TestParseScriptPushes_Opcode(t *testing.T) {
	r := bytes.NewReader([]byte{0x52})
	items, err := parseScriptPushes(r)
	if err != nil {
		t.Fatalf("parseScriptPushes: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("opcode should be skipped, got %d items", len(items))
	}
}

func TestParseScriptPushes_Empty(t *testing.T) {
	r := bytes.NewReader([]byte{})
	items, err := parseScriptPushes(r)
	if err != nil {
		t.Fatalf("parseScriptPushes: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("empty input should yield 0 items, got %d", len(items))
	}
}
