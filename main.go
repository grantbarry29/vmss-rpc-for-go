package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
)

const (
	PORT = "22333"
	TYPE = "tcp"
)

type Instance struct {
	mutex           sync.Mutex
	Subnet          string
	HostIp          string
	RegisteredPeers map[string]bool
}

func newInstance() *Instance {
	return &Instance{
		mutex:           sync.Mutex{},
		Subnet:          "",
		HostIp:          "",
		RegisteredPeers: make(map[string]bool),
	}
}

func main() {
	host := newInstance()

	err := host.getHostIP()
	if err != nil {
		fmt.Print(err)
	}

	err = host.getSubnet()
	if err != nil {
		fmt.Print(err)
	}

	err = host.discoverLivePeers()
	if err != nil {
		fmt.Print(err)
	}

	host.registerWithPeers()

	host.listenForPeers()
}

func (ins *Instance) registerWithPeers() {
	for peer := range ins.RegisteredPeers {
		go func(ip string) {
			conn, err := net.Dial(TYPE, ip+":"+PORT)
			if err != nil {
				fmt.Print("Dial error: ", err, "\n")
				return
			}
			defer conn.Close()

			fmt.Print("sending registration to ", ip, " \n")
			_, err = conn.Write([]byte(ins.HostIp))
			if err != nil {
				fmt.Print(err)
			}
		}(peer)
	}

}

func (ins *Instance) listenForPeers() {
	listen, err := net.Listen(TYPE, ins.HostIp+":"+PORT)
	if err != nil {
		fmt.Print("listen error: ", err, "\n")
		os.Exit(1)
	}
	defer listen.Close()

	for {
		fmt.Print("Listening on port ", PORT, "...")
		conn, err := listen.Accept()
		if err != nil {
			fmt.Print("listen error: ", err, "\n")
			os.Exit(1)
		}
		go ins.handlePeerRegistration(conn)
	}
}

func (ins *Instance) handlePeerRegistration(conn net.Conn) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		return
	}

	peerIp := bytes.NewBuffer(buffer).String()
	ins.RegisteredPeers[peerIp] = true
	fmt.Print("registered new peer: ", peerIp)
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
				addr := trimSubnet(a.String())
				if addr == ins.HostIp {
					ins.Subnet = v.String()
					fmt.Print("My subnet is: ", ins.Subnet, "\n")
					return nil
				}
			}
		}
	}

	return nil
}

func (ins *Instance) discoverLivePeers() error {
	// create list of subnet hosts
	subnetHosts, err := getAllHosts(ins.Subnet)
	if err != nil {
		return err
	}

	// send ping to discover live peers. using go routines to ping all peers concurrently
	ch := make(chan string, len(subnetHosts))
	for _, ip := range subnetHosts {
		go func(peerip string) {
			err := sendPing(peerip)
			if err == nil {
				ch <- peerip
			}
			ch <- ""
		}(ip)
	}

	// collect all the active peer ip's from the channel
	var peerip string = ""
	ins.RegisteredPeers = make(map[string]bool)
	for j := 0; j < len(subnetHosts); j++ {
		peerip = <-ch
		if peerip != "" && peerip != ins.HostIp {
			ins.RegisteredPeers[peerip] = false
		}
	}

	fmt.Print("Number of live peers: ", len(ins.RegisteredPeers), "\n")

	return nil
}

func sendPing(ip string) error {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	pinger.SetPrivileged(true)
	pinger.Count = 10
	pinger.Size = 56
	pinger.Interval = 100 * time.Millisecond
	pinger.Timeout = 1 * time.Second

	err = pinger.Run()
	if err != nil {
		return err
	}

	if pinger.PacketsRecv == 0 {
		return errors.New("Ping received no packets...")
	}

	return nil
}

func getAllHosts(cidr string) ([]string, error) {

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// remove network address and broadcast address
	return ips[0 : len(ips)-1], nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func trimSubnet(subnet string) string {
	addr := ""
	addrSplit := strings.SplitAfter(subnet, "/")
	if len(addrSplit) > 0 {
		addr = addrSplit[0]
		addr = strings.TrimSuffix(addr, "/")
	}

	return addr
}
