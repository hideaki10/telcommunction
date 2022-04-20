package arp

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
)

type arp struct {
	hardwareType          []byte
	protcolType           []byte
	hardwareAddressLength []byte
	protcolAddressLength  []byte
	operation             []byte
	senderMacAddress      []byte
	senderIpAddress       []byte
	targetMacAddress      []byte
	targetIpAddress       []byte
}

// host order to network order
// htons -> 2byte
// htonl -> 4byte
// https://stackoverflow.com/questions/19207745/htons-function-in-socket-programing
// https://zhuanlan.zhihu.com/p/30955591
// >>
// <<
// &  0&0 -> 0  1&0 -> 0
//    0&1 -> 0  1&1 -> 1
// |  0|0 -> 0  1|0 -> 1
//    0|1 -> 1  1|1 -> 1
func htons(host uint16) uint16 {
	//return (host&0xff)<<8 | (host >> 8)
	return (host<<8)&0xff00 | host>>8
}

func NewArpRequest(localMacAddress, localIPAddress, targetIpAddress []byte) *arp {

	arp := &arp{

		//ethernet 0x0001
		hardwareType: []byte{0x00, 0x01},

		// ip 0x0800
		protcolType: []byte{0x08, 0x00},

		// 0x06
		hardwareAddressLength: []byte{0x06},

		// 0x04
		protcolAddressLength: []byte{0x04},

		// arp request 0x0001 arp replay 0x0002
		operation: []byte{0x00, 0x01},

		//

		senderMacAddress: localMacAddress,
		senderIpAddress:  localIPAddress,

		// mac address -> 00-00-00-00-00-00-00-00
		targetMacAddress: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},

		targetIpAddress: targetIpAddress,
	}
	return arp

}

func (arp *arp) send(ifindex int, packet []byte) arp {

	// syscall.ARPHARD_ETHER -> ethernet
	addr := syscall.SockaddrLinklayer{
		Protocol: syscall.ETH_P_ARP,
		Ifindex:  ifindex,
		Hatype:   syscall.ARPHRD_ETHER,
	}

	// AF_PACKET -> LOW-level packer interface
	// ETH_P_ALL ->
	// SOCK_RAW -> the socket for raw network protoctl access can be used for new communcation protocols
	// return fd,err
	// fd -> file descriptor
	//
	// https://man7.org/linux/man-pages/man2/socket.2.html
	// https://man7.org/linux/man-pages/man7/packet.7.html
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	if err != nil {
		log.Fatalln("create socket is failed ")
	}

	defer syscall.Close(fd)

	err = syscall.Sendto(fd, packet, 0, &addr)
	if err != nil {
		log.Fatal("send to err ")
	}

	for {
		recvBuf := make([]byte, 80)
		// recvform
		//
		_, _, err := syscall.Recvfrom(fd, recvBuf, 0)
		if err != nil {
			log.Fatalln("read err ")
		}

		return parseArpPacket(recvBuf)
	}

}

func parseArpPacket(packet []byte) arp {
	return arp{
		hardwareType:          []byte{packet[0], packet[1]},
		protcolType:           []byte{packet[2], packet[3]},
		hardwareAddressLength: []byte{packet[4]},
		protcolAddressLength:  []byte{packet[5]},
		operation:             []byte{packet[6], packet[7]},
		senderMacAddress:      []byte{packet[8], packet[9], packet[10], packet[11], packet[12], packet[13]},
		senderIpAddress:       []byte{packet[14], packet[15], packet[16], packet[17]},
		targetMacAddress:      []byte{packet[18], packet[19], packet[20], packet[21], packet[22], packet[23]},
		targetIpAddress:       []byte{packet[24], packet[25], packet[26], packet[27]},
	}
}

func arpProtocol() {

}

func GetLocalAddress() {
	addres, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalln("get local address is failed")
	}

	var currentIP, currentNetworkHardwareName string

	for _, addr := range addres {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println("Current IP address : ", ipnet.IP.String())
				currentIP = ipnet.IP.String()
			}
		}
	}

	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {
		addrs, err := interf.Addrs()
		if err != nil {
			log.Fatalln("get address is failed")
		}
		for _, adr := range addrs {
			if strings.Contains(adr.String(), currentIP) {
				currentNetworkHardwareName = interf.Name
			}
		}
	}

	netInterface, err := net.InterfaceByName(currentNetworkHardwareName)
	if err != nil {
		log.Fatalln(err)
	}

	name := netInterface.Name
	macAddress := netInterface.HardwareAddr

	hwAddr, err := net.ParseMAC(macAddress.String())
	if err != nil {
		fmt.Println("No able to parse MAC address :", err)
		os.Exit(1)
	}

	fmt.Printf("Physical hardware address : %s , %s \n", name, hwAddr.String())

}
