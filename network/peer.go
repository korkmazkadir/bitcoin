package network

import (
	"errors"

	"github.com/korkmazkadir/bitcoin/common"
)

var ErrorNoCorrectPeerAvailable = errors.New("there are no correct peers available")

type PeerSet struct {
	peers []*P2PClient
}

func (p *PeerSet) AddPeer(IPAddress string, portNumber int) error {

	client, err := NewClient(IPAddress, portNumber)
	if err != nil {
		return err
	}

	p.peers = append(p.peers, client)

	return nil
}

func (p *PeerSet) DissaminateBlock(block *common.Block) {

	for i := 0; i < len(p.peers); i++ {
		peer := p.peers[i]
		peer.SendBlock(block)
	}

	if len(p.peers) == 0 {
		panic(ErrorNoCorrectPeerAvailable)
	}
}

func (p *PeerSet) SendSubleaderRequest(peerIndex int, subleaderReq common.SubleaderRequest) {

	peer := p.peers[peerIndex]
	peer.SendSubleaderRequest(subleaderReq)
}
