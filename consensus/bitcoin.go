package consensus

import (
	"fmt"
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
}

func NewBitcoin(demux *common.Demux, nodeConfig registery.NodeConfig, peerSet network.PeerSet, statLogger *common.StatLogger) *Bitcoin {

	consensus := &Bitcoin{
		demux:      demux,
		config:     nodeConfig,
		peerSet:    peerSet,
		statLogger: statLogger,
		ledger:     NewLedger(nodeConfig.LeaderCount),
	}

	return consensus
}

// MineBlock implements simulated mining
func (b *Bitcoin) MineBlock(block common.Block) {

	simulatedMiningTime := 123
	blockChan := b.demux.GetBlockChan()

	for {
		select {

		case blockToDisseminate := <-b.ledger.readyToDisseminate:
			b.peerSet.DissaminateBlock(blockToDisseminate)

		case blockToAppend := <-blockChan:
			b.ledger.AppendBlock(blockToAppend)
			result := b.ledger.IsBlockMined(block.Height)
			if result {
				fmt.Println("end of round")
				return
			}

		case <-time.After(time.Duration(simulatedMiningTime) * time.Second):
			//
			// TODO: Handle the case where a microblock mined but there is already an appended microblock
			// A correct node does not submit a macroblock if there is already one mined!!!
			// Current node should try to mine a new block.
			// Nonce field must be random!!!
			//
		}

	}

}
