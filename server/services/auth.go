package services

import (
	"context"
	"log"

	"google.golang.org/grpc"

	pb "github.com/corrreia/chatroom-grpc/proto"
	"github.com/corrreia/chatroom-grpc/server/types"
)

type authServer struct {
	pb.UnimplementedAuthServiceServer
}

var State *types.ServerState = nil

func StartAuthServer(s *grpc.Server, state *types.ServerState) {
	log.Printf("Starting Auth server")

	State = state
	pb.RegisterAuthServiceServer(s, &authServer{})
}

func (s *authServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Login request from %v", req.Username)

	// check if user exists
	user := State.GetUserByUsername(req.Username)
	//id is User{} if user does not exist
	if user.GetId() == "" {
		return &pb.LoginResponse{Status: pb.LoginResponse_INVALID_CREDENTIALS}, nil
	}

	// check if user is banned
	if user.IsBanned() {
		log.Printf("User %v is banned", req.Username)
		return &pb.LoginResponse{Status: pb.LoginResponse_USER_BANNED}, nil
	}

	// check if user is already connected
	if user.IsConnected() {
		log.Printf("User %v is already connected", req.Username)
		return &pb.LoginResponse{Status: pb.LoginResponse_ALREADY_LOGGED_IN}, nil
	}

	// check if password is correct
	if !user.CheckPassword(req.Password) {
		log.Printf("Invalid password for user %v", req.Username)
		return &pb.LoginResponse{Status: pb.LoginResponse_INVALID_CREDENTIALS}, nil
	}

	// login user
	user.SetConnected(true)
	user.RegenerateToken()
	log.Printf("User %v logged in", req.Username)

	return &pb.LoginResponse{Status: pb.LoginResponse_SUCCESS, Token: user.GetToken()}, nil
}

