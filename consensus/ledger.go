package consensus

import (
	"fmt"
	"log"

	"github.com/korkmazkadir/bitcoin/common"
)

type Ledger struct {
	concurrencyLevel int

	waitList           []common.Block
	blockMap           map[int][]common.Block
	readyToDisseminate chan common.Block
}

// NewLedger creates, and initialize a leader, returns a pointer to it
func NewLedger(concurrencyLevel int) *Ledger {
	ledger := &Ledger{
		concurrencyLevel:   concurrencyLevel,
		blockMap:           make(map[int][]common.Block),
		readyToDisseminate: make(chan common.Block, 1024),
	}

	// initiates the genesis block
	ledger.blockMap[0] = []common.Block{{Issuer: []byte("initial block"), Height: 0, Nonce: 12123423423435, Payload: []byte("hello world")}}

	return ledger
}

// AppendBlock thy to append the given block to the ledger
func (l *Ledger) AppendBlock(block common.Block) {

	appendResult := l.append(block)

	// could not append the block so nothing todo
	if !appendResult {
		log.Printf("putting a block to the waiting list. Block height is %d\n", block.Height)
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

	log.Printf("Appended:\t\t%x\n", block.Hash())
}

// GetMinedBlock returns true with a  list of microblocks if all the microblocks for a specific height are available otherwise returns false
func (l *Ledger) GetMacroBlock(height int) ([]common.Block, bool) {

	heightBlocks, ok := l.blockMap[height]

	// returns the genesis block
	if height == 0 {
		if len(heightBlocks) != 1 {
			panic("no genesis block")
		}

		return heightBlocks, true
	}

	// there is no block so return false
	if !ok {
		return []common.Block{}, false
	}

	blocks := make([]common.Block, l.concurrencyLevel)
	microblockIndexes := make([]bool, l.concurrencyLevel)
	for _, block := range heightBlocks {
		microblockIndexes[block.Nonce%int64(l.concurrencyLevel)] = true
		blocks[block.Nonce%int64(l.concurrencyLevel)] = block
	}

	for _, isMicroBlockAvailable := range microblockIndexes {
		if !isMicroBlockAvailable {
			return []common.Block{}, false
		}
	}

	return blocks, true
}

func (l *Ledger) GetMicroblock(height int, macroblockIndex int) (common.Block, bool) {

	heightBlocks, ok := l.blockMap[height]

	// there is no block so return false
	if !ok {
		return common.Block{}, false
	}

	for _, block := range heightBlocks {
		if (block.Nonce % int64(l.concurrencyLevel)) == int64(macroblockIndex) {
			return block, true
		}
	}

	return common.Block{}, false
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

func (l *Ledger) PrintStatus() {

	genesisBlock, _ := l.GetMacroBlock(0)
	status := fmt.Sprintf("[0 | %d]", len(genesisBlock))
	for i := 1; ; i++ {

		mb, ok := l.blockMap[i]

		if !ok {
			break
		}

		status = fmt.Sprintf("%s <-[%d | %d]", status, i, len(mb))
	}

	log.Println(status)

}
