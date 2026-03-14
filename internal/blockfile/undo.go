package blockfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/btcsuite/btcd/wire"
)

// PrevOut holds the previous output data decoded from a rev (undo) file.
type PrevOut struct {
	Value        int64
	ScriptPubKey []byte
	Height       uint32
	IsCoinbase   bool
}

// TxUndo is the undo data for one transaction (all its inputs).
type TxUndo struct {
	PrevOuts []PrevOut
}

// BlockUndo holds undo data for all non-coinbase transactions in one block.
type BlockUndo struct {
	TxUndos   []TxUndo
	HashBlock [32]byte // prev-block hash from the rev file
	TxCount   int      // len(TxUndos)
}

// ReadBlockUndos parses all block-undo records from a rev*.dat file.
func ReadBlockUndos(revPath string, xorKey XORKey) ([]*BlockUndo, error) {
	raw, err := os.ReadFile(revPath)
	if err != nil {
		return nil, fmt.Errorf("open rev file: %w", err)
	}

	xorKey.Decode(raw, 0)

	var undos []*BlockUndo
	r := bytes.NewReader(raw)

	for r.Len() > 0 {
		var magic uint32
		if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil, fmt.Errorf("read undo magic: %w", err)
		}

		if magic != mainnetMagic {
			break
		}

		var dataSize uint32
		if err := binary.Read(r, binary.LittleEndian, &dataSize); err != nil {
			return nil, fmt.Errorf("read undo size: %w", err)
		}

		if dataSize == 0 || uint64(dataSize) > uint64(r.Len()) {
			break
		}

		data := make([]byte, dataSize)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, fmt.Errorf("read undo data: %w", err)
		}

		var hashBlock [32]byte
		if _, err := io.ReadFull(r, hashBlock[:]); err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				return nil, fmt.Errorf("read hashBlock: %w", err)
			}
		}

		blockUndo, err := parseBlockUndo(data)
		if err != nil {
			undos = append(undos, &BlockUndo{HashBlock: hashBlock})
			continue
		}
		blockUndo.HashBlock = hashBlock
		blockUndo.TxCount = len(blockUndo.TxUndos)
		undos = append(undos, blockUndo)
	}

	return undos, nil
}

// MatchUndosByHeight matches undo records to blocks by height, falling back to tx-count.
// Unmatched positions are nil.
func MatchUndosByHeight(undos []*BlockUndo, blocks []*wire.MsgBlock) []*BlockUndo {
	result := make([]*BlockUndo, len(blocks))

	type blockWithHeight struct {
		blkIdx int
		height int32
	}
	sorted := make([]blockWithHeight, 0, len(blocks))
	for i, block := range blocks {
		h := coinbaseHeight(block)
		sorted = append(sorted, blockWithHeight{i, h})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].height < sorted[j].height
	})

	heightOK := true
	for _, bwh := range sorted {
		if bwh.height <= 0 {
			heightOK = false
			break
		}
	}
	if heightOK && len(undos) == len(blocks) {
		for sortedIdx, bwh := range sorted {
			if sortedIdx < len(undos) {
				result[bwh.blkIdx] = undos[sortedIdx]
			}
		}
		return result
	}

	txCountToUndos := make(map[int][]int, len(undos))
	for i, u := range undos {
		if u != nil {
			txCountToUndos[u.TxCount] = append(txCountToUndos[u.TxCount], i)
		}
	}
	used := make(map[int]bool, len(undos))
	for blockIdx, block := range blocks {
		wantTxCount := len(block.Transactions) - 1
		if candidates, ok := txCountToUndos[wantTxCount]; ok {
			for _, ui := range candidates {
				if !used[ui] {
					result[blockIdx] = undos[ui]
					used[ui] = true
					break
				}
			}
		}
	}
	return result
}

// BlockHeight extracts the block height from the coinbase BIP34 scriptSig.
func BlockHeight(block *wire.MsgBlock) int32 {
	return coinbaseHeight(block)
}

// coinbaseHeight extracts the block height from the coinbase BIP34 scriptSig.
func coinbaseHeight(block *wire.MsgBlock) int32 {
	if len(block.Transactions) == 0 || len(block.Transactions[0].TxIn) == 0 {
		return -1
	}
	script := block.Transactions[0].TxIn[0].SignatureScript
	if len(script) < 2 {
		return -1
	}
	nBytes := int(script[0])
	if nBytes == 0 || nBytes > 4 || nBytes+1 > len(script) {
		return -1
	}
	var buf [4]byte
	copy(buf[:], script[1:1+nBytes])
	return int32(binary.LittleEndian.Uint32(buf[:]))
}

func parseBlockUndo(data []byte) (*BlockUndo, error) {
	r := bytes.NewReader(data)

	txCount, err := readCompactSize(r)
	if err != nil {
		return nil, fmt.Errorf("read vtxundo count: %w", err)
	}

	bu := &BlockUndo{TxUndos: make([]TxUndo, 0, txCount)}

	for i := uint64(0); i < txCount; i++ {
		inputCount, err := readCompactSize(r)
		if err != nil {
			return bu, fmt.Errorf("read vprevout count for tx %d: %w", i, err)
		}

		tu := TxUndo{PrevOuts: make([]PrevOut, 0, inputCount)}

		for j := uint64(0); j < inputCount; j++ {
			po, err := readCoin(r)
			if err != nil {
				// Partial read — accept what we have
				tu.PrevOuts = append(tu.PrevOuts, PrevOut{})
				continue
			}
			tu.PrevOuts = append(tu.PrevOuts, po)
		}

		bu.TxUndos = append(bu.TxUndos, tu)
	}

	return bu, nil
}

