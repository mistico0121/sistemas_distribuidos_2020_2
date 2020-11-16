package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

const (
	address = "127.0.0.1:3000"
)

const (
	maxDatagramSize = 8192
)

func main() {
	addr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		log.Fatal(err)
	}

	// Dial connects to the address on the named network.
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	count := 0
	for {
		time.Sleep(50 * time.Millisecond)
		count++
		conn.Write([]byte(strconv.Itoa(count)))
		fmt.Println("Sent Message, " + strconv.Itoa(count))
	}
}
