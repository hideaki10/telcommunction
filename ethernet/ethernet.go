package ethernet

type Ethernet struct {
	Destination []byte
	Source      []byte
	Ethtype     []byte
}

func NewEthernet(destinationMacAddr, sourceMacAddr []byte, ethType string) Ethernet {

	ethernet := Ethernet{
		Destination: destinationMacAddr,
		Source:      sourceMacAddr,
	}

	switch ethType {
	case "IPv4":
		ethernet.Ethtype = []byte{0x08, 0x00}
	case "ARP":
		ethernet.Ethtype = []byte{0x08, 0x06}
	}

	return ethernet
}
