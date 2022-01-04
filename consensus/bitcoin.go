package consensus

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"log"
	"time"

	"github.com/korkmazkadir/bitcoin/common"
	"github.com/korkmazkadir/bitcoin/network"
	"github.com/korkmazkadir/bitcoin/registery"
)

type Bitcoin struct {
	demux      *common.Demux
	config     registery.NodeConfig
	peerSet    network.PeerSet
	statLogger *common.StatLogger
	ledger     *Ledger
	publickKey []byte
	privateKey []byte
	nbinom     *NBinom
}

func NewBitcoin(demux *common.Demux, nodeConfig registery.NodeConfig, peerSet network.PeerSet, statLogger *common.StatLogger, nodeID int) *Bitcoin {

	// probability is calculated for the simulator
	// MiningTime is in seconds, I have converted it into ms by multiplying 1000
	prob := float64(nodeConfig.LeaderCount) / (nodeConfig.MiningTime*float64(1000*nodeConfig.NodeCount) + float64(nodeConfig.LeaderCount))

	consensus := &Bitcoin{
		demux:      demux,
		config:     nodeConfig,
		peerSet:    peerSet,
		statLogger: statLogger,
		ledger:     NewLedger(nodeConfig.LeaderCount),
		//nbinom:     NewNBinom(fmt.Sprintf("%d", nodeID), 1, prob),
		//TODO: revert this back
		nbinom: NewNBinom(fmt.Sprintf("%d", time.Now().UnixMilli()), 1, prob),
	}

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	consensus.publickKey = pubKey
	consensus.privateKey = privKey

	// starts a task to disseminate blocks in the background
	go consensus.disseminate()

	return consensus
}

func (b *Bitcoin) GetMacroBlock(round int) ([]common.BlockMetadata, bool) {

	return b.ledger.GetMacroBlock(round)
}

// MineBlock implements simulated mining
func (b *Bitcoin) MineBlock(block common.Block) []common.BlockMetadata {

	// sets block issuer
	block.Issuer = b.publickKey

	b.statLogger.NewRound(block.Height)

	blockChan := b.demux.GetBlockChan()

	miningTimeChan := b.miningTime()

	for {
		select {

		case blockToAppend := <-blockChan:

			microBlockIndex := MicroBlockIndex(blockToAppend.Nonce, blockToAppend.Siblings, b.ledger.concurrencyLevel)

			disseminationTime := int(time.Now().UnixMilli() - blockToAppend.Timestamp)
			b.statLogger.LogBlockReceived(blockToAppend.Height, disseminationTime, blockToAppend.HopCount)
			log.Printf("[%d] Received:\t%x\tHeight: %d \tDissTime: %d ms.\tHopCount: %d\tPrev: %x\n",
				microBlockIndex,
				blockToAppend.Hash()[:15],
				blockToAppend.Height,
				disseminationTime,
				blockToAppend.HopCount,
				blockToAppend.PrevBlockHash[:15],
			)

			// appends the received block to the ledger
			b.ledger.AppendBlock(blockToAppend)

			// gets the macroblock
			blocks, roundFinished := b.ledger.GetMacroBlock(block.Height)
			if roundFinished {
				macroblockHash := common.MacroblockHash(blocks)
				b.statLogger.LogEndOfRound(macroblockHash)
				return blocks
			}

			//b.updateSiblingsAndPrevBlock(&block)
			//TODO: this if statement could be the bug that increases the latency of the protocol!!!
			//if blockToAppend.Height == block.Height {
			b.updateSiblingsAndPrevBlock(&block)
			//}

		case <-miningTimeChan:

			block.Nonce = produceRandomNonce()
			microBlockIndex := MicroBlockIndex(block.Nonce, block.Siblings, b.ledger.concurrencyLevel)

			//strange way of calling a function
			blockPointer := &block
			blockPointer.SetEnqueueTime()

			// I have removed the microblock index check
			// I think it was impossible to have full microbloc index
			// because of siblings mechanism

			block.Signature = Sign(block.Hash(), b.privateKey)
			block.Timestamp = time.Now().UnixMilli()
			b.ledger.AppendBlock(block)

			log.Printf("[%d] Mined:\t\t%x\tHeight: %d\n", microBlockIndex, block.Hash()[:15], block.Height)

			// gets the macroblock
			blocks, roundFinished := b.ledger.GetMacroBlock(block.Height)
			if roundFinished {
				macroblockHash := common.MacroblockHash(blocks)
				b.statLogger.LogEndOfRound(macroblockHash)
				return blocks
			}

			miningTimeChan = b.miningTime()

		}

	}

}

