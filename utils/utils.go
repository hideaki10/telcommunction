package utils

import (
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strings"
)

func GetLocalAddress() (currentIpAddress, currentMacAddress []byte) {
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
				currentIpAddress = ipnet.IP
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

	currentMacAddress = hwAddr

	fmt.Printf("Physical hardware address : %s , %s \n", name, hwAddr.String())

	return currentIpAddress, currentMacAddress
}

func ToByteArr(value interface{}) []byte {

	vr := reflect.ValueOf(value)

	var arr []byte

	for i := 0; i < vr.NumField(); i++ {
		b := vr.Field(i).Interface().([]byte)
		arr = append(arr, b...)
	}

	return arr

}
