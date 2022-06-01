package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"vs_praktikum_BreiterSchandl_Di2x/iot_gateway/httpInterface"
)

//The Different sensor types //currently not used
const (
	Temperature string = "TMP"
	Humidity           = "HUM"
	Brightness         = "BRT"
)

var (
	TestLoggerP1 *log.Logger
	TestLoggerP2 *log.Logger
)

type Sensor struct {
	Id       string
	Type     string
	Addr     net.UDPAddr
	DataPort int //The udp port on which the sensor will listen for data requests
}

const (
	ReadBufferSize   int = 32              //The size of the buffers used to read from the udp sockets
	RequestInterval      = 1 * time.Second //Time delay between requesting data from all sensors
	RequestDelay         = 3 * time.Second
	RegistrationPort     = 5000 //Port on which the gateway is listening for new sensors
)

type SensorCollection struct {
	sensors map[string]Sensor
	mutex   sync.RWMutex
}

var startTime time.Time

func Uptime() time.Duration {
	return time.Since(startTime)
}

func init() {
	startTime = time.Now()
	logfileP1, err := os.OpenFile("/logs/P1RttLog.log", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("log directory could not be created. Try creating it manually: %v\n", err)
	}
	//TestLoggerP1 = log.New(logfileP1, "TEST: ", log.Ldate|log.Ltime|log.Lshortfile)
	TestLoggerP1 = log.New(logfileP1, "", 0)
	logfileP2, err := os.OpenFile("/logs/P2RttLog.log", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("log directory could not be created. Try creating it manually: %v\n", err)
	}
	TestLoggerP2 = log.New(logfileP2, "", 0)
}

func InitSensors() SensorCollection {
	return SensorCollection{
		sensors: make(map[string]Sensor),
		mutex:   sync.RWMutex{},
	}
}

var registeredSensors SensorCollection

func main() {
	addr, err := net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(RegistrationPort))
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
	time.Sleep(RequestDelay) //Give sensors time to register before starting to request data
	go func() {
		defer wg.Done()
		pollRegisteredSensors()
	}()

	wg.Wait()
}

func listenForSensorRegistration(conn *net.UDPConn) {
	log.Printf("Listening on port %d for sensor registrations...\n", RegistrationPort)
	for {
		var buf [ReadBufferSize]byte
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
	//Todo maybe start every request in separate go routine?
	TestLoggerP1.Printf("RequestInterval: %v\n", RequestInterval)
	TestLoggerP1.Printf("Gateway Uptime, Successful Requests, total registered sensors, avgRTT, minRTT, maxRTT\n") //write column names to log file
	TestLoggerP2.Printf("RequestInterval: %v\n", RequestInterval)
	TestLoggerP2.Printf("Gateway Uptime, Sensor count, RTT\n") //write column names to log file

	for {
		var rtts []time.Duration

		time.Sleep(RequestInterval)
		if len(registeredSensors.sensors) == 0 {
			continue
		}
		dataPackage := httpInterface.NewSensorDataPackage()
		dataPackage.SensorCount = int32(len(registeredSensors.sensors))

		successfulRequests := 0
		registeredSensors.mutex.RLock()
		for _, currentSensor := range registeredSensors.sensors {
			timeBefore := time.Now()
			addr, err := net.ResolveUDPAddr("udp4", currentSensor.Addr.IP.String()+":"+strconv.Itoa(currentSensor.DataPort))
			conn, err := net.DialUDP("udp", nil, addr)
			if err != nil {
				log.Printf("Could not create udp socket with address %v: %v\n", addr, err)
				continue
				//panic(err)
			}

			//log.Printf("Requesting data from sensor ID: %s & Addr: %v\n", currentSensor.Id, addr)
			_, err = conn.Write([]byte{1})
			if err != nil {
				log.Printf("Failed to request data from sensor ID: %s with Addr: %v\n", currentSensor.Id, addr)
				continue
				//panic(err)
			}

			//wait for data sent from sensors
			var dataBuffer [ReadBufferSize]byte
			conn.SetDeadline(time.Now().Add(2 * time.Second))
			length, err := conn.Read(dataBuffer[0:])
			if err != nil {
				log.Printf("Failed to recieve data from sensor ID: %s with Addr: %v\n", currentSensor.Id, addr)
				//panic(err)
				continue
			}
			conn.SetDeadline(time.Time{})
			log.Printf("Data Recieved from sensor ID:%s: %s\n", currentSensor.Id, dataBuffer[0:length])
			successfulRequests++
			conn.Close()
			rtt := time.Since(timeBefore)
			log.Printf("RTT:%v\n", rtt)
			//TestLoggerP1.Printf("%v\n", rtt) //print rtt
			rtts = append(rtts, rtt)
			dataPackage.Data[currentSensor.Id] = string(dataBuffer[0:length])
		}
		TestLoggerP1.Printf("%v, %d, %d, %v, %v, %v\n", int(Uptime().Seconds()), successfulRequests, len(registeredSensors.sensors), durationAvg(&rtts).Microseconds(), durationMinimum(&rtts).Microseconds(), durationMaximum(&rtts).Microseconds())
		log.Printf("Successfully requested data from %d out of total %d registered sensors with an avgRTT: %v minRTT: %v maxRTT: %v\n", successfulRequests, len(registeredSensors.sensors), durationAvg(&rtts), durationMinimum(&rtts), durationMaximum(&rtts))
		registeredSensors.mutex.RUnlock()
		timeBeforePost := time.Now()
		httpInterface.SendDataToCloudServer(dataPackage)
		rttPost := time.Since(timeBeforePost)
		TestLoggerP2.Printf("%v, %d, %v\n", int(Uptime().Seconds()), len(registeredSensors.sensors), rttPost)

	}
}

func durationAvg(durations *[]time.Duration) time.Duration {
	if len(*durations) < 1 {
		return 0
	}
	var totalTime int64 = 0
	for _, dur := range *durations {
		totalTime += int64(dur)
	}
	return time.Duration(totalTime / int64(len(*durations)))
}

func durationMaximum(durations *[]time.Duration) time.Duration {
	if len(*durations) < 1 {
		return 0
	}
	var maxTime int64 = 0
	for _, dur := range *durations {
		if int64(dur) > maxTime {
			maxTime = int64(dur)
		}
	}
	return time.Duration(maxTime)
}

func durationMinimum(durations *[]time.Duration) time.Duration {
	if len(*durations) < 1 {
		return 0
	}
	var minTime = int64((*durations)[0])
	for _, dur := range *durations {
		if int64(dur) < minTime {
			minTime = int64(dur)
		}
	}
	return time.Duration(minTime)
}
