package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
	_, err := conn.WriteToUDP([]byte("From server: Hello I got your mesage "), addr)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
}

func main() {
	p := make([]byte, 2048)
	addr := net.UDPAddr{
		Port: 8080,
		IP:   net.ParseIP("127.0.0.1"),
	}
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}

	// var success = false

	for {
		_, remoteaddr, err := ser.ReadFromUDP(p)
		if err != nil {
			log.Fatalln(err)
		}

		var input = string(p[:bytes.IndexByte(p, 0)])
		log.Println("rAddr: ", remoteaddr, "payload = ", input)

		log.Println("input is " + input)
		if input == "JOIN_REQ" {
			ser.WriteToUDP([]byte("PASS_REQ"), remoteaddr)
			log.Println("true")
		} else if input == "password" {
			ser.WriteToUDP([]byte("password entered successfully"), remoteaddr)
			log.Println("password entered")
		} else {
			ser.WriteToUDP([]byte("bad connection"), remoteaddr)
			log.Println("false")
			log.Println("failed: rAddr: ", remoteaddr, "payload = ", input)
		}
	}
}
