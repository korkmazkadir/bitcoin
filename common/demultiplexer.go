package common

import (
	"sync"
)

const (
	channelCapacity = 1024
)

// Demux provides message multiplexing service
// Network and consensus layer communicate using demux
type Demux struct {
	mutex sync.Mutex

	currentRound int

	// it is used to filter already processed messages
	processedMessageMap map[string]struct{}

	blockChan chan Block
}

// NewDemultiplexer creates a new demultiplexer with initial round value
func NewDemultiplexer(initialRound int) *Demux {

	demux := &Demux{currentRound: initialRound}
	demux.processedMessageMap = make(map[string]struct{})
	demux.blockChan = make(chan Block, channelCapacity)

	return demux
}

// EnqueBlockChunk enques a block chunk to be the consumed by consensus layer
func (d *Demux) EnqueBlock(block Block) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	round := block.Height
	blockHash := string(block.Hash())
	if d.isProcessed(blockHash) {
		// chunk is already processed
		return
	}

	d.blockChan <- block

	d.markAsProcessed(round, blockHash)
}

// EnqueBlockChunk enques a block chunk to be the consumed by consensus layer
func (d *Demux) GetBlockChan() chan Block {

	return d.blockChan
}

func (d *Demux) isProcessed(hash string) bool {

	chunkHashString := string(hash)
	_, ok := d.processedMessageMap[chunkHashString]
	return ok
}

func (d *Demux) markAsProcessed(round int, hash string) {

	d.processedMessageMap[hash] = struct{}{}
}
