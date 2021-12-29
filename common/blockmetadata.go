package common

type BlockMetadata struct {
	Height        int
	Siblings      [][]byte
	PrevBlockHash []byte
	PayloadSize   int
	index         int
	hash          []byte
}

func NewBlockBlockMetadata(height int, blockHash []byte, microblockIndex int, siblings [][]byte, prevBlockHash []byte, payloadSize int) BlockMetadata {
	return BlockMetadata{
		Height:        height,
		Siblings:      siblings,
		PrevBlockHash: prevBlockHash,
		PayloadSize:   payloadSize,
		index:         microblockIndex,
		hash:          blockHash,
	}
}

func (b BlockMetadata) Hash() []byte {
	return b.hash
}

func (b BlockMetadata) Index() int {
	return b.index
}
