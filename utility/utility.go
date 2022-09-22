package utility

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/go-ping/ping"
)

func SendPing(ip string) error {
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

func GetAllHosts(cidr string) ([]string, error) {

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); Inc(ip) {
		ips = append(ips, ip.String())
	}

	// remove network address and broadcast address
	return ips[0 : len(ips)-1], nil
}

func Inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func TrimSubnet(subnet string) string {
	addr := ""
	addrSplit := strings.SplitAfter(subnet, "/")
	if len(addrSplit) > 0 {
		addr = addrSplit[0]
		addr = strings.TrimSuffix(addr, "/")
	}

	return addr
}

const (
	TYPE = "tcp"
)

func SendTCP(ip string, port string, hostIp string, message string, onFailCallback func(string)) {
	// Open tcp connection with peer
	conn, err := net.Dial(TYPE, ip+":"+port)

	// Remove peer if unreachable
	if err != nil {
		onFailCallback(ip)
		return
	}
	defer conn.Close()

	// Send registration to peer
	fmt.Print("sending registration to ", ip, " \n")
	_, err = conn.Write([]byte(hostIp))
	if err != nil {
		fmt.Print(err)
	}
}

func ListenTCP(ip string, port string, onConnect func(net.Conn)) {
	listen, err := net.Listen(TYPE, ip+":"+port)
	fmt.Print("Listening on port ", port, "...\n")
	if err != nil {
		fmt.Print("listen error: ", err, "\n")
		os.Exit(1)
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Print("listen error: ", err, "\n")
			os.Exit(1)
		}
		onConnect(conn)
	}
}

func ReadFromConnection(conn net.Conn) (string, error) {
	buffer := make([]byte, 1024)
	size, err := conn.Read(buffer)
	if err != nil {
		return "", err
	}

	peerIp := bytes.NewBuffer(buffer).String()[:size]
	return peerIp, nil
}
