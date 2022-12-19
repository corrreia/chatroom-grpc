package services

import (
	"log"

	"google.golang.org/grpc"

	pb "github.com/corrreia/chatroom-grpc/proto"
	"github.com/corrreia/chatroom-grpc/server/types"
)

type communicationServer struct {
	pb.UnimplementedAnnouncementServiceServer
	pb.UnimplementedChatServiceServer
	pb.UnimplementedCommandServiceServer
}

var communicationState *types.ServerState = nil

func StartCommunicationServer(s *grpc.Server, state *types.ServerState) {
	log.Printf("Starting Communication server")

	pb.RegisterAnnouncementServiceServer(s, &communicationServer{})
	pb.RegisterChatServiceServer(s, &communicationServer{})
	pb.RegisterCommandServiceServer(s, &communicationServer{})
}

func (s *communicationServer) SendAnnouncement(ctx context.Context, req *pb.AnnouncementRequest) (*pb.AnnouncementResponse, error) {
