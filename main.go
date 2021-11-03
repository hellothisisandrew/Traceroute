/*//-------------------------------Notes------------------------------
-Change from depracated "syscall" library to sys 
*///------------------------------Notes-------------------------------



package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
)



func main() {
	if err := TraceRoute(os.Args[1]); err != nil {
		fmt.Println("Command line error")
	}
}

//Traceroute Open send/recieve sockets
func TraceRoute(website string) error {														//Maybe make this return the sockets themselves

	hops := 1 //This is the TTL
	timeValue := syscall.NsecToTimeval(1000 * 1000 * (int64)(2000)) //get timevalue


	//Open both outgoing and incoming sockets //DGRAM is a connectionless socket
	outGoingSocket,errA := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP) //openoutgoing socket --add error

	//The incoming socket must be raw so the headers aren't stripped
	incomingSocket, errB := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP) //Open recieve socke --Add error

	//defer closure of sockets
	defer syscall.Close(outGoingSocket)
	defer syscall.Close(incomingSocket) //Close only after enclosing function returns

	if errA != nil || errB != nil {
		fmt.Println(errB)
	}


	//Loop over until you reach your destination or you exceed maximum number of hops
	for hops <= 255 {

		syscall.SetsockoptInt(outGoingSocket, 0x0, syscall.IP_TTL, hops)
		syscall.SetsockoptTimeval(incomingSocket, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &timeValue)

		socketAdd, errC := getSocAddress() //Add error
		destAddArr, destAddStr, errD := getDestinationAddress(website) //Get the Destination address //gets both array Uint version as well as string version
		if errC != nil {
			fmt.Println("errC")
		}
		if errD != nil {
			fmt.Println("errD")
		}
		//Bind and send
		errE := syscall.Bind(incomingSocket, &syscall.SockaddrInet4{Port: 45000, Addr: socketAdd}) //Bind --add error
		errF := syscall.Sendto(outGoingSocket, []byte{0x0}, 0, &syscall.SockaddrInet4{Port: 45000, Addr: destAddArr})	//send the message --add error

		//
		if errE != nil {
			fmt.Println("Failed to Bind error:   ", errE )

		}

		if errF != nil {
			fmt.Println("errF")
		}

		packet := make([]byte, 128) // packet size
		_, from, errG := syscall.Recvfrom(incomingSocket, packet, 0)//Get socket address type

		if errG != nil {
			fmt.Println(hops, ". ","*")
			hops += 1
			continue
		}
		//get ip address and then lookup the host name
		ipAd := from.(*syscall.SockaddrInet4).Addr//grab the incoming addr

		ipStr := fmt.Sprintf("%v.%v.%v.%v", ipAd[0], ipAd[1], ipAd[2], ipAd[3])//convert string
		host, _ := net.LookupAddr(ipStr)//look up host name

		fmt.Println(hops,". ","host: ", host, "IPaddress: ", ipStr)

		//if you arrive at your destination break

		if ipStr == destAddStr {
			break
		}

		hops += 1 //Incrament hops

	}

	return nil
}
//Find a socket address
func getSocAddress() ([4]uint8, error) {

	socketAddress := [...]uint8{0, 0, 0, 0}//Socket adress

	address, err := net.InterfaceAddrs() //Get unicast interface address

	if err != nil {
		return socketAddress, err
	}
	//find an address loop
	for _, add := range address {
		if ipnet, ok := add.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {	//Make sure it isn't a loopback address
			if len(ipnet.IP.To4()) == net.IPv4len {				//Check If it is IPv4
				copy(socketAddress[:], ipnet.IP.To4())			//copy the socket address
				return socketAddress, nil
			}
		}
	}

	err = errors.New("No Internet")
	return socketAddress, err
}
//Get destination from string
func getDestinationAddress(dst string) (ipArr [4]byte,ipString string, e error) {
	destAddress := [...]uint8{0, 0, 0, 0}
	var address string					//Variable of address type

	add, err := net.LookupHost(dst) //Lookup the host IP //Can return multiple addresses if more than one exist

	if err != nil {
		return destAddress,"--", err
	}
	if len(add[0]) == net.IPv4len { //Use the IPv4 This can change
		address = add[0]
	}else {
		address = add[1]
	}
	ipAddress, err := net.ResolveIPAddr("ip", address) //Grab the actual IP

	if err != nil {
		return destAddress, "--", err
	}

	copy(destAddress[:], ipAddress.IP.To4()) //Copy to dest and return

	return destAddress, address, nil

}

























