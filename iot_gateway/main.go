package main

import (
	"flag"
	"log"
	"net"
	"strconv"
)

const (
	MAX_LENGTH int = 1024
)

func main() {
	log.Println("Listening on port 5000 for udp packets...")
	var port int
	flag.IntVar(&port, "port", 5000, "port the iot_gateway should listen on for udp packets")

	addr, err := net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for {
		handleClient(conn)
	}
}

func handleClient(conn *net.UDPConn) {
	var buf [MAX_LENGTH]byte
	length, addr, err := conn.ReadFromUDP(buf[0:])
	log.Printf("Message recieved from %v with content: %s\n", addr, buf)
	if err != nil {
		panic(err)
	}
	_, err = conn.WriteToUDP(buf[0:length], addr)
	if err != nil {
		panic(err)
	}
}