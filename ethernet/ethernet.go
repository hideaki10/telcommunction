package main

type ethernet struct {
	destination []byte
	source      []byte
	ethtype     []byte
}

func NewEthernet(destinationMacAddr, sourceMacAddr []byte, ethType string) *ethernet {

	ethernet := &ethernet{
		destination: destinationMacAddr,
		source:      sourceMacAddr,
	}

	switch ethType {
	case "IPv4":
		ethernet.ethtype = []byte{0x08, 0x00}
	case "ARP":
		ethernet.ethtype = []byte{0x08, 0x06}
	}

	return ethernet
}