// readCoin reads one Coin using Bitcoin Core's internal serialization.
func readCoin(r io.Reader) (PrevOut, error) {
	code, err := readVarInt(r)
	if err != nil {
		return PrevOut{}, fmt.Errorf("read coin code: %w", err)
	}

	height := uint32(code >> 1)

	// coinbase flag is stored as a separate byte in v27+ format
	var cbBuf [1]byte
	if _, err := io.ReadFull(r, cbBuf[:]); err != nil {
		return PrevOut{}, fmt.Errorf("read coin coinbase flag: %w", err)
	}
	isCoinbase := cbBuf[0] != 0

	compressedAmt, err := readVarInt(r)
	if err != nil {
		return PrevOut{}, fmt.Errorf("read coin amount: %w", err)
	}

	value := decompressAmount(compressedAmt)

	script, err := readCompressedScript(r)
	if err != nil {
		return PrevOut{}, fmt.Errorf("read coin script: %w", err)
	}

	return PrevOut{
		Value:        value,
		ScriptPubKey: script,
		Height:       height,
		IsCoinbase:   isCoinbase,
	}, nil
}

// readVarInt reads Bitcoin Core's 7-bit VarInt (MSB-first, bit7 = more bytes follow).
func readVarInt(r io.Reader) (uint64, error) {
	var n uint64
	buf := make([]byte, 1)
	for {
		if _, err := io.ReadFull(r, buf); err != nil {
			return 0, err
		}
		b := buf[0]
		n = (n << 7) | uint64(b&0x7F)
		if b&0x80 != 0 {
			n++
		} else {
			return n, nil
		}
	}
}

// readCompactSize reads a Bitcoin CompactSize (Wire VarInt) from r.
func readCompactSize(r io.Reader) (uint64, error) {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	switch buf[0] {
	case 0xFD:
		var v uint16
		if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
			return 0, err
		}
		return uint64(v), nil
	case 0xFE:
		var v uint32
		if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
			return 0, err
		}
		return uint64(v), nil
	case 0xFF:
		var v uint64
		if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
			return 0, err
		}
		return v, nil
	default:
		return uint64(buf[0]), nil
	}
}

// readCompressedScript reads a Bitcoin Core compressed script.
// nSize 0–5 are special types; nSize ≥ 6 means raw bytes of length (nSize-6).
func readCompressedScript(r io.Reader) ([]byte, error) {
	nSize, err := readVarInt(r)
	if err != nil {
		return nil, fmt.Errorf("read script nsize: %w", err)
	}

	const nSpecial = 6

	switch {
	case nSize == 0: // P2PKH
		hash := make([]byte, 20)
		if _, err := io.ReadFull(r, hash); err != nil {
			return nil, err
		}
		return buildP2PKH(hash), nil

	case nSize == 1: // P2SH
		hash := make([]byte, 20)
		if _, err := io.ReadFull(r, hash); err != nil {
			return nil, err
		}
		return buildP2SH(hash), nil

	case nSize == 2 || nSize == 3: // P2PK compressed
		keyData := make([]byte, 32)
		if _, err := io.ReadFull(r, keyData); err != nil {
			return nil, err
		}
		pubkey := make([]byte, 33)
		pubkey[0] = byte(nSize) // 0x02 or 0x03
		copy(pubkey[1:], keyData)
		return buildP2PK(pubkey), nil

	case nSize == 4 || nSize == 5: // P2PK uncompressed (compressed storage)
		keyData := make([]byte, 32)
		if _, err := io.ReadFull(r, keyData); err != nil {
			return nil, err
		}
		pubkey := make([]byte, 33)
		pubkey[0] = 0x04 // best effort: uncompressed-derived
		copy(pubkey[1:], keyData)
		return buildP2PK(pubkey), nil

	default: // raw script
		scriptLen := nSize - nSpecial
		if scriptLen > 4_000_000 { // sanity cap (Bitcoin block size limit)
			return nil, fmt.Errorf("script length %d exceeds block size limit", scriptLen)
		}
		script := make([]byte, scriptLen)
		if _, err := io.ReadFull(r, script); err != nil {
			return nil, err
		}
		return script, nil
	}
}

func buildP2PKH(hash []byte) []byte {
	// OP_DUP OP_HASH160 PUSH20 <hash> OP_EQUALVERIFY OP_CHECKSIG
	s := make([]byte, 25)
	s[0] = 0x76 // OP_DUP
	s[1] = 0xa9 // OP_HASH160
	s[2] = 0x14 // push 20 bytes
	copy(s[3:], hash)
	s[23] = 0x88 // OP_EQUALVERIFY
	s[24] = 0xac // OP_CHECKSIG
	return s
}

func buildP2SH(hash []byte) []byte {
	// OP_HASH160 PUSH20 <hash> OP_EQUAL
	s := make([]byte, 23)
	s[0] = 0xa9 // OP_HASH160
	s[1] = 0x14 // push 20 bytes
	copy(s[2:], hash)
	s[22] = 0x87 // OP_EQUAL
	return s
}

func buildP2PK(pubkey []byte) []byte {
	// PUSH<len> <pubkey> OP_CHECKSIG
	s := make([]byte, len(pubkey)+2)
	s[0] = byte(len(pubkey))
	copy(s[1:], pubkey)
	s[len(s)-1] = 0xac // OP_CHECKSIG
	return s
}

// decompressAmount converts a Bitcoin Core compressed amount to satoshis.
func decompressAmount(x uint64) int64 {
	if x == 0 {
		return 0
	}
	x--
	e := x % 10
	x /= 10
	var n uint64
	if e < 9 {
		d := (x % 9) + 1
		x /= 9
		n = x*10 + d
	} else {
		n = x + 1
	}
	for e > 0 {
		n *= 10
		e--
	}
	return int64(n)
}
