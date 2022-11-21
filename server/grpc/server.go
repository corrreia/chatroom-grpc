package grpc

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/corrreia/chatroom-grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Server struct {
	proto.UnimplementedConnectionServiceServer
	proto.UnimplementedDataServiceServer
	proto.UnimplementedBroadcastServiceServer
	proto.UnimplementedAdminServiceServer
}

type Server_State struct {
	Users map[string]*proto.User
	
	password string
	maxClients int
	connectedClients int

	subscribers map[string]*proto.BroadcastService_SubscribeServer
}

var server_state Server_State

func (s *Server) Connect(ctx context.Context, req *proto.ConnectRequest) (*proto.ConnectResponse, error) {
	
	p, _ := peer.FromContext(ctx)
  	ip := strings.Split(p.Addr.String(), ":")[0]

	log.Println("Connection request received from ", ip)
	
	// check if server is full
	if server_state.connectedClients >= server_state.maxClients {
		return &proto.ConnectResponse{
			Status: proto.ConnectResponse_SERVER_FULL,
		}, nil
	}

	// check if user is already connected
	if server_state.Users[req.Username].Connected {
		return &proto.ConnectResponse{
			Status: proto.ConnectResponse_ALREADY_CONNECTED,
		}, nil
	}

	// check if server password is correct
	if server_state.password != req.ServerPassword {
		return &proto.ConnectResponse{
			Status: proto.ConnectResponse_INVALID_SERVER_PASSWORD,
		}, nil
	}

	//check if user is registered
	if _, ok := server_state.Users[req.Username]; !ok {
		return &proto.ConnectResponse{
			Status: proto.ConnectResponse_USER_UNKNOWN,
		}, nil
	}

	//check if useris Banned
	if server_state.Users[req.Username].Banned {
		return &proto.ConnectResponse{
			Status: proto.ConnectResponse_USER_BANNED,
		}, nil
	}

	if server_state.Users[req.Username].PasswordHash != req.PasswordHash {
		return &proto.ConnectResponse{
			Status: proto.ConnectResponse_INVALID_CREDENTIALS,
		}, nil
	}
	token := generateToken()

	server_state.Users[req.Username].Token = token
	server_state.Users[req.Username].Connected = true
	server_state.Users[req.Username].LastKnownIp = ip
	server_state.Users[req.Username].LastSeen = time.Now().Unix()
	server_state.connectedClients++

	//broadcast user connected
	broadcastUserConnected(server_state.Users[req.Username])


	storeUsers()
		
	return &proto.ConnectResponse{
		Status: proto.ConnectResponse_SUCCESS,
		Token: token,
	}, nil
}

func (s *Server) Disconnect(ctx context.Context, req *proto.DisconnectRequest) (*proto.DisconnectResponse, error) {
	//get user ip
	p, _ := peer.FromContext(ctx)
	ip := strings.Split(p.Addr.String(), ":")[0]

	log.Println("Disconnection request received from ", ip)

	user, err := validateUser(ctx)
	if err != nil {
		return &proto.DisconnectResponse{
			Status: proto.DisconnectResponse_INVALID_TOKEN,
		}, nil
	}

	//disconnect user
	user.Connected = false
	user.Token = ""
	user.TimeConnected += time.Now().Unix() - user.LastSeen
	server_state.connectedClients--

	//broadcast user disconnected
	broadcastUserDisconnected(user)

	storeUsers()

	return &proto.DisconnectResponse{
		Status: proto.DisconnectResponse_SUCCESS,
	}, nil
}

func (s *Server) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	//get user ip
	p, _ := peer.FromContext(ctx)
	ip := strings.Split(p.Addr.String(), ":")[0]

	log.Println("Registration request received from ", ip)

	//check if user is already registered
	if _, ok := server_state.Users[req.Username]; ok {
		return &proto.RegisterResponse{
			Status: proto.RegisterResponse_USERNAME_TAKEN,
		}, nil

	}

	//check if user is banned
	if server_state.Users[req.Username].Banned {
		return &proto.RegisterResponse{
			Status: proto.RegisterResponse_USER_BANNED,
		}, nil
	}

	//register user
	server_state.Users[req.Username] = &proto.User{
		Username: req.Username,
		PasswordHash: req.PasswordHash,
		Connected: false,
		Banned: false,
		LastKnownIp: ip,
		LastSeen: time.Now().Unix(),
		TimeConnected: 0,
		Role: proto.User_USER,
	}

	//save users to file
	storeUsers()

	log.Println("User ", req.Username, " registered (", ip, ")")
	return &proto.RegisterResponse{
		Status: proto.RegisterResponse_SUCCESS,
	}, nil
}

