package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/go-ping/ping"
)

type Instance struct {
	Subnet string
	HostIp string
	Peers  []string
}

func newInstance() *Instance {
	return &Instance{}
}

func main() {
	self := Instance{}

	err := self.getHostIP()
	if err != nil {
		fmt.Print(err)
	}

	err = self.getSubnet()
	if err != nil {
		fmt.Print(err)
	}

	err = self.discoverLivePeers()
	if err != nil {
		fmt.Print(err)
	}

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
	active := make([]string, 0)
	for j := 0; j < len(subnetHosts); j++ {
		peerip = <-ch
		if peerip != "" {
			active = append(active, peerip)
		}
	}

	fmt.Print("number of live peers: ", len(active), "\n")

	return nil
}

func sendPing(ip string) error {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		return err
	}

	err = pinger.Run()
	if err != nil {
		return err
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
