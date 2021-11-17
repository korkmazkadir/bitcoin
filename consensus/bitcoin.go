package consensus

import (
	"crypto/ed25519"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
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
}

func NewBitcoin(demux *common.Demux, nodeConfig registery.NodeConfig, peerSet network.PeerSet, statLogger *common.StatLogger) *Bitcoin {

	consensus := &Bitcoin{
		demux:      demux,
		config:     nodeConfig,
		peerSet:    peerSet,
		statLogger: statLogger,
		ledger:     NewLedger(nodeConfig.LeaderCount),
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

func (b *Bitcoin) GetMacroBlock(round int) ([]common.Block, bool) {

	return b.ledger.GetMacroBlock(round)
}

// MineBlock implements simulated mining
func (b *Bitcoin) MineBlock(block common.Block) []common.Block {

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
			log.Printf("[%d] Received:\t%x\tHeight: %d \tDissTime: %d ms.\tHopCount: %d\n",
				microBlockIndex,
				blockToAppend.Hash()[:15],
				blockToAppend.Height,
				disseminationTime,
				blockToAppend.HopCount)

			// appends the received block to the ledger
			b.ledger.AppendBlock(blockToAppend)

			// gets the macroblock
			blocks, roundFinished := b.ledger.GetMacroBlock(block.Height)
			if roundFinished {
				macroblockHash := common.MacroblockHash(blocks)
				b.statLogger.LogEndOfRound(macroblockHash)
				return blocks
			}

			b.updateSiblingsAndPrevBlock(&block)
			//if b.updateSiblingsAndPrevBlock(&block) {
			//	miningTimeChan = b.miningTime()
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

	siblings, previousBlockHashes := b.ledger.GetSiblings(block.Height)

	if !Equal(siblings, block.Siblings) {
		block.Siblings = siblings
		siblingsUpdated = true
	}

	if !Equal(previousBlockHashes, block.PrevBlockHashes) {
		block.PrevBlockHashes = previousBlockHashes
		prevHashUpdated = true
	}

	if siblingsUpdated {

		var siblingStrings []string
		for _, sibling := range siblings {
			if len(sibling) > 0 {
				siblingStrings = append(siblingStrings, fmt.Sprintf("[ %x ]", sibling[:10]))
			} else {
				siblingStrings = append(siblingStrings, "[ ]")
			}

		}

		log.Printf("Siblings Updated %s\n", strings.Join(siblingStrings[:], "--"))
	}

	if prevHashUpdated {

		var prevHashString []string
		for _, prevHash := range previousBlockHashes {
			if len(prevHash) > 0 {
				prevHashString = append(prevHashString, fmt.Sprintf("[ %x ]", prevHash[:10]))
			} else {
				prevHashString = append(prevHashString, "[ ]")
			}

		}

		log.Printf("Previous Block Hash Updated %s\n", strings.Join(prevHashString[:], "--"))
	}

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

func (b *Bitcoin) miningTime() <-chan time.Time {

	// MiningTime / CC
	expected := b.config.MiningTime / b.ledger.concurrencyLevel
	simulatedMiningTime := int(-math.Log(1.0-rand.Float64()) * float64(expected) * 1 / (1 / float64(b.config.NodeCount)))
	log.Printf("[expected: %d]Mining time is %d \n", expected, simulatedMiningTime)
	return time.After(time.Duration(simulatedMiningTime) * time.Second)
}

func (b *Bitcoin) PrintLedgerStatus() {
	b.ledger.PrintStatus()
}
