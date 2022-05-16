package main

import (
	"log"
	"net"
	"os"
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

var TestLogger *log.Logger

type Sensor struct {
	Id       string
	Type     string
	Addr     net.UDPAddr
	DataPort int
}

const (
	MAX_LENGTH        int = 32
	REQUEST_INTERVALL     = 3 * time.Second //Time delay between requesting data from all sensors
	REGISTRATION_PORT     = 5000
)

type SensorCollection struct {
	sensors map[string]Sensor
	mutex   sync.RWMutex
}

func init() {
	file, err := os.OpenFile("test.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	TestLogger = log.New(file, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func InitSensors() SensorCollection {
	return SensorCollection{
		sensors: make(map[string]Sensor),
		mutex:   sync.RWMutex{},
	}
}

var registeredSensors SensorCollection

func main() {

	addr, err := net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(REGISTRATION_PORT))
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
	log.Printf("Listening on port %d for sensor registrations...\n", REGISTRATION_PORT)
	for {
		var buf [MAX_LENGTH]byte
		length, addr, err := conn.ReadFromUDP(buf[0:])
		log.Printf("New sensor with Addr %v requests registraton: %s\n", addr, buf)
		if err != nil {
			panic(err)
		}
		sensorData := strings.Split(string(buf[:]), "|")
		var sensor Sensor
		sensor.Type = sensorData[0]
		sensor.Id = sensorData[1]
		sensor.DataPort, err = strconv.Atoi(sensorData[2])
		if err != nil {
			panic(err)
		}
		sensor.Addr = *addr

		//Add new sensor to map
		registeredSensors.mutex.Lock()
		registeredSensors.sensors[sensor.Id] = sensor
		registeredSensors.mutex.Unlock()

		_, err = conn.WriteToUDP(buf[0:length], addr)
		if err != nil {
			panic(err)
		}
		log.Printf("sensor registriert: %v total:%d\n", sensor, len(registeredSensors.sensors))
	}
}

func pollRegisteredSensors() {
	//Todo maybe start every request in separate go routine
	for {
		time.Sleep(REQUEST_INTERVALL)
		successfullRequests := 0
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

			//log.Printf("Requesting data from sensor ID: %s & Addr: %v\n", currentSensor.Id, addr)
			_, err = conn.Write(buf[0:])
			if err != nil {
				log.Printf("Failed to request data from sensor ID: %s with Addr: %v\n", currentSensor.Id, addr)
				continue
				//panic(err)
			}

			//wait for data sent from sensors
			var dataBuffer [MAX_LENGTH]byte
			length, err := conn.Read(dataBuffer[0:])
			if err != nil {
				log.Printf("Failed to recieve data from sensor ID: %s with Addr: %v\n", currentSensor.Id, addr)
				//panic(err)
				continue
			}
			log.Printf("Data Recieved from sensor ID:%s: %s\n", currentSensor.Id, dataBuffer[0:length])
			successfullRequests++
			conn.Close()
			log.Printf("RTT:%v\n", time.Since(timeBefore))
			TestLogger.Printf("RTT:%v\n", time.Since(timeBefore))
		}
		log.Printf("Successfully requested data from %d out of total %d registered sensors\n", successfullRequests, len(registeredSensors.sensors))
		registeredSensors.mutex.RUnlock()
	}
}
