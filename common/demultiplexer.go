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

	subLeaderReqChan chan SubleaderRequest
}

// NewDemultiplexer creates a new demultiplexer with initial round value
func NewDemultiplexer(initialRound int) *Demux {

	demux := &Demux{currentRound: initialRound}
	demux.processedMessageMap = make(map[string]struct{})
	demux.blockChan = make(chan Block, channelCapacity)
	demux.subLeaderReqChan = make(chan SubleaderRequest, channelCapacity)

	return demux
}

// EnqueBlockChunk enques a block chunk to be the consumed by consensus layer
func (d *Demux) EnqueBlock(block Block) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	blockHash := string(block.Hash())
	if d.isProcessed(blockHash) {
		// chunk is already processed
		return
	}

	d.blockChan <- block

	d.markAsProcessed(blockHash)
}

// EnqueBlockChunk enques a block chunk to be the consumed by consensus layer
func (d *Demux) EnqueSubleaderRequest(subleaderReq SubleaderRequest) {

	d.subLeaderReqChan <- subleaderReq
}

// EnqueBlockChunk enques a block chunk to be the consumed by consensus layer
func (d *Demux) GetBlockChan() chan Block {

	return d.blockChan
}

func (d *Demux) GetSubleaderRequestChan() chan SubleaderRequest {

	return d.subLeaderReqChan
}

func (d *Demux) isProcessed(hash string) bool {

	chunkHashString := string(hash)
	_, ok := d.processedMessageMap[chunkHashString]
	return ok
}

func (d *Demux) MarkAsProcessed(hash string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.markAsProcessed(hash)
}

func (d *Demux) markAsProcessed(hash string) {

	d.processedMessageMap[hash] = struct{}{}
}
