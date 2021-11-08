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
	demux            *common.Demux
	config           registery.NodeConfig
	peerSet          network.PeerSet
	subleaderPeerSet network.PeerSet
	statLogger       *common.StatLogger
	ledger           *MacroblockLedger
	publickKey       []byte
	privateKey       []byte
}

func NewBitcoin(demux *common.Demux, nodeConfig registery.NodeConfig, peerSet network.PeerSet, subleaderPeerSet network.PeerSet, statLogger *common.StatLogger) *Bitcoin {

	consensus := &Bitcoin{
		demux:            demux,
		config:           nodeConfig,
		peerSet:          peerSet,
		subleaderPeerSet: subleaderPeerSet,
		statLogger:       statLogger,
		ledger:           NewMacroblockLedger(nodeConfig.LeaderCount),
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

func (b *Bitcoin) GetMacroBlock(round int) ([]common.Block, []byte, bool) {

	return b.ledger.GetMacroblock(round)
}

// MineBlock implements simulated mining
func (b *Bitcoin) MineBlock(block common.Block) ([]common.Block, []byte) {

	// sets block issuer
	block.Issuer = b.publickKey

	b.statLogger.NewRound(block.Height)

	miningTimeChan := b.miningTime()
	blockChan := b.demux.GetBlockChan()
	subleaderReqChan := b.demux.GetSubleaderRequestChan()

	for {
		select {

		case blockToAppend := <-blockChan:

			microBlockIndex := int(blockToAppend.Nonce % int64(b.ledger.concurrencyLevel))

			log.Printf("[%d] Received:\t%x\tHeight: %d\n", microBlockIndex, blockToAppend.Hash(), blockToAppend.Height)

			// appends the received block to the ledger
			b.ledger.AppendBlock(blockToAppend)

		case subleaderReq := <-subleaderReqChan:

			if subleaderReq.Height < b.ledger.GetHeight() {
				log.Printf("Discarts subleadership request for the round %d\n", subleaderReq.Height)
				break
			}

			block.Nonce = produceRandomNonce()
			block.MicroblockIndex = subleaderReq.MicroblockIndex
			block.PuzzleSolver = subleaderReq.PuzzleSolver

			log.Println("Submitting a sub block")
			b.ledger.AppendBlock(block)

		case <-miningTimeChan:

			block.Nonce = produceRandomNonce()
			block.MicroblockIndex = 0
			block.PuzzleSolver = b.publickKey

			// sends subleaderhip requests
			for i := 1; i < b.config.LeaderCount; i++ {
				subleaderReq := common.SubleaderRequest{
					Height:          block.Height,
					PuzzleSolver:    b.publickKey,
					MicroblockIndex: i,
				}
				b.subleaderPeerSet.SendSubleaderRequest(i, subleaderReq)
			}

			b.ledger.AppendBlock(block)

		}

		// checks for the end of round...
		blocks, macroblockHash, roundFinished := b.ledger.GetMacroblock(block.Height)
		if roundFinished {
			b.statLogger.LogEndOfRound()
			return blocks, macroblockHash
		}

	}

}

// disseminates blocks in the background
func (b *Bitcoin) disseminate() {
	for {
		blockToDisseminate := <-b.ledger.readyToDisseminate
		log.Printf("Forwarding:\t\t%x\n", blockToDisseminate.Hash())
		b.peerSet.DissaminateBlock(&blockToDisseminate)
	}
}

func (b *Bitcoin) miningTime() <-chan time.Time {

	expected := b.config.MiningTime
	simulatedMiningTime := int(-math.Log(1.0-rand.Float64()) * float64(expected) * 1 / (1 / float64(b.config.NodeCount)))
	log.Printf("[expected: %d]Mining time is %d \n", expected, simulatedMiningTime)
	return time.After(time.Duration(simulatedMiningTime) * time.Second)
}

func (b *Bitcoin) PrintLedgerStatus() {
	b.ledger.PrintStatus()
}
