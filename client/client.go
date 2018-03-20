package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var reader = bufio.NewReader(os.Stdin)

//error checking function, crash on fail
func eC(err error) {
	if err != nil {
		log.Fatalln("Fatal error:", err)
	}
}

//join request function
func jR(conn net.Conn) {
	//initialize message to server
	msg := "JOIN_REQ"
	buf := []byte(msg)

	//send initial message
	_, err := conn.Write(buf)

	//check for error
	eC(err)

	//go to listen on connection function
	lC(conn)
}

//listen on connection function
func lC(conn net.Conn) {
	//define payload
	p := make([]byte, 2048)

	for {
		//get length of payload
		n, err := conn.Read(p)

		//check for error
		eC(err)

		//convert payload to string
		var input = string(p[:bytes.IndexByte(p, 0)])

		log.Println(n) //will be used for parsing bytespace

		if input == "PASS_REQ" {
			auth(conn)
		} else {
			log.Println(input)
			conn.Close()
		}

		time.Sleep(time.Second * 1)
	}
}

//function to handle cleartext password delivery
func auth(conn net.Conn) {
	//prompt for password
	fmt.Print("Enter password:")

	//read password from stdout
	resp, err := reader.ReadString('\n')

	//check for error
	eC(err)

	// strip newline char from resp var
	resp = resp[:len(resp)-1]

	//convert password to byte array
	buff := []byte(resp)

	//send password to server
	_, err = conn.Write(buff)

	//check for error
	eC(err)

	lC(conn)

}

func main() {
	conn, err := net.Dial("udp", "127.0.0.1:8080")
	eC(err)

	//if all actions complete close connection
	defer conn.Close()

	//go to join request function
	jR(conn)
}
