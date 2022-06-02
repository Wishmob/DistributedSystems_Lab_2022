package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"strings"
	"testing"
	"time"
	"vs_praktikum_BreiterSchandl_Di2x/cloud_server/proto"
)

//go test -v ./...
//in cloud_server directory

func TestCRUD(t *testing.T) {
	//********************
	//START DATABASE SERVER
	//********************
	composeFilePaths := []string{"../docker-compose-test.yml"}
	identifier := strings.ToLower(uuid.New().String())
	compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)
	execError := compose.
		WithCommand([]string{"up", "-d", "--build"}).
		WithEnv(map[string]string{
			//"key1": "value1",
		}).
		Invoke()
	err := execError.Error
	if err != nil {
		fmt.Errorf("could not run compose file: %v - %v", composeFilePaths, err)
	}
	addr := "localhost:40401"
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Errorf("did not connect: %v", err)
	}
	defer conn.Close()

	//********************
	//TESTS
	//********************
	c := proto.NewDatabaseServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ts := time.Now()
	tsPB := timestamppb.New(ts)
	testSdp := NewSensorDataPackage()
	testSdp.Timestamp = ts
	testSdp.Data["123"] = "456"
	testSdp.SensorCount = 1

	//Test Create
	createResp, err := c.Create(ctx, &proto.SensorDataPackage{Timestamp: tsPB, Data: testSdp.Data, SensorCount: testSdp.SensorCount})
	if err != nil {
		fmt.Errorf("could not create: %v", err)
	}
	log.Printf("Response to Create RPC call: %v", createResp.GetSuccess())

	//Test Read
	readResp, err := c.Read(ctx, &proto.IDSensorDataPackageTimestamp{Timestamp: tsPB})

	if err != nil {
		t.Errorf("could not read with valid call: %v", err)
	}
	log.Printf("Response to Valid Read RPC call: %v,%v,%v\n", readResp.GetTimestamp().AsTime(), readResp.GetSensorCount(), readResp.GetData())

	readResp, err = c.Read(ctx, &proto.IDSensorDataPackageTimestamp{})

	expectedErrorMsg := "rpc error: code = Unknown desc = the requested sensordata package was not found"
	if err != nil && err.Error() != expectedErrorMsg {
		t.Errorf("could not read with invalid call (as expected) but got unexpected error\nExpected: %s\nGot: %v\n", expectedErrorMsg, err)
	}
	log.Printf("Response to Invalid Read RPC call: %v,%v,%v with expected error: %v", readResp.GetTimestamp(), readResp.GetSensorCount(), readResp.GetData(), err)

	//Test Update
	//Test Delete

	//********************
	//STOP DATABASE SERVER
	//********************
	execError = compose.Down()
	err = execError.Error
	if err != nil {
		fmt.Errorf("could not run compose file: %v - %v", composeFilePaths, err)
	}

}
