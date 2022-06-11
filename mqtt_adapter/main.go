package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"net"
	"sync"
	"time"
)

const (
	brokerProtocol = "tcp"
	brokerPort     = 1883
	clientID       = "testAdapter"
	topic          = "mqtt-sensor-data"
	mqttQosBit     = 2 // Quality of Service: exactly once
)

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

	// Subscribe to a topic
	if token := client.Subscribe(topic, mqttQosBit, simpleMessageHandler); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Unsubscribe()
	log.Printf("Subscribed to topic: %s\n", topic)

	// block until process is canceled
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

// message hander that prints the payload of received messages to logger
func simpleMessageHandler(client mqtt.Client, message mqtt.Message) {
	log.Printf("Message received: %s\n", message.Payload())
}
