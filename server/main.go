package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/corrreia/chatroom/proto"
	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedConnectionServiceServer
	proto.UnimplementedDataServiceServer
	proto.UnimplementedBroadcastServiceServer
	proto.UnimplementedAdminServiceServer
}

type Server_State struct {
	Users map[string]*proto.User

	logFile string
	
	password string
	maxClients int
	connectedClients int
}

var server_state Server_State

func (s *Server) Connect(ctx context.Context, req *proto.ConnectRequest) (*proto.ConnectResponse, error) {
	
	//get user ip
	ip := ctx.Value("ip").(string)

	log.Println("Connection request received from ", ip)
	
	// check if server is full
	if server_state.connectedClients >= server_state.maxClients {
		return &proto.ConnectResponse{
			Status: proto.ConnectResponse_SERVER_FULL,
			Message: "capacity:" + strconv.Itoa(server_state.maxClients),
		}, nil
	}

	// check if user already exists
	if _, ok := server_state.Users[req.Username]; ok {
		return &proto.ConnectResponse{
			Status: proto.ConnectResponse_ALREADY_CONNECTED,
		}, nil
	}

	// check if server password is correct
	if server_state.password != req.ServerPassword {
		return &proto.ConnectResponse{
			Status: proto.ConnectResponse_INVALID_CREDENTIALS,
		}, nil
	}

	// at this point everything is ok, check if user is already registered
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

	// if password is correct, generate token and add user to server
	if server_state.Users[req.Username].PasswordHash == req.PasswordHash {
		// generate token
		token := generateToken()

		server_state.Users[req.Username].Token = token
		server_state.Users[req.Username].Connected = true
		server_state.Users[req.Username].LastKnownIp = ip
		server_state.Users[req.Username].LastSeen = time.Now().Unix()
		server_state.connectedClients++

		//broadcast user connected
		
		return &proto.ConnectResponse{
			Status: proto.ConnectResponse_OK,
			Token: token,
		}, nil
	}

	storeUsers()

	return &proto.ConnectResponse{
		Status: proto.ConnectResponse_INVALID_CREDENTIALS,
	}, nil
}

func (s *Server) Disconnect(ctx context.Context, req *proto.DisconnectRequest) (*proto.DisconnectResponse, error) {
	//get user ip
	ip := ctx.Value("ip").(string)

	log.Println("Disconnection request received from ", ip)

	//get user by token
	user := getUserByToken(req.Token)
	if user == nil {
		return &proto.DisconnectResponse{
			Status: proto.DisconnectResponse_INVALID_TOKEN,
		}, nil
	}

	//disconnect user
	user.Connected = false
	user.Token = ""
	server_state.connectedClients--

	//broadcast user disconnected


	storeUsers()

	return &proto.DisconnectResponse{
		Status: proto.DisconnectResponse_OK,
	}, nil
}

func (s *Server) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	//get user ip
	ip := ctx.Value("ip").(string)

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
		Role: proto.User_USER,
	}

	//save users to file
	storeUsers()

	return &proto.RegisterResponse{
		Status: proto.RegisterResponse_OK,
	}, nil
}

func (s *Server) SendMessage(ctx context.Context, req *proto.SendMessageRequest) (*proto.SendMessageResponse, error) {
	//get user ip
	ip := ctx.Value("ip").(string)

	log.Println("Message received from ", ip)

	//get user by token
	user := getUserByToken(req.Token)
	if user == nil {
		return &proto.SendMessageResponse{
			Status: proto.SendMessageResponse_INVALID_TOKEN,
		}, nil
	}

	//check if user is connected
	if !user.Connected {
		return &proto.SendMessageResponse{
			Status: proto.SendMessageResponse_NOT_CONNECTED,
		}, nil
	}

	// if token is valid, send message to all users
	if user.Token == req.Token {

		//TODO: broadcast message

		log.Println("[" + time.Now().Format("2/1/2006 15:04:05") + "] " + user.Username + ": " + req.Message)

		return &proto.SendMessageResponse{
			Status: proto.SendMessageResponse_OK,
		}, nil

	}

	return &proto.SendMessageResponse{
		Status: proto.SendMessageResponse_INVALID_TOKEN,
	}, nil
}

func (s *Server) SendCommand(ctx context.Context, req *proto.SendCommandRequest) (*proto.SendCommandResponse, error) {
	//get user ip
	ip := ctx.Value("ip").(string)

	log.Println("Command received from ", ip)

	//TODO: implement command system

	return &proto.SendCommandResponse{
		Status: proto.SendCommandResponse_INVALID_COMMAND,
	}, nil
}

// flags: -flag <default value>
// -port 8421
// -password ""
// -max_clients 10
// -debug false
// -log_file "server.log"

func main() {
	// parse flags
	port := flag.String("port", "8421", "port to listen on")
	password := flag.String("password", "", "password to connect")
	maxClients := flag.Int("max_clients", 10, "maximum number of clients")
	logFile := flag.String("log_file", "", "log file")
	flag.Parse()

	// set up logging
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// create server_state
	server_state = Server_State{
		Users: make(map[string]*proto.User),
		logFile: *logFile,
		password: *password,
		maxClients: *maxClients,
	}

	//if there is file with users, load them
	if _, err := os.Stat("users-data.json"); err == nil {
		loadUsers()
		log.Println("Loaded users from file")
	} else {
		log.Println("No users file found")
	}

	// start server
	log.Println("Starting server...")
	lis, err := net.Listen("tcp", ("localhost:" + *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create grpc server
	srv := grpc.NewServer()
	proto.RegisterConnectionServiceServer(srv, &Server{})
	proto.RegisterDataServiceServer(srv, &Server{})
	proto.RegisterBroadcastServiceServer(srv, &Server{})
	proto.RegisterAdminServiceServer(srv, &Server{})

	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
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

	log.Println("Users loaded")
}

func getUserByToken(token string) *proto.User {
	for _, user := range server_state.Users {
		if user.Token == token {
			return user
		}
	}
	return nil
}

func generateToken() string {
	str := fmt.Sprintf("%x ", md5.Sum([]byte(time.Now().String())))
	return str
}
	