package main

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
	"time"
	"vs_praktikum_BreiterSchandl_Di2x/database/proto"
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

type dbServer struct {
	proto.UnimplementedDatabaseServiceServer
}

func main() {
	sensorDataCollection = NewSensorDataCollection()
	listener, err := net.Listen("tcp", ":40401")
	if err != nil {
		log.Println(err)
	}
	srv := grpc.NewServer()
	proto.RegisterDatabaseServiceServer(srv, &dbServer{})
	log.Printf("server listening at %v", listener.Addr())
	if err := srv.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *dbServer) Create(ctx context.Context, sdp *proto.SensorDataPackage) (*proto.Response, error) {
	//
	sensorDataPackage := NewSensorDataPackage()
	sensorDataPackage.Timestamp = sdp.GetTimestamp().AsTime()
	sensorDataPackage.Data = sdp.GetData()
	sensorDataPackage.SensorCount = sdp.GetSensorCount()
	sensorDataCollection.Mutex.Lock()
	sensorDataCollection.SensorData = append(sensorDataCollection.SensorData, sensorDataPackage)
	sensorDataCollection.Mutex.Unlock()
	return &proto.Response{Success: true}, nil
}

func (s *dbServer) Read(ctx context.Context, sdpTS *proto.IDSensorDataPackageTimestamp) (*proto.SensorDataPackage, error) {
	//https://pkg.go.dev/google.golang.org/protobuf/types/known/timestamppb
	//ts := timestamppb.New(time.Now())
	ts := sdpTS.GetTimestamp().AsTime()
	sdp := NewSensorDataPackage()
	sensorDataCollection.Mutex.RLock()
	for _, sDataP := range sensorDataCollection.SensorData {
		if sDataP.Timestamp == ts {
			sdp = sDataP
			sensorDataCollection.Mutex.RUnlock()
			return &proto.SensorDataPackage{Timestamp: sdpTS.GetTimestamp(), Data: sdp.Data, SensorCount: sdp.SensorCount}, nil
		}
	}
	sensorDataCollection.Mutex.RUnlock()

	return &proto.SensorDataPackage{}, errors.New("the requested sensordata package was not found")
	//t := ts.AsTime() //convert back to gotime

}

func (s *dbServer) Update(ctx context.Context, sdp *proto.SensorDataPackage) (*proto.Response, error) {
	ts := sdp.GetTimestamp().AsTime()
	sensorDataCollection.Mutex.Lock()
	for i, sDataP := range sensorDataCollection.SensorData {
		if sDataP.Timestamp == ts {
			//sDataP.Data = sdp.GetData()
			//sDataP.SensorCount = sdp.GetSensorCount()
			sensorDataCollection.SensorData[i].Data = sdp.GetData()
			sensorDataCollection.SensorData[i].SensorCount = sdp.GetSensorCount()
			sensorDataCollection.Mutex.Unlock()
			return &proto.Response{Success: true}, nil
		}
	}
	sensorDataCollection.Mutex.Unlock()
	return &proto.Response{Success: false}, errors.New("the requested sensordata package was not found")
}

//Delete deletes the sensordatapackage with the given timestamp and does not secure order of sensordatapackages
func (s *dbServer) Delete(ctx context.Context, sdpTS *proto.IDSensorDataPackageTimestamp) (*proto.Response, error) {
	ts := sdpTS.GetTimestamp().AsTime()
	sensorDataCollection.Mutex.Lock()
	for i, sDataP := range sensorDataCollection.SensorData {
		if sDataP.Timestamp == ts {
			sensorDataCollection.SensorData[i] = sensorDataCollection.SensorData[len(sensorDataCollection.SensorData)-1]
			sensorDataCollection.SensorData = sensorDataCollection.SensorData[:len(sensorDataCollection.SensorData)-1]
			sensorDataCollection.Mutex.Unlock()
			return &proto.Response{Success: true}, nil
		}
	}
	sensorDataCollection.Mutex.Unlock()
	return &proto.Response{Success: false}, errors.New("the requested sensordata package was not found")
}
