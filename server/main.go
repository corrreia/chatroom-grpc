package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/corrreia/chatroom-grpc/server/interceptors"
	"github.com/corrreia/chatroom-grpc/server/services"
	"github.com/corrreia/chatroom-grpc/utils"
	"github.com/corrreia/chatroom-grpc/server/types"


	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)



func main() {
	// parse flags
	port := flag.Int("port", 8421, "port to listen on")
	password := flag.String("password", "", "password to connect")
	maxClients := flag.Int("max_clients", 10, "maximum number of clients")
	logFile := flag.String("log_file", "", "log file")
	flag.Parse()

	// create server state
	state := types.NewServerState()

	state.SetServerPassword(*password)
	state.SetMaxClients(*maxClients)
	state.SetCaPath(utils.Path("ca_cert_client.pem"))
	state.SetCertPath(utils.Path("server_cert.pem"))
	state.SetKeyPath(utils.Path("server_key.pem"))

	// set up logging
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Println("Random test token: ", utils.GenerateToken())

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create tls based credential.
	creds, err := credentials.NewServerTLSFromFile(state.GetCertPath(), state.GetKeyPath())
	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}
	log.Println("Server credentials loaded")

	helloS := grpc.NewServer(grpc.UnaryInterceptor(interceptors.UnaryLogInterceptor)) //this is a hello server so there is no encryption
	mainS := grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(interceptors.UnaryLogInterceptor))

	//services.StartHelloServer(helloS, state.GetCaPath()) // hello service to get ca certificate  //! need to find a way to get the ca certificate from the server to the client
	services.StartAuthServer(mainS, state) // auth service to authenticate clients and get token
	services.StartCommunicationServer(mainS, state)  // communication service to send messages and commands


	if err := helloS.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	if err := mainS.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	
	//* AFTER THIS LINE, THE SERVER IS RUNNING AND LISTENING FOR CONNECTIONS
}