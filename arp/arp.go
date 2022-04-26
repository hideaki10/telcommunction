package arp

import (
	"fmt"
	"log"
	"net"
	"syscall"
	"telcommunction/ethernet"
	"telcommunction/utils"
)

type Arp struct {
	HardwareType          []byte
	ProtcolType           []byte
	HardwareAddressLength []byte
	ProtcolAddressLength  []byte
	Operation             []byte
	SenderMacAddress      []byte
	SenderIpAddress       []byte
	TargetMacAddress      []byte
	TargetIpAddress       []byte
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

func NewArpRequest(localMacAddress, localIPAddress, targetIpAddress []byte) Arp {

	arp := Arp{

		//ethernet 0x0001
		HardwareType: []byte{0x00, 0x01},

		// ip 0x0800
		ProtcolType: []byte{0x08, 0x00},

		// 0x06
		HardwareAddressLength: []byte{0x06},

		// 0x04
		ProtcolAddressLength: []byte{0x04},

		// arp request 0x0001 arp replay 0x0002
		Operation: []byte{0x00, 0x01},

		//

		SenderMacAddress: localMacAddress,
		SenderIpAddress:  localIPAddress,

		//
		//  brodercast
		TargetMacAddress: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},

		TargetIpAddress: targetIpAddress,
	}
	return arp

}

func (arp Arp) send(ifindex int, packet []byte) Arp {

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
		// recvform open the socket fd reads and store the data in buf
		// recvfrom() places the received message into the buffer buf.  The caller must specify the size of the buffer in len.
		//
		// flags is formed by ORing one or other
		// https://man7.org/linux/man-pages/man2/recv.2.html
		_, _, err := syscall.Recvfrom(fd, recvBuf, 0)
		if err != nil {
			log.Fatalln("read err ")
		}

		fmt.Println(recvBuf)
		if recvBuf[12] == 0x08 && recvBuf[13] == 0x06 {
			if recvBuf[20] == 0x00 && recvBuf[21] == 0x02 {
				return parseArpPacket(recvBuf[14:])
			}
		}

	}

}

func parseArpPacket(packet []byte) Arp {
	return Arp{
		HardwareType:          []byte{packet[0], packet[1]},
		ProtcolType:           []byte{packet[2], packet[3]},
		HardwareAddressLength: []byte{packet[4]},
		ProtcolAddressLength:  []byte{packet[5]},
		Operation:             []byte{packet[6], packet[7]},
		SenderMacAddress:      []byte{packet[8], packet[9], packet[10], packet[11], packet[12], packet[13]},
		SenderIpAddress:       []byte{packet[14], packet[15], packet[16], packet[17]},
		TargetMacAddress:      []byte{packet[18], packet[19], packet[20], packet[21], packet[22], packet[23]},
		TargetIpAddress:       []byte{packet[24], packet[25], packet[26], packet[27]},
	}
}

func ArpProtcol() {
	currentIpaddress, currentMacAddress := utils.GetLocalAddress()
	ethernet := ethernet.NewEthernet([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, currentMacAddress, "ARP")

	targetIpAddress := net.ParseIP("192.168.3.16")

	arpRequest := NewArpRequest(currentMacAddress, currentIpaddress, targetIpAddress)

	var sendArp []byte

	sendArp = append(sendArp, utils.ToByteArr(ethernet)...)
	sendArp = append(sendArp, utils.ToByteArr(arpRequest)...)
	netI, _ := net.InterfaceByName("ens33")
	index := netI.Index
	arpreply := arpRequest.send(index, sendArp)
	fmt.Printf("ARP Reply :%s\n", arpreply.SenderMacAddress)
}
