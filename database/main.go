package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"time"
	"vs_praktikum_BreiterSchandl_Di2x/database/proto"
)

type dbServer struct {
	proto.UnimplementedDatabaseServiceServer
}

func main() {
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
	return &proto.Response{Success: true}, nil
}

func (s *dbServer) Read(ctx context.Context, sdp *proto.IDSensorDataPackageTimestamp) (*proto.SensorDataPackage, error) {
	//https://pkg.go.dev/google.golang.org/protobuf/types/known/timestamppb
	ts := timestamppb.New(time.Now())
	//t := ts.AsTime() //convert back to gotime
	return &proto.SensorDataPackage{Timestamp: ts}, nil
}

func (s *dbServer) Update(ctx context.Context, sdp *proto.SensorDataPackage) (*proto.Response, error) {
	//
	return &proto.Response{Success: true}, nil
}

func (s *dbServer) Delete(ctx context.Context, sdp *proto.IDSensorDataPackageTimestamp) (*proto.Response, error) {
	//
	return &proto.Response{Success: true}, nil
}
