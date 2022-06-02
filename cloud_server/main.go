package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
	"vs_praktikum_BreiterSchandl_Di2x/cloud_server/proto"
)

const (
	Port int = 8080
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

type SensorDataCollection struct {
	SensorData []SensorDataPackage
	Mutex      sync.RWMutex
}

func NewSensorDataCollection() SensorDataCollection {
	return SensorDataCollection{
		SensorData: make([]SensorDataPackage, 0),
		Mutex:      sync.RWMutex{},
	}
}

var sensorDataCollection SensorDataCollection

var (
	TestLoggerP3 *log.Logger
)

var startTime time.Time

func Uptime() time.Duration {
	return time.Since(startTime)
}

func init() {
	startTime = time.Now()
	logfileP3, err := os.OpenFile("/logs/P3RttLog.log", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("log directory could not be created. Try creating it manually: %v\n", err)
	}
	TestLoggerP3 = log.New(logfileP3, "", 0)

}
func main() {
	sensorDataCollection = NewSensorDataCollection()

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
	//TODO fix
	//fs := http.FileServer(http.Dir("./static"))
	//mux.PathPrefix("/").Handler(http.StripPrefix("/static/", fs))
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
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer req.Body.Close()
		sensorDataPackage := NewSensorDataPackage()
		err = json.Unmarshal(body, &sensorDataPackage)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sensorDataCollection.Mutex.Lock()
		sensorDataCollection.SensorData = append(sensorDataCollection.SensorData, sensorDataPackage)
		sensorDataCollection.Mutex.Unlock()
		timeBeforeCreateRPC := time.Now()
		CreateSDPinDB(&sensorDataPackage)
		rttCreateRPC := time.Since(timeBeforeCreateRPC)
		TestLoggerP3.Printf("rtt rpc: %v", rttCreateRPC)
		w.WriteHeader(http.StatusOK)
		log.Printf("recieved data:%v", sensorDataPackage)
	default:
		fmt.Fprintf(w, "Only GET and POST methods are supported for this url.")

	}
}

func viewDataHandler(w http.ResponseWriter, req *http.Request) {
	//display existing sensor data in html form
	RenderTemplate(w, req)
}

//RenderTemplate renders hardcoded template and data
func RenderTemplate(w http.ResponseWriter, req *http.Request) {
	pathToTemplate := fmt.Sprintf("./templates/%s", "index.tmpl")
	t, err := template.New("index.tmpl").ParseFiles(pathToTemplate)
	if err != nil {
		log.Println(err)

	}

	buf := new(bytes.Buffer)
	sensorDataCollection.Mutex.RLock()
	err = t.Execute(buf, sensorDataCollection.SensorData)
	sensorDataCollection.Mutex.RUnlock()
	if err != nil {
		log.Println(err)
	}
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Println("error writing template to browser", err)
	}

}

//CreateSDPinDB sends the given sensorDataPackage to the database server via RPC call
func CreateSDPinDB(sdp *SensorDataPackage) {
	ips, err := net.LookupIP("database")
	if err != nil {
		log.Println("Database server could not be found.")
		return
	}
	addr := fmt.Sprintf("%s:40401", ips[0])
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewDatabaseServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	protoBTimestamp := timestamppb.New(sdp.Timestamp)
	r, err := c.Create(ctx, &proto.SensorDataPackage{Timestamp: protoBTimestamp, Data: sdp.Data, SensorCount: sdp.SensorCount})
	if err != nil {
		log.Printf("could not create: %v", err)
	}
	log.Printf("creation response: %v", r.GetSuccess())

	res, err := c.Read(ctx, &proto.IDSensorDataPackageTimestamp{Timestamp: protoBTimestamp})

	if err != nil {
		log.Printf("could not create: %v", err)
	}

	log.Printf("read response: %v,%v,%v", res.GetTimestamp().AsTime(), res.GetSensorCount(), res.GetData())
}
