package blockfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/btcsuite/btcd/wire"
)

const (
	mainnetMagic = uint32(0xD9B4BEF9) // little-endian: f9 be b4 d9
)

// ReadBlocks reads all blocks from a blk*.dat file and returns them in file order.
func ReadBlocks(blkPath string, xorKey XORKey) ([]*wire.MsgBlock, error) {
	raw, err := os.ReadFile(blkPath)
	if err != nil {
		return nil, fmt.Errorf("open blk file: %w", err)
	}

	xorKey.Decode(raw, 0)

	var blocks []*wire.MsgBlock
	r := bytes.NewReader(raw)

	for r.Len() > 0 {
		var magic uint32
		if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil, fmt.Errorf("read magic: %w", err)
		}

		if magic != mainnetMagic {
			// Skip trailing zeros or unknown data in the file
			// Try to re-sync by scanning for the magic
			break
		}

		var blockSize uint32
		if err := binary.Read(r, binary.LittleEndian, &blockSize); err != nil {
			return nil, fmt.Errorf("read block size: %w", err)
		}

		if blockSize == 0 || uint64(blockSize) > uint64(r.Len()) {
			break
		}

		blockData := make([]byte, blockSize)
		if _, err := io.ReadFull(r, blockData); err != nil {
			return nil, fmt.Errorf("read block data: %w", err)
		}

		block := &wire.MsgBlock{}
		if err := block.Deserialize(bytes.NewReader(blockData)); err != nil {
			return nil, fmt.Errorf("deserialize block: %w", err)
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}
