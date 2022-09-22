package main

import (
	"vmss-rpc-for-go/mesh"
	"vmss-rpc-for-go/rpc"
)

func main() {
	host := mesh.NewInstance()

	host.Init()

	//host.LoopAndShowPeers()

	rpc.Server()
	rpc.Client()
}
