package consensus

import (
	"crypto/ed25519"
	"log"
	"math"
	"math/rand"
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

	simulatedMiningTime := b.miningTime()
	blockChan := b.demux.GetBlockChan()

	log.Printf("Mining time is %d \n", simulatedMiningTime)

	for {
		select {

		case blockToAppend := <-blockChan:

			microBlockIndex := int(blockToAppend.Nonce % int64(b.ledger.concurrencyLevel))

			log.Printf("[%d] Received:\t%x\tHeight: %d\n", microBlockIndex, blockToAppend.Hash(), blockToAppend.Height)

			// appends the received block to the ledger
			b.ledger.AppendBlock(blockToAppend)

			// gets the macroblock
			blocks, roundFinished := b.ledger.GetMacroBlock(block.Height)
			if roundFinished {
				b.statLogger.LogEndOfRound()
				return blocks
			}

		case <-time.After(time.Duration(simulatedMiningTime) * time.Second):

			block.Nonce = produceRandomNonce()
			microBlockIndex := b.getBlockIndex(block.Nonce)
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
			// if its is here, it means that there are missing microblocks. The current node should try to mine
			simulatedMiningTime = b.miningTime()
			log.Printf("Mining time is %d \n", simulatedMiningTime)

		}

	}

}

func (b *Bitcoin) getBlockIndex(nonce int64) int {

	return int(nonce % int64(b.ledger.concurrencyLevel))
}

// disseminates blocks in the background
func (b *Bitcoin) disseminate() {
	for {
		blockToDisseminate := <-b.ledger.readyToDisseminate
		log.Printf("Forwarding:\t\t%x\n", blockToDisseminate.Hash())
		b.peerSet.DissaminateBlock(blockToDisseminate)
	}
}

func (b *Bitcoin) miningTime() int {

	return int(-math.Log(1.0-rand.Float64()) * 100 * 1 / (1 / float64(b.config.NodeCount)))
}

func (b *Bitcoin) PrintLedgerStatus() {
	b.ledger.PrintStatus()
}
