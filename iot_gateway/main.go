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
	REQUEST_INTERVAL      = 3 * time.Second //Time delay between requesting data from all sensors
	REGISTRATION_PORT     = 5000
)

type SensorCollection struct {
	sensors map[string]Sensor
	mutex   sync.RWMutex
}

func init() {
	file, err := os.OpenFile("/logs/test.log", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}

	//TestLogger = log.New(file, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)
	TestLogger = log.New(file, "", 0)
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
	TestLogger.Printf("RequestInterval: %v\n", REQUEST_INTERVAL)
	TestLogger.Printf("RTT\n") //write column names to log file
	for {
		var rtts []time.Duration
		time.Sleep(REQUEST_INTERVAL)
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
			length, err := conn.Read(dataBuffer[0:]) //Todo add controlled timout?
			if err != nil {
				log.Printf("Failed to recieve data from sensor ID: %s with Addr: %v\n", currentSensor.Id, addr)
				//panic(err)
				continue
			}
			log.Printf("Data Recieved from sensor ID:%s: %s\n", currentSensor.Id, dataBuffer[0:length])
			successfullRequests++
			conn.Close()
			rtt := time.Since(timeBefore)
			log.Printf("RTT:%v\n", rtt)
			//TestLogger.Printf("%v\n", rtt) //print rtt
			rtts = append(rtts, rtt)
		}
		TestLogger.Printf("Successful Requests: %d out of total %d registered sensors with avgRTT: %v minRTT: %v maxRTT: %v\n", successfullRequests, len(registeredSensors.sensors), durationAvg(&rtts), durationMinimum(&rtts), durationMaximum(&rtts))
		log.Printf("Successfully requested data from %d out of total %d registered sensors with an avg RTT of: %v\n", successfullRequests, len(registeredSensors.sensors), time.Duration(durationAvg(&rtts)))
		registeredSensors.mutex.RUnlock()
	}
}

func durationAvg(durations *[]time.Duration) time.Duration {
	var totalTime int64 = 0
	//todo save max and min rtt
	//TestLogger.Printf("totalRTT_bef: %v", totalTime)
	for _, dur := range *durations {
		totalTime += int64(dur)
	}
	//TestLogger.Printf("totalRTT: %d , total durations: %d, avg: %v", totalTime, int64(len(*durations)), time.Duration(totalTime/int64(len(*durations))))
	return time.Duration(totalTime / int64(len(*durations)))
}

func durationMaximum(durations *[]time.Duration) time.Duration {
	var maxTime int64 = 0
	for _, dur := range *durations {
		if int64(dur) > maxTime {
			maxTime = int64(dur)
		}
	}
	return time.Duration(maxTime)
}

func durationMinimum(durations *[]time.Duration) time.Duration {
	var minTime int64 = int64((*durations)[0])
	for _, dur := range *durations {
		if int64(dur) < minTime {
			minTime = int64(dur)
		}
	}
	return time.Duration(minTime)
}
