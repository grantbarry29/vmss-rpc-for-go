package main

import (
	"vmss-rpc-for-go/mesh"
)

func main() {
	host := mesh.NewInstance()

	host.Init()

	host.LoopAndShowPeers()
}
