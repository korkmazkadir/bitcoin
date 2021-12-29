package common

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// Block defines blockchain block structure
type Block struct {
	Issuer []byte

	PrevBlockHash []byte

	Height int

	Nonce int64

	// contains the hashes of sibling blocks
	Siblings [][]byte

	Signature []byte

	Payload []byte

	// time in seconds
	Timestamp int64

	HopCount int

	receiveTime int64
}

func (b *Block) SetEnqueueTime() {
	b.receiveTime = time.Now().UnixMilli()
}

func (b Block) GetEnqueueElapsedTime() int {

	return int(time.Now().UnixMilli() - b.receiveTime)
}

// Hash produces the digest of a Block.
// It considers all fields of a Block.
func (b Block) Hash() []byte {

	str := fmt.Sprintf("%x,%x,%d,%d,%x", b.Issuer, b.PrevBlockHash, b.Height, b.Nonce, b.Payload)
	h := sha256.New()
	_, err := h.Write([]byte(str))
	if err != nil {
		panic(err)
	}

	return h.Sum(nil)
}
