package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

const (
	brokerProtocol      = "tcp"
	brokerPort          = 1883
	topic               = "mqtt-sensor-data"
	mqttQosBit          = 2                // Quality of Service: exactly once
	DataPublishInterval = 0 * time.Second  //Time delay between publishing data
	DataPublishDelay    = 10 * time.Second //Time delay before the mqtt sensor starts publishing data after start
)

var sensorID string

func generateData() int {
	max := 100
	min := 10
	randomData := rand.Intn(max-min) + min
	return randomData
}

type SensorData struct {
	SensorID  string    `json:"sensorid"`
	Timestamp time.Time `json:"timestamp"`
	Data      int       `json:"data"`
}

func NewSensorData() SensorData {
	return SensorData{
		SensorID:  sensorID,
		Timestamp: time.Now(),
		Data:      generateData(),
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	time.Sleep(DataPublishDelay) //give broker time to start

	ips, err := net.LookupIP("mosquitto_broker")
	if err != nil {
		log.Println("Mosquitto broker could not be found.")
		return
	}
	brokerAddress := ips[0]

	// create the broker string
	brokerURI := fmt.Sprintf("%s://%s:%d", brokerProtocol, brokerAddress, brokerPort)
	// create and configure the client options
	options := mqtt.NewClientOptions()
	options.AddBroker(brokerURI)
	sensorID = fmt.Sprintf("sensor[%s]", os.Getenv("HOSTNAME"))
	options.SetClientID(sensorID)

	client := mqtt.NewClient(options)

	// Connect to MQTT broker
	token := client.Connect()
	for token.Wait() && token.Error() != nil {
		log.Println(token.Error())
		time.Sleep(3 * time.Second)
		token = client.Connect()
	}
	defer client.Disconnect(0)
	log.Printf("Connected to MQTT broker: %s\n", brokerURI)

	for {
		sendData(&client)
		time.Sleep(DataPublishInterval)
	}
}

func sendData(client *mqtt.Client) {
	data := NewSensorData()
	dataJSON, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
	}
	if token := (*client).Publish(topic, mqttQosBit, false, dataJSON); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
	}
	log.Printf("Published data: %v\n", data)
}
