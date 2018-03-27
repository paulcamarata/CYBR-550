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
	sPass       string          //server password
	iFile       string          //input file
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

//authentication (server side) function
func authS(ser *net.UDPConn, remoteaddr *net.UDPAddr) {

	//send PASS_REQ and payload to client
	ser.WriteToUDP([]byte(PASS_REQ+epload), remoteaddr)

	//create empty 1000 byte splice
	l := make([]byte, 1000)

	//password attempt loop
	for i := 1; i < 3; i++ {

		//read from buffer
		n, remoteaddr, err := ser.ReadFromUDP(l)

		//check for error
		eC(err)

		//extract header from buffer
		header := string(l[0:2])

		//extract payload from header
		password := string(l[6:n])

		//only attempt to authenticate if header is correct
		if header == PASS_RESP {

			//check password validity
			if password == sPass {

				//send client password success header
				ser.WriteToUDP([]byte(PASS_ACCEPT+epload), remoteaddr)

				//authentication successful server side log statement
				log.Println("Password accepted", "Password attempt #", i)

				//go to send file function
				sF(ser, remoteaddr)

			} else {

				//authentication failure server side log statement
				log.Println("Bad Password: Attempt #", i)

				//send client password rejection header
				ser.WriteToUDP([]byte(REJECT+epload), remoteaddr)
			}
		} else {
			//unnhandled header
			log.Fatal("Unhandled header: ABORT")
		}
	}

	//if three bad passwords, send TERMINATE header
	ser.WriteToUDP([]byte(TERMINATE+epload), remoteaddr)

	//last action, return to listening state
	defer listenS(ser)
}

//listen and wait function
func listenS(ser *net.UDPConn) {

	//listener loop
	for {

		//create a 100 byte buffer
		p := make([]byte, 1000)

		//read buffer
		n, remoteaddr, err := ser.ReadFromUDP(p)

		//check for error
		eC(err)

		//extract header from buffer
		header := string(p[0:2])

		//total input from buffer
		input := string(p[:bytes.IndexByte(p, 0)])

		//Server side logging statement
		log.Println("rAddr:", remoteaddr, "payload =", input, "length =", n)

		//handle potential cases
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
