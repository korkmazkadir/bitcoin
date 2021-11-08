package consensus

import (
	"bytes"
	"fmt"
	"log"

	"github.com/korkmazkadir/bitcoin/common"
)

type MacroblockLedger struct {
	concurrencyLevel int

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

	for puzzleSolver, microblocks := range l.waitingMap {

		prevBlockHash := microblocks[0].PrevBlockHash
		prevMacroblock, prevBlockAvailable := l.blockMap[string(prevBlockHash)]
		if len(microblocks) == l.concurrencyLevel && prevBlockAvailable {
			height := microblocks[0].Height
			macroblock := NewMacroblock(prevMacroblock, microblocks, height)
			l.blockMap[string(macroblock.Hash())] = macroblock

			if l.head.Height() < macroblock.Height() {
				l.head = macroblock
			}

			delete(l.waitingMap, puzzleSolver)
			return true
		}
	}

	return false
}

func (l *MacroblockLedger) addToWaitingList(block common.Block) []common.Block {

	puzzleSolver := block.PuzzleSolver
	waitingList := l.waitingMap[string(puzzleSolver)]

	for _, b := range waitingList {
		if b.MicroblockIndex == block.MicroblockIndex {
			panic("same microblock index already exists")
		}

		if b.Height != block.Height {
			panic("the block height is not same")
		}

		if bytes.Equal(b.PrevBlockHash, block.PrevBlockHash) {
			panic("previous block hashes are not same")
		}
	}

	waitingList = append(waitingList, block)

	l.waitingMap[string(puzzleSolver)] = waitingList

	return waitingList
}

func (l *MacroblockLedger) PrintStatus() {

	last := l.head
	status := fmt.Sprintf("{count:%d}", len(l.blockMap))
	for last != nil {
		status = fmt.Sprintf("%s->[%x]", status, last.Hash()[:5])
	}

	log.Println(status)
}
