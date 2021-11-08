package consensus

import (
	"crypto/sha256"
	"sort"

	"github.com/korkmazkadir/bitcoin/common"
)

type Macroblock struct {
	previous *Macroblock
	height   int
	blocks   []common.Block
	hash     []byte
}

func NewMacroblock(previousMacroblock *Macroblock, blocks []common.Block, height int) *Macroblock {

	// sort microblocks to have a consistent hashing
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].MicroblockIndex < blocks[j].MicroblockIndex
	})

	m := &Macroblock{
		previous: previousMacroblock,
		height:   height,
		blocks:   blocks,
		hash:     calculateHash(blocks),
	}

	return m
}

func (m *Macroblock) Hash() []byte {

	if len(m.hash) == 0 {
		m.hash = calculateHash(m.blocks)
	}

	return m.hash
}

func (m *Macroblock) Height() int {
	return m.height
}

func calculateHash(blocks []common.Block) []byte {

	h := sha256.New()
	for i := 0; i < len(blocks); i++ {
		_, err := h.Write(blocks[i].Hash())
		if err != nil {
			panic(err)
		}
	}

	return h.Sum(nil)
}
