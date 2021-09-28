package consensus

import (
	"bytes"
	"math"
	"math/rand"
	"testing"

	"github.com/korkmazkadir/bitcoin/common"
)

func createBlock(round int, previousBlockHashes [][]byte, blockSize int, leaderCount int) common.Block {

	payloadSize := int(math.Ceil(float64(blockSize) / float64(leaderCount)))

	block := common.Block{
		Height:          round,
		Payload:         getRandomByteSlice(payloadSize),
		PrevBlockHashes: previousBlockHashes,
	}

	return block
}

func getRandomByteSlice(size int) []byte {
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return data
}

func TestLedger(t *testing.T) {

	ledger := NewLedger(1)

	genesisBlock, _ := ledger.GetMacroBlock(0)

	b1 := createBlock(1, [][]byte{genesisBlock[0].Hash()}, 100000, 1)

	ledger.AppendBlock(b1)

	b1r, ok := ledger.GetMacroBlock(1)

	if !ok {
		t.Errorf("block is not available")
	}

	if !bytes.Equal(b1.Hash(), b1r[0].Hash()) {
		t.Errorf("retreived block is faulty")
	}

	blockToDisseminate := <-ledger.readyToDisseminate

	if !bytes.Equal(b1.Hash(), blockToDisseminate.Hash()) {
		t.Errorf("retreived block is faulty")
	}

	b2 := createBlock(2, [][]byte{b1.Hash()}, 100000, 1)

	ledger.AppendBlock(b2)

	b2r, ok := ledger.GetMacroBlock(2)

	if !ok {
		t.Errorf("block is not available")
	}

	if !bytes.Equal(b2.Hash(), b2r[0].Hash()) {
		t.Errorf("retreived block is faulty")
	}
}