func (b *Bitcoin) updateSiblingsAndPrevBlock(block *common.Block) bool {

	siblingsUpdated := false
	prevHashUpdated := false

	siblings, previousBlockHash := b.ledger.GetSiblings(block.Height)

	if len(siblings) > 0 && !Equal(siblings, block.Siblings) {
		block.Siblings = siblings
		siblingsUpdated = true
	}

	if len(previousBlockHash) > 0 && !bytes.Equal(previousBlockHash, block.PrevBlockHash) {
		log.Printf("(updateSiblingsAndPrevBlock)Prev block hash updatated [ %x ] \n", previousBlockHash)
		block.PrevBlockHash = previousBlockHash
		prevHashUpdated = true
	}

	//TODO: enable sibling logging later...
	// if siblingsUpdated {

	// 	var siblingStrings []string
	// 	for _, sibling := range siblings {
	// 		if len(sibling) > 0 {
	// 			siblingStrings = append(siblingStrings, fmt.Sprintf("[ %x ]", sibling[:10]))
	// 		} else {
	// 			siblingStrings = append(siblingStrings, "[ ]")
	// 		}

	// 	}

	// 	log.Printf("Siblings Updated %s\n", strings.Join(siblingStrings[:], "--"))
	// }

	if siblingsUpdated {
		siblingCount := 0
		for _, sibling := range siblings {
			if len(sibling) > 0 {
				siblingCount++
			}
		}
		log.Printf("[Sibling Count Updated] %d/%d\n", siblingCount, b.config.LeaderCount)
	}

	//if prevHashUpdated {
	//	log.Printf("Previous Block Hash Updated %x\n", previousBlockHash)
	//}

	return siblingsUpdated || prevHashUpdated
}

// disseminates blocks in the background
func (b *Bitcoin) disseminate() {
	for {
		blockToDisseminate := <-b.ledger.readyToDisseminate
		processingTime := blockToDisseminate.GetEnqueueElapsedTime()
		b.statLogger.LogProcessingTime(processingTime)
		// increments the hope count
		blockToDisseminate.HopCount++
		log.Printf("Forwarding:\t\t%x\n", blockToDisseminate.Hash()[:15])
		b.peerSet.DissaminateBlock(blockToDisseminate)
	}
}

/*
func (b *Bitcoin) miningTime() <-chan time.Time {

	// MiningTime / CC
	expected := b.config.MiningTime / b.ledger.concurrencyLevel
	simulatedMiningTime := int(-math.Log(1.0-rand.Float64()) * float64(expected) * 1 / (1 / float64(b.config.NodeCount)))
	log.Printf("[expected: %d]Mining time is %d \n", expected, simulatedMiningTime)
	return time.After(time.Duration(simulatedMiningTime) * time.Second)
}*/

func (b *Bitcoin) miningTime() <-chan time.Time {

	simulatedMiningTime := b.nbinom.Random()
	log.Printf("[expected: %f ms]Mining time is %d ms\n", b.config.MiningTime*1000, simulatedMiningTime)
	return time.After(time.Duration(simulatedMiningTime) * time.Millisecond)
}

func (b *Bitcoin) PrintLedgerStatus() {
	b.ledger.PrintStatus()
}
