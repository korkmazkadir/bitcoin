package network

import (
	"github.com/korkmazkadir/bitcoin/common"
)

type P2PServer struct {
	demux *common.Demux
}

func NewServer(demux *common.Demux) *P2PServer {
	server := &P2PServer{demux: demux}
	return server
}

func (s *P2PServer) HandleBlock(block *common.Block, reply *int) error {

	s.demux.EnqueBlock(*block)

	return nil
}

func (s *P2PServer) HandleSubleaderRequest(subleaderReq *common.SubleaderRequest, reply *int) error {

	s.demux.EnqueSubleaderRequest(*subleaderReq)

	return nil
}
