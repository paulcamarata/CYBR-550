package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
)

var reader = bufio.NewReader(os.Stdin)
var conn, err = net.Dial("udp", "127.0.0.1:8080")

func auth() {
	log.Println("true")
	fmt.Print("Enter password:")
	resp, err := reader.ReadString('\n')
	// check for error
	if err != nil {
		// log error and exit
		log.Fatalln(err)
	} else {
		// strip newline char from resp var
		resp = resp[:len(resp)-1]
		buff := []byte(resp)
		_, err = conn.Write(buff)
		// call next() method
		next()
	}
}

func next() {
	// do some shit
}

func main() {
	p := make([]byte, 2048)
	if err != nil {
		log.Fatalln(err)
	}

	msg := "JOIN_REQ"
	buf := []byte(msg)
	// for {

	_, err = conn.Write(buf)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = bufio.NewReader(conn).Read(p)
	log.Println(string(p))
	if err != nil {
		log.Fatalln(err)
	}
	if string(p[:bytes.IndexByte(p, 0)]) == "PASS_REQ" {
		log.Println("if")
		auth()
		// break
	} else {
		log.Println("else")
		log.Fatal("false")
	}
	log.Println("code kept executing")
	// time.Sleep(10 * time.Second)
	// }

	//  defer conn.Close()
}
