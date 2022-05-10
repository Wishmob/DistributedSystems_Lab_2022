package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

//The Different sensor types
const (
	Temperature string = "TMP"
	Humidity           = "HUM"
	Brightness         = "BRT"
)

type Sensor struct {
	Id   string
	Type string
	Addr net.UDPAddr
}

const (
	MAX_LENGTH int = 1024
)

var registeredSensors map[string]Sensor

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

	registeredSensors = make(map[string]Sensor)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listenForSensorRegistration(conn)
	}()

	wg.Wait()
}

func listenForSensorRegistration(conn *net.UDPConn) {
	for {
		var buf [MAX_LENGTH]byte
		length, addr, err := conn.ReadFromUDP(buf[0:])
		log.Printf("Message recieved from %v with content: %s\n", addr, buf)
		if err != nil {
			panic(err)
		}
		sensorData := strings.Split(string(buf[:]), " ")
		var sensor Sensor
		sensor.Type = sensorData[0]
		sensor.Id = sensorData[1]
		sensor.Addr = *addr

		//Add new sensor to map
		registeredSensors[sensor.Id] = sensor

		_, err = conn.WriteToUDP(buf[0:length], addr)
		if err != nil {
			panic(err)
		}
		fmt.Printf("sensor registriert : %v\n", sensor)
	}
}
