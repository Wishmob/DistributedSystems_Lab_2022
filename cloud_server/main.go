package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	Port int = 8080
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

var sensorData []SensorDataPackage

func main() {

	//sensorData = make([]SensorDataPackage, 0)
	//Create the default mux
	mux := http.NewServeMux()
	mux.HandleFunc("/", viewDataHandler)
	mux.HandleFunc("/post-data", postDataHandler)

	//Create the server.
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", Port),
		Handler: mux,
	}
	log.Printf("Listening on port %d for http requests...\n", Port)
	s.ListenAndServe()
}

func postDataHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		data := []byte("This url is only for sending new sensor data. Use request method POST instead of GET please")
		w.Write(data)
	case "POST":
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		defer req.Body.Close()
		sensorDataPackage := NewSensorDataPackage()
		err = json.Unmarshal(body, &sensorDataPackage)
		if err != nil {
			panic(err)
		}
		sensorData = append(sensorData, sensorDataPackage)
		log.Printf("recieved data:%v", sensorDataPackage)
	default:
		fmt.Fprintf(w, "Only GET and POST methods are supported for this url.")

	}
}

func viewDataHandler(w http.ResponseWriter, req *http.Request) {
	//Todo display existing sensor data in html form
	//	html := `
	//<!DOCTYPE html>
	//<html lang="de">
	//<head>
	//    <meta charset="UTF-8"/>
	//    <title>Sensor Data</title>
	//</head>
	//<body>
	//<h1>hello</h1>
	//</body>
	//</html>`
	//	data := []byte(html)
	//	w.Write(data)
	http.ServeFile(w, req, "index.html")
}
