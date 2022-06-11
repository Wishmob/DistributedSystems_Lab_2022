package httpInterface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

type SensorDataPackage struct {
	Timestamp   time.Time         `json:"timestamp"`
	SensorCount int32             `json:"sensorcount"`
	Data        map[string]string `json:"data"`
}

func NewSensorDataPackage() SensorDataPackage {
	return SensorDataPackage{
		Timestamp:   time.Now(),
		SensorCount: 0,
		Data:        make(map[string]string),
	}
}

func SendDataToCloudServer(data SensorDataPackage) {

	json_data, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}
	ips, err := net.LookupIP("cloud_server")
	if err != nil {
		log.Println("Cloud server could not be found.")
		return
	}
	//log.Printf("Sending data to cloud server at %s\n", ips[0])
	addr := fmt.Sprintf("http://%s:8080/post-data", ips[0])
	resp, err := http.Post(addr, "application/json",
		bytes.NewBuffer(json_data))
	if err != nil {
		log.Println(err)
	}
	log.Printf("Successfully sent package with timestamp %v to cloud server via HTTP Post. Got: %s\n", data.Timestamp, resp.Status)

}
