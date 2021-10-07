package consensus

import (
	"bytes"
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

			log.Printf("[%d] Received:\t%x\tHeight: %d\n", microBlockIndex, blockToAppend.Hash(), blockToAppend.Height)

			// appends the received block to the ledger
			b.ledger.AppendBlock(blockToAppend)

			// gets the macroblock
			blocks, roundFinished := b.ledger.GetMacroBlock(block.Height)
			if roundFinished {
				b.statLogger.LogEndOfRound()
				return blocks
			}

			if b.updateSiblings(block) {
				// siblings are updated
				miningTimeChan = b.miningTime()
			}

		case <-miningTimeChan:

			block.Nonce = produceRandomNonce()
			microBlockIndex := MicroBlockIndex(block.Nonce, block.Siblings, b.ledger.concurrencyLevel)

			_, blockAvailable := b.ledger.GetMicroblock(block.Height, microBlockIndex)
			// appends the mined block if there is not a block mined for the specific index
			if !blockAvailable {
				// signs the block
				block.Signature = Sign(block.Hash(), b.privateKey)
				b.ledger.AppendBlock(block)

				log.Printf("[%d] Mined:\t\t%x\tHeight: %d\n", microBlockIndex, block.Hash(), block.Height)
			}

			// gets the macroblock
			blocks, roundFinished := b.ledger.GetMacroBlock(block.Height)
			if roundFinished {
				b.statLogger.LogEndOfRound()
				log.Println("end of round")
				return blocks
			}

			log.Println("Unsuccessful mining...")
			miningTimeChan = b.miningTime()

		}

	}

}

func (b *Bitcoin) updateSiblings(block common.Block) bool {

	siblingsUpdated := false
	siblings := b.ledger.GetSiblings(block.Height)

	var siblingStrings []string

	for i := 0; i < b.ledger.concurrencyLevel; i++ {

		if len(siblings[i]) > 0 {

			if len(block.Siblings[i]) > 0 && bytes.Equal(siblings[i], block.Siblings[i]) == false {
				panic(fmt.Errorf("a sibling is reassigned"))
			}

			if len(block.Siblings[i]) == 0 {
				block.Siblings[i] = siblings[i]
				siblingsUpdated = true
			}

			siblingStrings = append(siblingStrings, fmt.Sprintf("[%x]", block.Siblings[i][:5]))

		} else {

			siblingStrings = append(siblingStrings, "[ ]")
		}

	}

	if siblingsUpdated {
		log.Printf("Siblings Updated %s\n", strings.Join(siblingStrings[:], "--"))
	}

	return siblingsUpdated
}

// disseminates blocks in the background
func (b *Bitcoin) disseminate() {
	for {
		blockToDisseminate := <-b.ledger.readyToDisseminate
		log.Printf("Forwarding:\t\t%x\n", blockToDisseminate.Hash())
		b.peerSet.DissaminateBlock(blockToDisseminate)
	}
}

func (b *Bitcoin) miningTime() <-chan time.Time {

	simulatedMiningTime := int(-math.Log(1.0-rand.Float64()) * 80 * 1 / (1 / float64(b.config.NodeCount)))
	log.Printf("Mining time is %d \n", simulatedMiningTime)
	return time.After(time.Duration(simulatedMiningTime) * time.Second)
}

func (b *Bitcoin) PrintLedgerStatus() {
	b.ledger.PrintStatus()
}
