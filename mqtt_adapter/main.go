package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	brokerProtocol     = "tcp"
	brokerPort         = 1883
	clientID           = "testAdapter"
	topic              = "mqtt-sensor-data"
	mqttQosBit         = 2               // Quality of Service: exactly once
	IotGatewayDataPort = 7777            //Port of iot gateway to which the adapter forwards the data received via mqtt
	SubscribeDelay     = 5 * time.Second //Time delay before the mqtt adapter subscribes to the data topic
)

type SensorData struct {
	SensorID  string    `json:"sensorid"`
	Timestamp time.Time `json:"timestamp"`
	Data      int       `json:"data"`
}

var udpAddrOfGateway *net.UDPAddr

func main() {

	time.Sleep(SubscribeDelay) //give broker & iot gateway time to start

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

	// resolve address von iot_gateway to forward data to
	udpAddrOfGateway, err = net.ResolveUDPAddr("udp4", "iot_gateway"+":"+strconv.Itoa(IotGatewayDataPort))
	if err != nil {
		log.Println(err)
	}

	// Connect to MQTT broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(0)
	log.Printf("Connected to MQTT broker: %s\n", brokerURI)

	// Subscribe to a topic
	if token := client.Subscribe(topic, mqttQosBit, processSensorMessages); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Unsubscribe()
	log.Printf("Subscribed to topic: %s\n", topic)

	// block until process is canceled
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

// processSensorMessages is the message handler that forwards the mqtt messages retrieved from the sensors to the iot gateway via udp
func processSensorMessages(client mqtt.Client, message mqtt.Message) {

	//*************
	//just to check if data arrives correctly at adapter
	var data SensorData
	err := json.Unmarshal(message.Payload(), &data)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Message received: %v\n", data)
	//*************

	// forward data to gateway
	conn, err := net.DialUDP("udp", nil, udpAddrOfGateway)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	_, err = conn.Write(message.Payload())
	if err != nil {
		log.Println(err)
	}
}
