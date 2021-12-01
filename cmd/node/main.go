package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/korkmazkadir/bitcoin/common"
	"github.com/korkmazkadir/bitcoin/consensus"
	"github.com/korkmazkadir/bitcoin/network"
	"github.com/korkmazkadir/bitcoin/registery"
)

func main() {

	hostname := getEnvWithDefault("NODE_HOSTNAME", "127.0.0.1")
	registryAddress := getEnvWithDefault("REGISTRY_ADDRESS", "localhost:1234")

	demux := common.NewDemultiplexer(0)
	server := network.NewServer(demux)

	err := rpc.Register(server)
	if err != nil {
		panic(err)
	}

	rpc.HandleHTTP()
	l, e := net.Listen("tcp", fmt.Sprintf("%s:", hostname))
	if e != nil {
		log.Fatal("listen error:", e)
	}

	// start serving
	go func() {
		for {
			conn, _ := l.Accept()
			go func() {
				rpc.ServeConn(conn)
			}()
		}
	}()

	log.Printf("p2p server started on %s\n", l.Addr().String())
	nodeInfo := getNodeInfo(l.Addr().String())

	registry := registery.NewRegistryClient(registryAddress, nodeInfo)

	nodeInfo.ID = registry.RegisterNode()
	log.Printf("node registeration successful, assigned ID is %d\n", nodeInfo.ID)

	nodeConfig := registry.GetConfig()

	var nodeList []registery.NodeInfo

	for {
		nodeList = registry.GetNodeList()
		nodeCount := len(nodeList)
		if nodeCount == nodeConfig.NodeCount {
			break
		}
		time.Sleep(2 * time.Second)
		log.Printf("received node list %d/%d\n", nodeCount, nodeConfig.NodeCount)
	}

	peerSet := createPeerSet(nodeList, nodeConfig.GossipFanout, nodeInfo)
	statLogger := common.NewStatLogger(nodeInfo.ID)
	// pass node id to use as the seed of pseudo randomnumber generator
	bitcoin := consensus.NewBitcoin(demux, nodeConfig, peerSet, statLogger, nodeInfo.ID)

	runConsensus(bitcoin, nodeConfig.EndRound, nodeConfig.NodeCount, nodeConfig.LeaderCount, nodeConfig.BlockSize)

	// collects stats abd uploads to registry
	log.Printf("uploading stats to the registry\n")
	events := statLogger.GetEvents()
	statList := common.StatList{IPAddress: nodeInfo.IPAddress, PortNumber: nodeInfo.PortNumber, NodeID: nodeInfo.ID, Events: events}
	registry.UploadStats(statList)

	log.Printf("reached target round count. Shutting down in 5 minute\n")
	time.Sleep(5 * time.Minute)

	bitcoin.PrintLedgerStatus()

	log.Printf("exiting as expected...\n")
}

func runConsensus(bitcoinPP *consensus.Bitcoin, numberOfRounds int, nodeCount int, leaderCount int, blockSize int) {

	time.Sleep(5 * time.Second)
	log.Println("Consensus started")

	// previous block is set to genesis block
	previousBlock, _ := bitcoinPP.GetMacroBlock(0)

	currentRound := 1
	for currentRound <= numberOfRounds {

		log.Printf("+++++++++ Round %d +++++++++++++++\n", currentRound)

		block := createBlock(currentRound, hashMacroblock(previousBlock), blockSize, leaderCount)
		minedBlock := bitcoinPP.MineBlock(block)

		payloadSize := 0
		for i := range minedBlock {
			payloadSize += len(minedBlock[i].Payload)
		}

		log.Printf("Appended payload size is %d bytes\n", payloadSize)

		previousBlock = minedBlock
		//log.Printf("decided block hash %x\n", encodeBase64(block.Hash()[:15]))

		roundValue, macroBlockHash := sanityCheck(leaderCount, currentRound, minedBlock)

		//log.Printf("Appended block: %x\n", encodeBase64(singleBlockHash(minedBlock)[:15]))
		log.Printf("Appended Block\t%d\t%x\n", roundValue, macroBlockHash[:15])

		currentRound++
	}

}

func hashMacroblock(blocks []common.Block) [][]byte {
	var hashes [][]byte

	for _, b := range blocks {
		hashes = append(hashes, b.Hash())
	}

	return hashes
}
