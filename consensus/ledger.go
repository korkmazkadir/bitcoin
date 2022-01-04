package consensus

import (
	"fmt"
	"log"
	"time"

	"github.com/korkmazkadir/bitcoin/common"
)

type Ledger struct {
	concurrencyLevel int

	waitList            []common.Block
	blockMap            map[int][]common.BlockMetadata
	availablePrevBlocks map[string]struct{}
	readyToDisseminate  chan common.Block
}

// NewLedger creates, and initialize a leader, returns a pointer to it
func NewLedger(concurrencyLevel int) *Ledger {
	ledger := &Ledger{
		concurrencyLevel:    concurrencyLevel,
		blockMap:            make(map[int][]common.BlockMetadata),
		availablePrevBlocks: make(map[string]struct{}),
		readyToDisseminate:  make(chan common.Block, 1024),
	}

	// initiates the genesis block
	gb := common.Block{Issuer: []byte("initial block"), Height: 0, Nonce: 12123423423435, Payload: []byte("hello world")}
	ledger.blockMap[0] = []common.BlockMetadata{common.NewBlockBlockMetadata(gb.Height, gb.Hash(), 0, nil, nil, len(gb.Payload))}
	ledger.availablePrevBlocks[string(gb.Hash())] = struct{}{}

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

	log.Printf("Appended:\t\t%x\n", block.Hash()[:15])
}

// GetMinedBlock returns true with a  list of microblocks if all the microblocks for a specific height are available otherwise returns false
func (l *Ledger) GetMacroBlock(height int) ([]common.BlockMetadata, bool) {

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
		return []common.BlockMetadata{}, false
	}

	blocks, count, _ := GetFullestMacroblock(l.concurrencyLevel, heightBlocks)

	return blocks, count == l.concurrencyLevel
}

func (l *Ledger) GetSiblings(height int) ([][]byte, []byte) {

	heightBlocks := l.blockMap[height]

	// returns the genesis block
	if height == 0 {
		if len(heightBlocks) != 1 {
			panic("no genesis block")
		}

		return nil, nil
	}

	blocks, _, previousBlockHash := GetFullestMacroblock(l.concurrencyLevel, heightBlocks)
	//log.Printf("(GetSiblings)Prev block hash updatated [ %x ] \n", previousBlockHash)

	siblingHashes := make([][]byte, l.concurrencyLevel)
	for i := 0; i < len(blocks); i++ {
		block := blocks[i]
		if block.PayloadSize > 0 {
			siblingHashes[i] = block.Hash()
		}
	}

	return siblingHashes, previousBlockHash
}

func (l *Ledger) append(block common.Block) bool {

	previousRoundBlocks, ok := l.blockMap[block.Height-1]

	if !ok {
		// retuning because all of the previous blocks are missing
		return false
	}

	//TODO: I can improve this part by constructing different searchtree for different prev blocks....
	prevBlockStr := string(block.PrevBlockHash)
	_, ok = l.availablePrevBlocks[prevBlockStr]
	if !ok {

		start := time.Now()
		searchTree := common.NewSearchTree(previousRoundBlocks)
		eSearchTreeCreation := time.Since(start)

		_, prevblockAvailable := searchTree.IsHashAvailable(previousRoundBlocks, block.PrevBlockHash)
		eFindingPreviousHash := time.Since(start)

		log.Printf("Search Tree created	:	%s\n", eSearchTreeCreation)
		log.Printf("Prev Hash Calculated:	%s\n", eFindingPreviousHash)

		if !prevblockAvailable {
			// returning because one of the prev blocks is missing!!!
			log.Printf("-----> Prev block is missing: %x\n", block.PrevBlockHash[:10])
			return false
		}

		// prevblock hash is available so put to map...
		l.availablePrevBlocks[prevBlockStr] = struct{}{}
	}

	// apending block top the ledger
	currentRoundBlocks := l.blockMap[block.Height]
	if !areAllSiblingsAvailable(block, currentRoundBlocks) {
		log.Println("-----> A sibling is missing")
		return false
	}

	// emulatest the validation cost for a block
	l.emulateCost(len(block.Payload))

	microblockIndex := MicroBlockIndex(block.Nonce, block.Siblings, l.concurrencyLevel)
	blockMetadata := common.NewBlockBlockMetadata(block.Height, block.Hash(), microblockIndex, block.Siblings, block.PrevBlockHash, len(block.Payload))
	currentRoundBlocks = append(currentRoundBlocks, blockMetadata)
	l.blockMap[block.Height] = currentRoundBlocks

	// the block is validated, and appended to the ledger.
	// the node should disseminate it
	l.readyToDisseminate <- block

	return true
}

func areAllSiblingsAvailable(block common.Block, currentRoundBlocks []common.BlockMetadata) bool {

	siblings := block.Siblings

	currentRoundBlockMap := make(map[string]common.BlockMetadata)
	for i := 0; i < len(currentRoundBlocks); i++ {
		currentRoundBlockMap[string(currentRoundBlocks[i].Hash())] = currentRoundBlocks[i]
	}

	for i := 0; i < len(siblings); i++ {
		if len(siblings[i]) > 0 {
			siblingBlock, ok := currentRoundBlockMap[string(siblings[i])]
			if !ok {
				return false
			}

			if block.Height != siblingBlock.Height {
				panic("siblings height is not correct")
			}

		}
	}

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

func (l *Ledger) emulateCost(payloadSize int) {
	//
	// 0.13 ms unit cost to validate a transaction
	// to emulate the cost of merkle tree creation, we have add 0.003 to the unit cost.
	//
	sleepTime := (float64(0.133) * float64(payloadSize/512))
	sleepDuration := time.Duration(sleepTime) * time.Millisecond
	log.Printf("the node will sleep to emulate tx validation, and merkle tree construction %s \n", sleepDuration)
	time.Sleep(sleepDuration)
}
