package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

//The Different sensor types
const (
	Temperature string = "TMP"
	Humidity           = "HUM"
	Brightness         = "BRT"
)

type Sensor struct {
	Id       int
	Type     string
	Addr     net.UDPAddr
	DataPort int
}

const (
	MAX_LENGTH int = 1024
)

type SensorCollection struct {
	sensors map[int]Sensor
	mutex   sync.RWMutex
}

func InitSensors() SensorCollection {
	return SensorCollection{
		sensors: make(map[int]Sensor),
		mutex:   sync.RWMutex{},
	}
}

var registeredSensors SensorCollection

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

	registeredSensors = InitSensors()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listenForSensorRegistration(conn)
	}()

	wg.Add(1)
	time.Sleep(3 * time.Second) //Todo remove
	go func() {
		defer wg.Done()
		pollRegisteredSensors()
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
		sensorData := strings.Split(string(buf[:]), "|")
		var sensor Sensor
		sensor.Type = sensorData[0]
		sensor.Id, err = strconv.Atoi(sensorData[1])
		if err != nil {
			panic(err)
		}
		sensor.DataPort, err = strconv.Atoi(sensorData[2])
		if err != nil {
			panic(err)
		}
		sensor.Addr = *addr

		//Add new sensor to map
		//TODO add mutex
		registeredSensors.mutex.Lock()
		registeredSensors.sensors[sensor.Id] = sensor
		registeredSensors.mutex.Unlock()

		_, err = conn.WriteToUDP(buf[0:length], addr)
		if err != nil {
			panic(err)
		}
		log.Printf("sensor registriert : %v\n", sensor)
	}
}

func pollRegisteredSensors() {
	//Todo maybe start every request in separate go routine
	for {
		time.Sleep(3 * time.Millisecond) //Todo remove
		registeredSensors.mutex.RLock()
		for _, currentSensor := range registeredSensors.sensors {
			buf := [1]byte{1}
			timeBefore := time.Now()
			addr, err := net.ResolveUDPAddr("udp4", currentSensor.Addr.IP.String()+":"+strconv.Itoa(currentSensor.DataPort))
			conn, err := net.DialUDP("udp", nil, addr)
			if err != nil {
				log.Printf("Could not create udp socket with address %v: %v\n", addr, err)
				continue
				//panic(err)
			}

			log.Printf("Requesting data from sensor ID: %d & Addr: %v\n", currentSensor.Id, addr)
			_, err = conn.Write(buf[0:])
			if err != nil {
				log.Printf("Failed to request data from sensor ID: %d with Addr: %v\n", currentSensor.Id, addr)
				continue
				//panic(err)
			}

			//wait for data sent from sensors
			var dataBuffer [MAX_LENGTH]byte
			length, err := conn.Read(dataBuffer[0:])
			if err != nil {
				log.Printf("Failed to recieve data from sensor ID: %d with Addr: %v\n", currentSensor.Id, addr)
				//panic(err)
				continue
			}
			log.Printf("Data Recieved from sensor ID:%d: %s\n", currentSensor.Id, dataBuffer[0:length])
			conn.Close()
			log.Printf("RTT:%v\n", time.Since(timeBefore))
		}
		registeredSensors.mutex.RUnlock()
	}
}
