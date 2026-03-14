package blockfile

import "os"

// XORKey is the 8-byte obfuscation key stored in xor.dat.
// A zero-length or all-zero key is a no-op.
type XORKey []byte

// LoadXORKey reads the XOR key from the given file path.
func LoadXORKey(path string) (XORKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return XORKey(data), nil
}

// Decode XOR-decodes src in-place from the given file offset.
func (k XORKey) Decode(src []byte, startOffset int64) {
	if len(k) == 0 {
		return
	}
	for i := range src {
		src[i] ^= k[(int64(i)+startOffset)%int64(len(k))]
	}
}
