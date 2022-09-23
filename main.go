package main

import (
	"fmt"
	"math/rand"
	"time"
	"vmss-rpc-for-go/mesh"
	"vmss-rpc-for-go/rpc"
)

func main() {
	host := mesh.NewInstance()
	host.Init()

	//host.LoopAndShowPeers()

	// Start RPC Server
	rpcServer := rpc.NewRPCServer()
	rpcServer.ListenAndRegister()

	// Start an RPC client
	rpcClient := rpc.NewRPCClient()

	// Generate server ID
	serverID := rand.Intn(1000)

	// Wait for a couple of peers to populate
	for len(host.Peers()) < 2 {
		time.Sleep(time.Millisecond * 100)
	}

	// Send RPC to all peers, refresh peers if connection fails
	for _, peer := range host.Peers() {
		fmt.Print("peer: ", peer, "\n")
		err := rpcClient.CallGetSendServerName(peer, fmt.Sprint(serverID))
		if err != nil {
			host.DiscoverLivePeers()
			time.Sleep(time.Millisecond * 100)
			err = rpcClient.CallGetSendServerName(peer, fmt.Sprint(serverID))
		}
	}

	// Wait and receive all messages
	time.Sleep(time.Second * 10)

}
