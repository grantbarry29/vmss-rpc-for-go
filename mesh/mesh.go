package mesh

import (
	"fmt"
	"net"
	"sync"
	"time"
	"vmss-rpc-for-go/utility"
)

const (
	PORT    = "22333"
	ACKPORT = "22334"
	TYPE    = "tcp"
)

type Instance struct {
	mutex           sync.Mutex
	Subnet          string
	HostIp          string
	RegisteredPeers map[string]struct{}
}

func NewInstance() *Instance {
	return &Instance{
		mutex:           sync.Mutex{},
		Subnet:          "",
		HostIp:          "",
		RegisteredPeers: make(map[string]struct{}),
	}
}

func (ins *Instance) Init() {
	err := ins.getHostIP()
	if err != nil {
		fmt.Print(err)
	}

	err = ins.getSubnet()
	if err != nil {
		fmt.Print(err)
	}

	err = ins.DiscoverLivePeers()
	if err != nil {
		fmt.Print(err)
	}

	// Listen for registration messages
	go utility.ListenTCP(ins.HostIp, PORT, ins.handlePeerRegistration)

	// Listen for acknowledgement messages
	go utility.ListenTCP(ins.HostIp, ACKPORT, ins.handleAcknowledgement)
}

func (ins *Instance) LoopAndShowPeers() {
	ticker := time.NewTicker(time.Second * 1)

	for _ = range ticker.C {
		ins.mutex.Lock()
		fmt.Print("Current registered peers: ")
		for peer := range ins.RegisteredPeers {
			fmt.Print(peer, " ")
		}
		fmt.Print("\n")
		ins.mutex.Unlock()
	}
}

func (ins *Instance) handleAcknowledgement(conn net.Conn) {
	peerIp, err := utility.ReadFromConnection(conn)
	if err != nil {
		return
	}

	ins.registerPeer(peerIp)
}

func (ins *Instance) handlePeerRegistration(conn net.Conn) {

	peerIp, err := utility.ReadFromConnection(conn)
	if err != nil {
		return
	}

	ins.registerPeer(peerIp)

	utility.SendTCP(peerIp, ACKPORT, ins.HostIp, ins.HostIp, ins.unregisterPeer)
}

func (ins *Instance) DiscoverLivePeers() error {
	// create list of subnet hosts
	subnetHosts, err := utility.GetAllHosts(ins.Subnet)
	if err != nil {
		return err
	}

	// send ping to discover live peers. using go routines to ping all peers concurrently
	ch := make(chan string, len(subnetHosts))
	for _, ip := range subnetHosts {
		go func(peerip string) {
			err := utility.SendPing(peerip)
			if err == nil {
				ch <- peerip
			}
			ch <- ""
		}(ip)
	}

	// collect all the active peer ip's from the channel
	var peerip string = ""
	potentialPeers := make(map[string]struct{})
	for j := 0; j < len(subnetHosts); j++ {
		peerip = <-ch
		if peerip != "" && peerip != ins.HostIp {
			potentialPeers[peerip] = struct{}{}
		}
	}

	// attempt register with all potential peers
	for peer := range potentialPeers {
		go utility.SendTCP(peer, PORT, ins.HostIp, ins.HostIp, ins.unregisterPeer)
	}

	ins.mutex.Lock()
	fmt.Print("Number of live peers: ", len(ins.RegisteredPeers), "\n")
	ins.mutex.Unlock()

	return nil
}

func (ins *Instance) getSubnet() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, a := range addrs {
			switch v := a.(type) {

			case *net.IPNet:
				addr := utility.TrimSubnet(a.String())
				if addr == ins.HostIp {
					ins.Subnet = v.String()
					return nil
				}
			}
		}
	}

	return nil
}

func (ins *Instance) getHostIP() error {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ins.HostIp = localAddr.IP.String()
	fmt.Print("My host ip is: ", ins.HostIp, "\n")

	return nil
}

func (ins *Instance) registerPeer(peer string) {
	ins.mutex.Lock()
	ins.RegisteredPeers[peer] = struct{}{}
	fmt.Print("registered new peer: ", peer, "\n")
	ins.mutex.Unlock()
}

func (ins *Instance) unregisterPeer(peer string) {
	ins.mutex.Lock()
	delete(ins.RegisteredPeers, peer)
	fmt.Print("unregistered peer: ", peer, "\n")
	ins.mutex.Unlock()
}
