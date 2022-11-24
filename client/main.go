/*
 *
 * Copyright 2018 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Binary client is an example client.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	pb "github.com/corrreia/chatroom-grpc/proto"
	"github.com/corrreia/chatroom-grpc/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var addr = flag.String("addr", "localhost:8421", "the address to connect to")

func callUnaryEcho(client pb.EchoClient, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.UnaryEcho(ctx, &pb.EchoRequest{Message: message})
	if err != nil {
		log.Fatalf("client.UnaryEcho(_) = _, %v: ", err)
	}
	fmt.Println("UnaryEcho: ", resp.Message)
}

func main() {
	flag.Parse()

	creds, err := credentials.NewClientTLSFromFile(utils.Path("/ca_cert.pem"), "chat.dev.tomascorreia.net")
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}

	// split host and port
	host, port, err := net.SplitHostPort(*addr)
	if err != nil {
		log.Fatalf("failed to split host and port: %v", err)
	}
	portInt , _ := strconv.Atoi(port)

	//hello port is port+1

	getCA(host, portInt, utils.Path("/"))

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Make a echo client and send an RPC.
	rgc := pb.NewEchoClient(conn)
	callUnaryEcho(rgc, "hello world")
}

func getCA(host string, port int, path string) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port+1), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Make a echo client and send an RPC.
	rgc := pb.NewHelloServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := rgc.Hello(ctx, &pb.HelloClient{})
	if err != nil {
		log.Fatalf("could not get CA certificate %v: ", err)
	}

	CACert := resp.CA

	//create file
	caFile, err := os.Create(path+"/ca.pem")
	if err != nil {
		log.Fatalf("could not create file: %v", err)
	}

	//write to file
	_, err = caFile.Write([]byte(CACert))
	if err != nil {
		log.Fatalf("could not write to file: %v", err)
	}
	caFile.Close()
}