package common

import (
	"crypto/sha256"
	"fmt"
)

type SubleaderRequest struct {
	PuzzleSolver    []byte
	MicroblockIndex int
	Height          int
}

// Block defines blockchain block structure
type Block struct {
	Issuer []byte

	PuzzleSolver []byte

	MicroblockIndex int

	PrevBlockHash []byte

	Height int

	Nonce int64

	Payload []byte
}

func handleWriteError(err error) {
	if err != nil {
		panic(err)
	}
}

// Hash produces the digest of a Block.
// It considers all fields of a Block.
func (b Block) Hash() []byte {

	// handles non byte fields
	str := fmt.Sprintf("%d,%d,%d", b.MicroblockIndex, b.Height, b.Nonce)
	h := sha256.New()

	_, err := h.Write([]byte(str))
	handleWriteError(err)

	_, err = h.Write(b.Issuer)
	handleWriteError(err)
	_, err = h.Write(b.PuzzleSolver)
	handleWriteError(err)
	_, err = h.Write(b.PrevBlockHash)
	handleWriteError(err)
	_, err = h.Write(b.Payload)
	handleWriteError(err)

	return h.Sum(nil)
}
