package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
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

	getCA(*addr, utils.Path("/"))

	creds, err := credentials.NewClientTLSFromFile(utils.Path("/ca_cert_client.pem"), "chat.dev.tomascorreia.net")
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Make a echo client and send an RPC.
	rgc := pb.NewEchoClient(conn)
	callUnaryEcho(rgc, "hello world")
}

func getCA(addr string, path string) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
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
	caFile, err := os.Create(path+"/ca_cert_client.pem")
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