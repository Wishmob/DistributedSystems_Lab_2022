package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

const (
	brokerProtocol = "tcp"
	brokerPort     = 1883
	clientID       = "testSensor"
	topic          = "mqtt-sensor-data"
	mqttQosBit     = 2 // Quality of Service: exactly once

)

func generateData() int {
	max := 100
	min := 10
	randomData := rand.Intn(max-min) + min
	return randomData
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
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
	options.SetClientID(clientID)

	client := mqtt.NewClient(options)

	time.Sleep(5 * time.Second) //give broker time to start

	// Connect to MQTT broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(0)
	log.Printf("Connected to MQTT broker: %s\n", brokerURI)

	for {
		sendData(&client)
		time.Sleep(5 * time.Second)
	}
}

func sendData(client *mqtt.Client) {
	data := strconv.Itoa(generateData())
	if token := (*client).Publish(topic, mqttQosBit, false, data); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	log.Printf("Published data: %d\n", data)
}