func (s *Server) SendMessage(ctx context.Context, req *proto.SendMessageRequest) (*proto.SendMessageResponse, error) {
	//get user ip
	p, _ := peer.FromContext(ctx)
	ip := strings.Split(p.Addr.String(), ":")[0]
	log.Println("Message received from ", ip)

	user, err := validateUser(ctx)
	if err != nil {
		return &proto.SendMessageResponse{
			Status: proto.SendMessageResponse_INVALID_TOKEN,
		}, nil
	}

	log.Println( user.Username, " : ", req.Message, " (", ip, ")")
	
	//broadcast message
	broadcastMessage(user, req.Message)

	return &proto.SendMessageResponse{
		Status: proto.SendMessageResponse_INVALID_TOKEN,
	}, nil
}

func (s *Server) SendCommand(ctx context.Context, req *proto.SendCommandRequest) (*proto.SendCommandResponse, error) {
	//get user ip
	p, _ := peer.FromContext(ctx)
	ip := strings.Split(p.Addr.String(), ":")[0]
	log.Println("Command received from ", ip)

	//TODO: implement command system

	return &proto.SendCommandResponse{
		Status: proto.SendCommandResponse_INVALID_COMMAND,
	}, nil
}

func (s *Server) Subscribe(in *proto.SubscribeRequest, srv proto.BroadcastService_SubscribeServer) error {
	//get user ip
	p, _ := peer.FromContext(srv.Context())
	ip := strings.Split(p.Addr.String(), ":")[0]

	log.Println("Subscription request received from ", ip)

	user, err := validateUser(srv.Context())
	if err != nil {
		//send error to client
		srv.Send(&proto.SubscribeResponse{
			Status: proto.SubscribeResponse_INVALID_TOKEN,
		})
		return err
	}

	//add user to subscribers map
	server_state.subscribers[user.Username] = &srv

	//dend success response
	srv.Send(&proto.SubscribeResponse{
		Status: proto.SubscribeResponse_SUCCESS,
	})

	return nil
}

func StartServer(port int, password string, maxClients int, keyPem string, certPem string) {
	
	server_state = Server_State{
		Users: make(map[string]*proto.User),
		password: password,
		maxClients: maxClients,
		connectedClients: 0,
	}
	
	//if there is file with users, load them
	if _, err := os.Stat("users-data.json"); err == nil {
		loadUsers()
		log.Println("Loaded users from file")
	} else {
		log.Println("No users file found")
	}
	
	log.Println("Starting server on port: ", port)

	cert, err := tls.LoadX509KeyPair(certPem, keyPem)
	if err != nil {
		log.Fatalf("failed to load key pair: %s", err)
	}

	opts := []grpc.ServerOption{
		// Enable TLS for all incoming connections.
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	}

	s := grpc.NewServer(opts...)

	proto.RegisterConnectionServiceServer(s, &Server{})
	proto.RegisterDataServiceServer(s, &Server{})
	proto.RegisterBroadcastServiceServer(s, &Server{})
	proto.RegisterAdminServiceServer(s, &Server{})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func broadcastUserConnected(user *proto.User) {
	for _, srv := range server_state.subscribers {
		(*srv).Send(&proto.SubscribeResponse{
			Type: proto.SubscribeResponse_LOGIN,
			Username: user.Username,
		})
	}
}

func broadcastUserDisconnected(user *proto.User) {
	for _, srv := range server_state.subscribers {
		(*srv).Send(&proto.SubscribeResponse{
			Type: proto.SubscribeResponse_LOGOUT,
			Username: user.Username,
		})
	}
}

func broadcastMessage(user *proto.User, message string) {
	for _, srv := range server_state.subscribers {
		(*srv).Send(&proto.SubscribeResponse{
			Type: proto.SubscribeResponse_MESSAGE,
			Message: message,
			Username: user.Username,
		})
	}
}

//function to store users in a file
func storeUsers() {
	//create a file
	log.Println("Storing users...")

	file, _ := json.MarshalIndent(server_state.Users, "", " ")
	_ = ioutil.WriteFile("users-data.json", file, 0644)

	log.Println("Users stored")
}

//function to load users from a file
func loadUsers() {
	//open file
	log.Println("Loading users...")

	file, _ := ioutil.ReadFile("users-data.json")
	_ = json.Unmarshal([]byte(file), &server_state.Users)

	for _, user := range server_state.Users {
		user.Connected = false
	}

	log.Println("Users loaded")
}

func validateUser(ctx context.Context) (*proto.User, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}
	token := strings.TrimPrefix(md["authorization"][0], "Bearer ")

	for _, user := range server_state.Users {
		if user.Token == token {
			return user, nil
		}
	}
	return nil, status.Errorf(codes.Unauthenticated, "invalid token")
}

func generateToken() string {
	str := fmt.Sprintf("%x ", md5.Sum([]byte(time.Now().String())))
	return str
}
