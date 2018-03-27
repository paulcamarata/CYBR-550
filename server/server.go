package main

import (
	"bytes"
	"crypto/sha1"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

var (
	JOIN_REQ    string = "JR"
	PASS_REQ    string = "PQ"
	PASS_RESP   string = "PR"
	PASS_ACCEPT string = "PA"
	DATA        string = "DA"
	TERMINATE   string = "TE"
	REJECT      string = "RE"
	epload      string = "1111" //empty payload
	dpload      string = "1112" //digest payload
	sPass       string
	iFile       string
)

//initialization funcion
func init() {
	//print line numbers for debugging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//error checking function, crash on fail
func eC(err error) {
	if err != nil {
		log.Fatalln("Fatal error, ABORT:", err)
	}
}

//send file function
func sF(ser *net.UDPConn, remoteaddr *net.UDPAddr) {
	//create a new empty buffer
	buf := bytes.NewBuffer(nil)

	//open file on local system
	f, err := os.Open(iFile)

	//check for error
	eC(err)

	io.Copy(buf, f)

	//server side logging statement
	log.Println("Attempting to send:", f.Name())

	//stop working with file
	f.Close()

	//converting buffer to string for later processing
	s := string(buf.Bytes())

	//sending file
	_, err = ser.WriteToUDP([]byte(DATA+epload+s), remoteaddr)

	//check for error
	eC(err)

	//build sha1 digest
	r := sha1.Sum([]byte(s))

	//possibly for the future, if max payload exceeded then crash
	//max := make([]byte, len(DATA)+len(dpload)+sha1.Size)

	//build an empty buffer
	buff := []byte{}

	//append packet type to buffer
	buff = append(buff, DATA...)

	//append payload type to buffer
	buff = append(buff, dpload...)

	//build checksum
	for i := 1; i < sha1.Size; i++ {
		buff = append(buff, r[i])
	}

	//server side logging statement
	log.Println("Sending checksum")

	//send checksum
	_, err = ser.WriteToUDP(buff, remoteaddr)

	//
	log.Println("Checksum sent")

	//check for error
	eC(err)

	//last action; return to listen state
	defer listenS(ser)
}

func authS(ser *net.UDPConn, remoteaddr *net.UDPAddr) {
	ser.WriteToUDP([]byte(PASS_REQ+epload), remoteaddr)
	l := make([]byte, 2048)

	for i := 1; i < 3; i++ {
		n, remoteaddr, err := ser.ReadFromUDP(l)
		eC(err)

		header := string(l[0:2])
		password := string(l[6:n])

		if header == PASS_RESP {
			if password == sPass {
				ser.WriteToUDP([]byte(PASS_ACCEPT+epload), remoteaddr)
				log.Println("Password accepted", "Password attempt #", i)
				sF(ser, remoteaddr)
			} else {
				log.Println("Bad Password: Attempt #", i)
				ser.WriteToUDP([]byte(REJECT+epload), remoteaddr)
			}
		} else {
			//handle termination unhandled message
		}
	}
	ser.WriteToUDP([]byte(TERMINATE+epload), remoteaddr)
	defer listenS(ser)
}

func listenS(ser *net.UDPConn) {
	for {
		p := make([]byte, 2048)
		n, remoteaddr, err := ser.ReadFromUDP(p)
		eC(err)
		header := string(p[0:2])

		input := string(p[:bytes.IndexByte(p, 0)])

		//Server side logging statement
		log.Println("rAddr:", remoteaddr, "payload =", input, "length =", n)

		switch header {
		case JOIN_REQ:
			//handle join request
			authS(ser, remoteaddr)
		default:
			//close connection
			ser.Close()

			//Server side log
			log.Println("ABORT")
		}
	}
}

func main() {
	//capture command arguments to be leveraged n connectoin creation
	clo := os.Args

	//verify correct number of arguments presented to initialize server
	if len(clo) != 4 {
		log.Fatal("Incorrect setup: ./server <server port> <password> <input file>")
	}

	//server log statement
	log.Println("Inital settings: port =", clo[1], "password =", clo[2], "input file =", clo[3])

	//save password to be used in authentication
	sPass = clo[2]

	//save file to be used in sendFile
	iFile = clo[3]

	//convert "port" string value to int
	port, err := strconv.Atoi(clo[1])

	//check for error
	eC(err)

	//define server lsten bind parameters
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("127.0.0.1"),
	}

	//server initialize log statement
	log.Println("Starting server on", addr)

	//start listener
	ser, err := net.ListenUDP("udp", &addr)

	//check for error
	eC(err)

	//listener handler
	defer listenS(ser)

	//server started log statement
	log.Println("Server started successfully")

}
