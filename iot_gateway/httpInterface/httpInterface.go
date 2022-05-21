package httpInterface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
)

type SensorDataPackage struct {
	SensorCount int               `json:"sensorcount"`
	Data        map[string]string `json:"data"`
}

func NewSensorDataPackage() SensorDataPackage {
	return SensorDataPackage{
		SensorCount: 0,
		Data:        make(map[string]string),
	}
}

func SendDataToCloudServer(data SensorDataPackage) {

	json_data, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	ips, _ := net.LookupIP("cloud_server")
	addr := fmt.Sprintf("http://%s:8080/post-data", ips[0])
	_, err = http.Post(addr, "application/json",
		bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatal(err)
	}
}
