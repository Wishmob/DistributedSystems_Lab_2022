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
		w.WriteHeader(http.StatusOK)
		timeBeforeCreateRPC := time.Now()
		CreateSDPinDB(&sensorDataPackage)
		rttCreateRPC := time.Since(timeBeforeCreateRPC)
		TestLoggerP3.Printf("rtt rpc: %v", rttCreateRPC)
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
	ipsDB1, err := net.LookupIP("database1")
	if err != nil {
		log.Println("Database 1 server could not be found.")
		return
	}
	ipsDB2, err := net.LookupIP("database2")
	if err != nil {
		log.Println("Database 2 server could not be found.")
		return
	}

	addrDB1 := fmt.Sprintf("%s:40401", ipsDB1[0])
	addrDB2 := fmt.Sprintf("%s:40401", ipsDB2[0])

	connDB1, err := grpc.Dial(addrDB1, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
		return
	}
	defer connDB1.Close()
	clientDB1 := proto.NewDatabaseServiceClient(connDB1)

	connDB2, err := grpc.Dial(addrDB2, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
		return
	}
	defer connDB2.Close()
	clientDB2 := proto.NewDatabaseServiceClient(connDB2)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	pingRespDB1, err := clientDB1.Ping(ctx, &proto.Request{Id: 5})
	if err != nil {
		log.Printf("Error pinging DB1: %v", err)
		return
	}
	pingRespDB2, err := clientDB2.Ping(ctx, &proto.Request{Id: 5})
	if err != nil {
		log.Printf("Error pinging DB2: %v", err)
		return
	}

	if pingRespDB1.GetSuccess() && pingRespDB2.GetSuccess() {
		protoBTimestamp := timestamppb.New(sdp.Timestamp)
		createRespDB1, err := clientDB1.Create(ctx, &proto.SensorDataPackage{Timestamp: protoBTimestamp, Data: sdp.Data, SensorCount: sdp.SensorCount})
		if err != nil {
			log.Printf("could not create in db1: %v", err)
		}
		createRespDB2, err := clientDB2.Create(ctx, &proto.SensorDataPackage{Timestamp: protoBTimestamp, Data: sdp.Data, SensorCount: sdp.SensorCount})
		if err != nil {
			log.Printf("could not create in db2: %v", err)
		}
		log.Printf("successfully saved package with timestamp %v to databases via RPC. Got responses: %v and %v", sdp.Timestamp, createRespDB1.GetSuccess(), createRespDB2.GetSuccess())
	} else {
		log.Println("Not both databases were ready to receive data. Saving to databases aborted...")
		return
	}

	// Test if package has actually been saved successfully to database
	//res, err := clientDB1.Read(ctx, &proto.IDSensorDataPackageTimestamp{Timestamp: protoBTimestamp})
	//
	//if err != nil {
	//	log.Printf("Did not find recently created package in database: %v", err)
	//}
	//
	//log.Printf("read response: %v,%v,%v", res.GetTimestamp().AsTime(), res.GetSensorCount(), res.GetData())
}
