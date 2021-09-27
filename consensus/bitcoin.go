package consensus

import (
	"crypto/ed25519"
	"fmt"
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

	log.Println("mining...")

	// sets block issuer
	block.Issuer = b.publickKey

	b.statLogger.NewRound(block.Height)

	simulatedMiningTime := b.miningTime()
	blockChan := b.demux.GetBlockChan()

	log.Printf("simulated mining time is %d \n", simulatedMiningTime)

	for {
		select {

		case blockToAppend := <-blockChan:

			log.Printf("Block received. hash %x \n", blockToAppend.Hash())

			// appends the received block to the ledger
			b.ledger.AppendBlock(blockToAppend)

			// gets the macroblock
			blocks, roundFinished := b.ledger.GetMacroBlock(block.Height)
			if roundFinished {
				b.statLogger.LogEndOfRound()
				fmt.Println("end of round")
				return blocks
			}

		case <-time.After(time.Duration(simulatedMiningTime) * time.Second):

			block.Nonce = produceRandomNonce()
			microBlockIndex := int(block.Nonce % int64(b.ledger.concurrencyLevel))
			_, blockAvailable := b.ledger.GetMicroblock(block.Height, microBlockIndex)
			// appends the mined block if there is not a block mined for the specific index
			if !blockAvailable {
				// signs the block
				block.Signature = Sign(block.Hash(), b.privateKey)
				b.ledger.AppendBlock(block)
			}

			// gets the macroblock
			blocks, roundFinished := b.ledger.GetMacroBlock(block.Height)
			if roundFinished {
				b.statLogger.LogEndOfRound()
				log.Println("end of round")
				return blocks
			}

			log.Println("unsuccessful mining...")
			// if its is here, it means that there are missing microblocks. The current node should try to mine
			simulatedMiningTime = b.miningTime()
			log.Printf("simulated mining time is %d \n", simulatedMiningTime)

		}

	}

}

// disseminates blocks in the background
func (b *Bitcoin) disseminate() {
	for {
		blockToDisseminate := <-b.ledger.readyToDisseminate
		log.Println("forwarding a block...")
		b.peerSet.DissaminateBlock(blockToDisseminate)
	}
}

func (b *Bitcoin) miningTime() int {

	return int(-math.Log(1.0-rand.Float64()) * 30 * 1 / (1 / float64(b.config.NodeCount)))
}
