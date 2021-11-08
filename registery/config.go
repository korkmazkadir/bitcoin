package registery

import (
	"crypto/sha256"
	"fmt"
)

type NodeConfig struct {
	NodeCount int

	EndRound int

	GossipFanout int

	LeaderCount int

	BlockSize int

	MiningTime int
}

func (nc NodeConfig) Hash() []byte {

	str := fmt.Sprintf("%v", nc)

	h := sha256.New()
	_, err := h.Write([]byte(str))
	if err != nil {
		panic(err)
	}

	return h.Sum(nil)
}

func (nc *NodeConfig) CopyFields(cp NodeConfig) {
	nc.NodeCount = cp.NodeCount
	nc.EndRound = cp.EndRound
	nc.GossipFanout = cp.GossipFanout
	nc.LeaderCount = cp.LeaderCount
	nc.BlockSize = cp.BlockSize
	nc.MiningTime = cp.MiningTime
}
