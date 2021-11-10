package consensus

import (
	"fmt"
	"log"

	"github.com/korkmazkadir/bitcoin/common"
)

type MacroblockLedger struct {
	concurrencyLevel int

	// change key from puzzlesolver to height-puzzlesolver
	// because same leader can solve two puzzles
	waitingMap map[string][]common.Block
	head       *Macroblock
	blockMap   map[string]*Macroblock

	readyToDisseminate chan common.Block
}

// NewLedger creates, and initialize a leader, returns a pointer to it
func NewMacroblockLedger(concurrencyLevel int) *MacroblockLedger {
	ledger := &MacroblockLedger{

		concurrencyLevel:   concurrencyLevel,
		waitingMap:         make(map[string][]common.Block),
		blockMap:           make(map[string]*Macroblock),
		readyToDisseminate: make(chan common.Block, 1024),
	}

	// initiates the genesis block
	blocks := []common.Block{{Issuer: []byte("initial block"), Height: 0, Nonce: 12123423423435, Payload: []byte("hello world")}}

	genesisBlock := NewMacroblock(nil, blocks, 0)
	ledger.blockMap[string(genesisBlock.Hash())] = genesisBlock
	ledger.head = genesisBlock

	return ledger
}

func (l *MacroblockLedger) GetHeight() int {
	return l.head.Height()
}

func (l *MacroblockLedger) AppendBlock(block common.Block) {

	l.addToWaitingList(block)

	for l.appendWaitingBlocks() {
	}

	// here it is disseminated without any check
	// validate the block here, just before disseminating
	l.readyToDisseminate <- block
}

func (l *MacroblockLedger) GetMacroblock(height int) ([]common.Block, []byte, bool) {
	if l.head.Height() < height {
		return nil, nil, false
	}

	// find the correct height
	heightBlock := l.head
	for heightBlock.Height() != height {
		heightBlock = heightBlock.previous
	}

	return heightBlock.blocks, heightBlock.Hash(), true
}

func (l *MacroblockLedger) appendWaitingBlocks() bool {

	for key, microblocks := range l.waitingMap {

		prevBlockHash := microblocks[0].PrevBlockHash
		prevMacroblock, prevBlockAvailable := l.blockMap[string(prevBlockHash)]
		if len(microblocks) == l.concurrencyLevel && prevBlockAvailable {
			height := microblocks[0].Height
			macroblock := NewMacroblock(prevMacroblock, microblocks, height)
			l.blockMap[string(macroblock.Hash())] = macroblock

			if l.head.Height() < macroblock.Height() {
				l.head = macroblock
			}

			delete(l.waitingMap, key)
			return true
		}
	}

	return false
}

func (l *MacroblockLedger) addToWaitingList(block common.Block) {

	waitingMapKey := l.waitingMapKey(block.Height, block.PuzzleSolver)
	waitingList := l.waitingMap[waitingMapKey]

	for _, b := range waitingList {
		if b.MicroblockIndex == block.MicroblockIndex {
			panic("same microblock index already exists")
		}

		if b.Height != block.Height {
			panic("the block height is not same")
		}

		//TODO: solve this problem properly!!!
		//if !bytes.Equal(b.PrevBlockHash, block.PrevBlockHash) {
		//	panic("previous block hashes are not same")
		//}
	}

	waitingList = append(waitingList, block)

	l.waitingMap[waitingMapKey] = waitingList

}

func (l *MacroblockLedger) PrintStatus() {

	log.Println("Printing the ledger status...")

	last := l.head
	status := fmt.Sprintf("{count:%d}", len(l.blockMap))
	for last != nil {
		status = fmt.Sprintf("%s->[%x]", status, last.Hash()[:5])
		last = last.previous
	}

	log.Println(status)
}

func (l *MacroblockLedger) waitingMapKey(height int, puzzleSolver []byte) string {
	return fmt.Sprintf("%d-%x", height, puzzleSolver)
}
