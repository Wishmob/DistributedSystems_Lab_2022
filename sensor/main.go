package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
)

const (
	MAX_LENGTH int = 1024
)

func main() {
	var host string
	var port int
	flag.StringVar(&host, "host", "iot_gateway", "host the client should send udp packets to")
	flag.IntVar(&port, "port", 5000, "port the client should send udp packets to")

	addr, err := net.ResolveUDPAddr("udp4", host+":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	log.Printf("sending message to %v\n", addr)
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	//reader := bufio.NewReader(os.Stdin)
	//fmt.Print("Enter message: ")
	//request, _ := reader.ReadString('\n')

	_, err = conn.Write([]byte("fuck my life"))
	if err != nil {
		panic(fmt.Sprintf("sensor write failed %v", err))
	}

	var buf [MAX_LENGTH]byte
	length, err := conn.Read(buf[0:])
	if err != nil {
		panic(fmt.Sprintf("sensor read failed %v", err))
	}
	fmt.Printf("Reply is: %s\n", buf[0:length])
}
