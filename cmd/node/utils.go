package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/korkmazkadir/bitcoin/common"
	"github.com/korkmazkadir/bitcoin/network"
	"github.com/korkmazkadir/bitcoin/registery"
)

func createPeerSet(nodeList []registery.NodeInfo, fanOut int, nodeInfo registery.NodeInfo) network.PeerSet {

	var copyNodeList []registery.NodeInfo
	copyNodeList = append(copyNodeList, nodeList...)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(copyNodeList), func(i, j int) { copyNodeList[i], copyNodeList[j] = copyNodeList[j], copyNodeList[i] })

	peerSet := network.PeerSet{}

	peerCount := 0
	for i := 0; i < len(copyNodeList); i++ {

		if peerCount == fanOut {
			break
		}

		peer := copyNodeList[i]
		if peer.ID == nodeInfo.ID || peer.IPAddress == nodeInfo.IPAddress {
			continue
		}

		err := peerSet.AddPeer(peer.IPAddress, peer.PortNumber)
		if err != nil {
			panic(err)
		}
		log.Printf("new peer added: %s:%d ID %d\n", peer.IPAddress, peer.PortNumber, peer.ID)
		peerCount++
	}

	return peerSet
}

func getNodeInfo(netAddress string) registery.NodeInfo {
	tokens := strings.Split(netAddress, ":")

	ipAddress := tokens[0]
	portNumber, err := strconv.Atoi(tokens[1])
	if err != nil {
		panic(err)
	}

	return registery.NodeInfo{IPAddress: ipAddress, PortNumber: portNumber}
}

func encodeBase64(hex []byte) string {
	return base64.StdEncoding.EncodeToString([]byte(hex))
}

func getEnvWithDefault(key string, defaultValue string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		val = defaultValue
	}

	log.Printf("%s=%s\n", key, val)
	return val
}

func isElectedAsLeader(nodeList []registery.NodeInfo, round int, nodeID int, leaderCount int) bool {

	// assumes that node list is same for all nodes
	// shuffle the node list using round number as the source of randomness
	rand.Seed(int64(round))
	rand.Shuffle(len(nodeList), func(i, j int) { nodeList[i], nodeList[j] = nodeList[j], nodeList[i] })

	var electedLeaders []int
	for i := 0; i < leaderCount; i++ {
		electedLeaders = append(electedLeaders, nodeList[i].ID)
		if nodeList[i].ID == nodeID {
			log.Println("elected as leader")
			return true
		}
	}

	log.Printf("Elected leaders: %v\n", electedLeaders)

	return false
}

func createBlock(round int, previousBlockHash []byte, blockSize int, leaderCount int) common.Block {

	//payloadSize := int(math.Ceil(float64(blockSize) / float64(leaderCount)))

	block := common.Block{
		Height:        round,
		Payload:       getRandomByteSlice(blockSize),
		PrevBlockHash: previousBlockHash,
		Siblings:      make([][]byte, leaderCount),
	}

	return block
}

func getRandomByteSlice(size int) []byte {
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return data
}

func sanityCheck(cc int, currentRound int, blocks []common.BlockMetadata) (round int, macroBlockHash []byte) {

	if len(blocks) != cc {
		panic(fmt.Errorf("microblock count(%d) is not equal to cc(%d) value", len(blocks), cc))
	}

	microblockBlockHashes := make(map[string]struct{})
	for i := 0; i < len(blocks); i++ {
		microBlock := blocks[i]
		key := string(microBlock.Hash())
		_, ok := microblockBlockHashes[key]
		if ok {
			panic(fmt.Errorf("same microblock appended twice %x", microBlock.Hash()))
		}
		microblockBlockHashes[key] = struct{}{}
	}

	previousBlockHash := blocks[0].PrevBlockHash
	for i := 0; i < len(blocks); i++ {
		microBlock := blocks[i]

		if microBlock.Height != currentRound {
			panic(fmt.Errorf("microblock height(%d) is not equal to current round number(%d)", microBlock.Height, currentRound))
		}

		if !bytes.Equal(previousBlockHash, microBlock.PrevBlockHash) {
			panic(fmt.Errorf("previous block hashes are different [%d] != [%d]", previousBlockHash, microBlock.PrevBlockHash))
		}

		for _, siblingHash := range microBlock.Siblings {
			if len(siblingHash) > 0 {
				_, ok := microblockBlockHashes[string(siblingHash)]
				if !ok {
					log.Println(fmt.Errorf("sibling is not in the list of appended blocks %x", siblingHash))
				}
			}
		}

	}

	round = currentRound
	macroBlockHash = common.MacroblockHash(blocks)
	return
}
