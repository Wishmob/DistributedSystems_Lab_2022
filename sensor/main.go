package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

type Socket struct {
	Host string
	Port int
}

const (
	MAX_LENGTH  int    = 1024
	SENSOR_TYPE string = "TMP"
	SENSOR_ID   int    = 1
)

func generateData() int {
	max := 100
	min := 10
	randomData := rand.Intn(max-min) + min
	return randomData
}

//registerToGateway sends a UDP package over the given socket to notify the gateway that it has been started
func registerToGateway(socket Socket) error {
	addr, err := net.ResolveUDPAddr("udp4", socket.Host+":"+strconv.Itoa(socket.Port))
	if err != nil {
		return err
	}
	log.Printf("Sending registration request to %v\n", addr)
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	//Sending registration request to gateway
	request := fmt.Sprintf("%s %d", SENSOR_TYPE, SENSOR_ID)
	_, err = conn.Write([]byte(request))
	if err != nil {
		return err
	}

	//check if registration was successfull
	var buf [MAX_LENGTH]byte
	length, err := conn.Read(buf[0:])
	if err != nil {
		return err
	}
	log.Printf("Registration to gateway worked. Got: %s\n", buf[0:length])
	return nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	var socket = Socket{}
	flag.StringVar(&socket.Host, "host", "iot_gateway", "host the client should send udp packets to")
	flag.IntVar(&socket.Port, "port", 5000, "port the client should send udp packets to")

	err := registerToGateway(socket)
	for err != nil {
		log.Printf("Registration to gateway failed: %v. Retrying in 5 seconds", err)
		time.Sleep(5 * time.Second)
		err = registerToGateway(socket)
	}

	addrForDataRequests, err := net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(7030))
	if err != nil {
		panic(err)
	}

	connForDataRequests, err := net.ListenUDP("udp", addrForDataRequests)
	if err != nil {
		panic(err)
	}
	defer connForDataRequests.Close()

	handleDataRequest(connForDataRequests)
	//
	//addr, err := net.ResolveUDPAddr("udp4", socket.Host+":"+strconv.Itoa(socket.Port))
	//if err != nil {
	//	panic(err)
	//}
	//log.Printf("sending message to %v\n", addr)
	//conn, err := net.DialUDP("udp", nil, addr)
	//if err != nil {
	//	panic(err)
	//}
	//defer conn.Close()
	//
	////reader := bufio.NewReader(os.Stdin)
	////fmt.Print("Enter message: ")
	////request, _ := reader.ReadString('\n')
	//
	//_, err = conn.Write([]byte("fuck my life"))
	//if err != nil {
	//	panic(fmt.Sprintf("sensor write failed %v", err))
	//}
	//
	//var buf [MAX_LENGTH]byte
	//length, err := conn.Read(buf[0:])
	//if err != nil {
	//	panic(fmt.Sprintf("sensor read failed %v", err))
	//}
	//fmt.Printf("Reply is: %s\n", buf[0:length])
}

func handleDataRequest(conn *net.UDPConn) {
	for {
		var buf [MAX_LENGTH]byte
		_, addr, err := conn.ReadFromUDP(buf[0:])
		log.Printf("Data Request recieved from %v\n", addr)
		if err != nil {
			panic(err)
		}
		data := generateData()
		//replace data sent with random data
		_, err = conn.WriteToUDP([]byte(strconv.Itoa(data)), addr)
		if err != nil {
			panic(err)
		}
	}
}
