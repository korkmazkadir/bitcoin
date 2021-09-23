package consensus

import (
	"fmt"

	"github.com/korkmazkadir/bitcoin/common"
)

type Ledger struct {
	concurrencyLevel int

	waitList           []common.Block
	blockMap           map[int][]common.Block
	readyToDisseminate chan common.Block
}

//TODO: create genesisblocks in a determinicstic way
// NewLedger creates, and initialize a leader, returns a pointer to it
func NewLedger(concurrencyLevel int) *Ledger {
	ledger := &Ledger{
		concurrencyLevel:   concurrencyLevel,
		blockMap:           make(map[int][]common.Block),
		readyToDisseminate: make(chan common.Block, 1024),
	}
	return ledger
}

// AppendBlock thy to append the given block to the ledger
func (l *Ledger) AppendBlock(block common.Block) {

	appendResult := l.append(block)

	// could not append the block so nothing todo
	if !appendResult {

		l.waitList = append(l.waitList, block)
		return
	}

	// trying to append waiting blocks.
	for appendResult {

		appendResult = false
		for _, wb := range l.waitList {
			// when you append a waiting block
			// you should retry remaning blocks to append
			appendResult = appendResult || l.append(wb)
		}
	}

}

// IsBlockMined returns true if all the microblocks for a specific height are available otherwose returns false
func (l *Ledger) IsBlockMined(height int) bool {

	heightBlocks, ok := l.blockMap[height]

	// there is no block so return false
	if !ok {
		return false
	}

	microblockIndexes := make([]bool, l.concurrencyLevel)
	for _, block := range heightBlocks {
		microblockIndexes[block.Nonce%int64(l.concurrencyLevel)] = true
	}

	for _, isMicroBlockAvailable := range microblockIndexes {
		if !isMicroBlockAvailable {
			return false
		}
	}

	return true
}

// BlockHashesToMine returns block hash to mine on top of them
func (l *Ledger) BlockHashesToMine(height int) [][]byte {

	heightBlocks, ok := l.blockMap[height]
	// there is no block so return false
	if !ok {
		panic(fmt.Errorf("there are no blocks to mine on top of"))
	}

	microblockHashes := make([][]byte, l.concurrencyLevel)
	blockHashCount := 0

	for _, block := range heightBlocks {

		microBlockIndex := block.Nonce % int64(l.concurrencyLevel)

		if len(microblockHashes[microBlockIndex]) == 0 {
			microblockHashes[microBlockIndex] = block.Hash()
			blockHashCount++
		}
	}

	if blockHashCount != l.concurrencyLevel {
		panic(fmt.Errorf("there are missing microblocks"))
	}

	return microblockHashes
}

func (l *Ledger) append(block common.Block) bool {

	previousRoundBlocks, ok := l.blockMap[block.Height-1]

	if !ok {
		// retuning because all of the previous blocks are missing
		return false
	}

	// creates an available block map
	availableBlocks := make(map[string]struct{})
	for _, b := range previousRoundBlocks {
		availableBlocks[string(b.Hash())] = struct{}{}
	}

	// checks for all prev block hashes
	for _, h := range block.PrevBlockHashes {
		_, ok := availableBlocks[string(h)]
		if !ok {
			// returning because one of the prev blocks is missing!!!
			return false
		}
	}

	//TODO: validate block, and simulate the cost of validation here

	// apending block top the ledger
	currentRoundBlocks := l.blockMap[block.Height]
	currentRoundBlocks = append(currentRoundBlocks, block)
	l.blockMap[block.Height] = currentRoundBlocks

	// the block is validated, and appended to the ledger.
	// the node should disseminate it
	l.readyToDisseminate <- block

	return true
}
