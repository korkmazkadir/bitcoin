package network

import (
	"fmt"
	"net/rpc"

	"github.com/korkmazkadir/bitcoin/common"
)

// Client implements P2P client
type P2PClient struct {
	IPAddress  string
	portNumber int

	rpcClient *rpc.Client

	blockChan chan common.Block

	err error
}

// NewClient creates a new client
func NewClient(IPAddress string, portNumber int) (*P2PClient, error) {

	rpcClient, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", IPAddress, portNumber))
	if err != nil {
		return nil, err
	}

	client := &P2PClient{}
	client.IPAddress = IPAddress
	client.portNumber = portNumber
	client.rpcClient = rpcClient

	client.blockChan = make(chan common.Block, 1024)

	return client, nil
}

// Start starts the main loop of client. It blocks the calling goroutine
func (c *P2PClient) Start() {

	c.mainLoop()
}

// SendBlockChunk enques a chunk of a block to send
func (c *P2PClient) SendBlock(block common.Block) {

	c.blockChan <- block
}

func (c *P2PClient) mainLoop() {

	for {
		select {

		case block := <-c.blockChan:
			go c.rpcClient.Call("P2PServer.HandleBlock", block, nil)

		}
	}
}
