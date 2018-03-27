package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

//global varables
var (
	JOIN_REQ    string = "JR"
	PASS_REQ    string = "PQ"
	PASS_RESP   string = "PR"
	PASS_ACCEPT string = "PA"
	DATA        string = "DA"
	TERMINATE   string = "TE"
	REJECT      string = "RE"
	pload       string = "1111" //padding to meet spec
	oFile       string          //output file
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

//function to handle cleartext password delivery
func auth(conn net.Conn) {
	//prompt for password
	fmt.Print("Enter password:")

	//read password from stdout
	resp, err := bufio.NewReader(os.Stdin).ReadString('\n')

	//check for error
	eC(err)

	// strip newline char from resp var
	resp = PASS_RESP + pload + resp[:len(resp)-1]

	//convert password to byte array
	buff := []byte(resp)

	//send password to server
	_, err = conn.Write(buff)

	//check for error
	eC(err)

	//reurn to listening state
	lC(conn)

}

//listen on connection function
func lC(conn net.Conn) {

	//create byte slice for storing payload
	p := make([]byte, 1000)

	for {
		//get length of payload
		n, err := conn.Read(p)

		//check for error
		eC(err)

		//extract header from buffer
		header := string(p[0:2])

		//extract payload from buffer
		payload := string(p[2:6])

		//extract content from buffer
		therest := string(p[6:n])

		/* //Debugging code (comment out when not needed)
		log.Println("header=", header, "payload=", payload, "therest =", therest)
		log.Println(n)     //debug for length of byte array received
		log.Println(p[0:n]) //debug for content of byte array received
		*/

		//handle potential cases
		switch header {
		case PASS_REQ:

			//go to authentication
			auth(conn)

		case PASS_ACCEPT:

			//ready to receive data
			lC(conn)

		case DATA:

			//dynamic payload handler '1111' = data; '1112' = checksum
			if payload == "1111" {

				//output content to declared file
				err := ioutil.WriteFile(oFile, p[6:n], 0644)

				//check for error
				eC(err)

			} else if payload == "1112" {

				//create a new empty buffer
				buf := bytes.NewBuffer(nil)

				//open file on local system
				f, err := os.Open(oFile)

				//check for error
				eC(err)

				io.Copy(buf, f)

				//stop working with file
				f.Close()

				//converting buffer to string for later processing
				s := string(buf.Bytes())

				//build sha1 digest
				r := sha1.Sum([]byte(s))

				//build an empty buffer
				buff := []byte{}

				//build checksum on locally received flie
				for i := 1; i < sha1.Size; i++ {
					buff = append(buff, r[i])
				}

				//verify checksum on local file matches received checksum
				if string(buff) == therest {

					//log successful flie transfer and shutdown
					log.Fatalln("Checksum verified: OK")

				} else {

					//log unsuccessful file transfer and shut down
					log.Fatalln("Checksum failed: Abort")
				}
			}

		case REJECT:

			//password has been rejected
			log.Println("server rejected password, please try again")

			//return to authentication function
			auth(conn)

		case TERMINATE:

			//connection terminated
			log.Fatalln("ABORT")
		default:

			//unhandled header
			log.Fatalln("Unhandled header: ABORT")

		}
	}
}

//join request function
func jR(conn net.Conn) {

	//initialize message request to server
	buf := []byte(JOIN_REQ + pload)

	//send initial message
	_, err := conn.Write(buf)

	//check for error
	eC(err)

	//go to listen on connection function
	lC(conn)
}

func main() {

	//capture command arguments to be leveraged n connectoin creation
	clo := os.Args

	//verify correct number of arguments presented to initialize server
	if len(clo) != 4 {
		log.Fatal("Incorrect setup: ./client <server name> <server port> <output file>")
	}

	//capture desired output location
	oFile = clo[3]

	//server log statement
	log.Println("Inital settings: server =", clo[1], "port =", clo[2], "output file =", clo[3])

	//connect to server
	conn, err := net.Dial("udp", clo[1]+":"+clo[2])
	eC(err)

	//if all actions complete close connection
	defer conn.Close()

	//go to join request function
	jR(conn)
}
