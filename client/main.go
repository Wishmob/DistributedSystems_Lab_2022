package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
)

const (
	MAX_LENGTH int = 1024
)

func main() {
	var host string
	var port int
	flag.StringVar(&host, "host", "localhost", "host the client should send udp packets to")
	flag.IntVar(&port, "port", 5000, "port the client should send udp packets to")

	addr, err := net.ResolveUDPAddr("udp4", host+":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter message: ")
	request, _ := reader.ReadString('\n')

	_, err = conn.Write([]byte(request))
	if err != nil {
		panic(err)
	}

	var buf [MAX_LENGTH]byte
	length, err := conn.Read(buf[0:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("Reply is: %s", buf[0:length])
}
