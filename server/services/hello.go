package services

import (
	"context"
	"io/ioutil"
	"log"

	pb "github.com/corrreia/chatroom-grpc/proto"
	"google.golang.org/grpc"
)

type helloServer struct {
	pb.UnimplementedHelloServiceServer
}

var caCert string = ""

func StartHelloServer(s *grpc.Server, caPath string) {
	log.Println("Starting Hello Server")

	// read ca certificate
	cacert, err := ioutil.ReadFile(caPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("CA certificate loaded")

	caCert = string(cacert)

	pb.RegisterHelloServiceServer(s, &helloServer{})
}

func (s *helloServer) Hello(ctx context.Context, in *pb.HelloClient) (*pb.HelloServer, error) {
	return &pb.HelloServer{CA: caCert}, nil
}