package common

import (
	"crypto/sha256"
	"fmt"
)

// Block defines blockchain block structure
type Block struct {
	Issuer []byte

	PrevBlockHashes [][]byte

	Height int

	Nonce int64

	Payload []byte
}

// Hash produces the digest of a Block.
// It considers all fields of a Block.
func (b Block) Hash() []byte {

	str := fmt.Sprintf("%x,%x,%d,%x", b.Issuer, b.PrevBlockHashes, b.Height, b.Payload)
	h := sha256.New()
	_, err := h.Write([]byte(str))
	if err != nil {
		panic(err)
	}

	return h.Sum(nil)
}
