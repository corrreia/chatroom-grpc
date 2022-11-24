package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/corrreia/chatroom-grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type helloServer struct {
	pb.UnimplementedHelloServiceServer
}

var cacert string = ""

func HelloServer(port int, CAcert string){
	log.Println("Starting Hello Server in port ", port)

	//start unencrypted server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterHelloServiceServer(s, &helloServer{})

	cacert = CAcert

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *helloServer) Hello(ctx context.Context, in *pb.HelloClient) (*pb.HelloServer, error) {
	//get ip
	p, _ := peer.FromContext(ctx)
	ip, _, _ := net.SplitHostPort(p.Addr.String())
	
	log.Println("CA certificate request from: ", ip)

	return &pb.HelloServer{CA: cacert}, nil
}