package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"os"
)

var (
	JOIN_REQ    string = "JR"
	PASS_REQ    string = "PQ"
	PASS_RESP   string = "PR"
	PASS_ACCEPT string = "PA"
	DATA        string = "DA"
	TERMINATE   string = "TE"
	REJECT      string = "RE"
	p                  = make([]byte, 2048)
)

//error checking function, crash on fail
func eC(err error) {
	if err != nil {
		log.Fatalln("Fatal error:", err)
	}
}

//send file function
func sF(ser *net.UDPConn, remoteaddr *net.UDPAddr) {
	buf := bytes.NewBuffer(nil)
	f, _ := os.Open("./somefile") // Error handling elided for brevity.
	io.Copy(buf, f)               // Error handling elided for brevity.
	log.Println(f.Name())
	f.Close()

	s := string(buf.Bytes())
	log.Println(s)
	ser.WriteToUDP([]byte(DATA), remoteaddr)

	ser.WriteToUDP(buf.Bytes(), remoteaddr)
	defer listenS(ser)
}

func authS(ser *net.UDPConn, remoteaddr *net.UDPAddr) {
	ser.WriteToUDP([]byte(PASS_REQ), remoteaddr)
	for i := 0; i <= 1; i++ {
		_, remoteaddr, err := ser.ReadFromUDP(p)
		eC(err)
		input := string(p[:bytes.IndexByte(p, 0)])
		if input == "password" {
			ser.WriteToUDP([]byte(PASS_ACCEPT), remoteaddr)
			log.Println("Password accepted")
			sF(ser, remoteaddr)
		} else {
			log.Println("Bad Password: Attempt", i)
			ser.WriteToUDP([]byte(PASS_REQ), remoteaddr)
		}
	}
	ser.WriteToUDP([]byte(REJECT), remoteaddr)
	defer listenS(ser)
}

func listenS(ser *net.UDPConn) {
	for {
		n, remoteaddr, err := ser.ReadFromUDP(p)
		eC(err)

		input := string(p[:bytes.IndexByte(p, 0)])
		log.Println("rAddr:", remoteaddr, "payload =", input, "length =", n)

		if bytes.HasPrefix([]byte(input), []byte(JOIN_REQ)) {
			authS(ser, remoteaddr)
		} else if bytes.HasPrefix([]byte(input), []byte(PASS_RESP)) {
			// PASS_ACCEPT()
		} else {
			//handle rejection function
			ser.WriteToUDP([]byte(REJECT), remoteaddr)
			log.Println("failed: rAddr: ", remoteaddr, "payload = ", input)
		}
	}
}

func main() {

	addr := net.UDPAddr{
		Port: 8080,
		IP:   net.ParseIP("127.0.0.1"),
	}

	log.Println("Starting server on", addr)
	ser, err := net.ListenUDP("udp", &addr)
	eC(err)

	defer listenS(ser)

	log.Println("Server started successfully")

}
