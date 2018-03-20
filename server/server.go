package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"os"
)

const (
	Data        uint16 = 5 //because
	HdrSize     int    = 2
	PyldLenSize int    = 4
	PackIDSize  int    = 4
)

var p = make([]byte, 2048)

//error checking function, crash on fail
func eC(err error) {
	if err != nil {
		log.Fatalln("Fatal error:", err)
	}
}
func sF(ser *net.UDPConn, remoteaddr *net.UDPAddr) {
	f, err := os.Open("./somefile")
	defer f.Close()
	eC(err)

	fi, err := f.Stat()
	eC(err)

	dat := make([]byte, 1000)
	s := fi.Size()
	packID := uint32(0)
	for i := int64(0); i < s; {
		n, err := f.Read(dat)
		eC(err)
		packLen := HdrSize + PyldLenSize + PackIDSize + n
		pack := make([]byte, packLen)
		binary.LittleEndian.PutUint16(pack[0:], Data)
		binary.LittleEndian.PutUint32(pack[2:], uint32(n))
		binary.LittleEndian.PutUint32(pack[6:], packID)
		copy(pack[10:], dat[0:n])
		ser.WriteToUDP(pack, remoteaddr)
		i += int64(n)
		packID++
	}

	listenS(ser)
}

func authS(ser *net.UDPConn, remoteaddr *net.UDPAddr) {
	ser.WriteToUDP([]byte("PASS_REQ"), remoteaddr)
	for i := 0; i <= 1; i++ {
		_, remoteaddr, err := ser.ReadFromUDP(p)
		eC(err)
		input := string(p[:bytes.IndexByte(p, 0)])
		if input == "password" {
			ser.WriteToUDP([]byte("PASS_ACC"), remoteaddr)
			log.Println("Password accepted")
			sF(ser, remoteaddr)
		} else {
			log.Println("Bad Password: Attempt", i)
			ser.WriteToUDP([]byte("PASS_REQ"), remoteaddr)
		}
	}
	ser.WriteToUDP([]byte("REJECT"), remoteaddr)
	defer listenS(ser)
}

func listenS(ser *net.UDPConn) {
	for {
		_, remoteaddr, err := ser.ReadFromUDP(p)
		eC(err)

		input := string(p[:bytes.IndexByte(p, 0)])
		log.Println("rAddr: ", remoteaddr, "payload = ", input)

		if input == "JOIN_REQ" {
			authS(ser, remoteaddr)
		} else {
			ser.WriteToUDP([]byte("bad connection"), remoteaddr)
			log.Println("false")
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
	log.Println("Server started successfully")

	listenS(ser)

}

//initialize listener? (after fail for example)
//authloop, client gets 4 tries (handle this better)
//need help with the file transfer
