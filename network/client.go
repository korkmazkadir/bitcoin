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
	rpcClient  *rpc.Client
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

	return client, nil
}

func (c *P2PClient) SendBlock(block *common.Block) {

	go c.rpcClient.Call("P2PServer.HandleBlock", block, nil)
}

func (c *P2PClient) SendSubleaderRequest(req common.SubleaderRequest) {

	go c.rpcClient.Call("P2PServer.HandleSubleaderRequest", req, nil)
}
