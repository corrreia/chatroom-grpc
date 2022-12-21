package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/corrreia/chatroom-grpc/server/interceptors"
	"github.com/corrreia/chatroom-grpc/server/services"
	"github.com/corrreia/chatroom-grpc/server/types"
	"github.com/corrreia/chatroom-grpc/utils"

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

	certPath := "./certs"

	// create server state
	state := types.NewServerState()

	state.SetServerPassword(*password)
	state.SetMaxClients(*maxClients)
	state.SetPort(*port)
	state.SetCaPath(filepath.Join(certPath, "ca_cert.pem"))
	state.SetCertPath(filepath.Join(certPath, "server_cert.pem"))
	state.SetKeyPath(filepath.Join(certPath, "server_key.pem"))

	// set up logging
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Println("Random test token:", utils.GenerateToken())

	// open sockets
	tcpSock, udpSock, err := openSockets(state.GetPort())
	if err!= nil { 
		log.Fatal(err)
	}

	// Create tls based credential.
	creds, err := credentials.NewServerTLSFromFile(state.GetCertPath(), state.GetKeyPath())
	if err != nil {
		log.Fatal(err)
	}
	
	log.Println("Server credentials loaded")

	//create grpc servers
	authS := grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(interceptors.UnaryLogInterceptor))
	mainS := grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(interceptors.UnaryLogInterceptor)) //!TODO: add auth interceptor

	services.StartAuthServer(authS, state) // auth service to authenticate clients and get token
	services.StartCommunicationServer(mainS, state)  // communication service to send messages and commands
	
	errCh := make(chan error)

	go func() {
        err = services.StartHelloServer(udpSock, state.GetCaPath())
		if err!= nil {
            errCh <- err
        }
	}()

	go func () {  //start hello server in a goroutine and send errors to channel
		err = authS.Serve(tcpSock)
		if err!= nil {
            errCh <- err
        }
	}()

	go func () { //start main server in a goroutine and send errors to channel
		err = mainS.Serve(tcpSock)
		if err!= nil {
            errCh <- err
        }		
	}()

	for { //wait for errors and exit if there is one
		if err := <-errCh; err != nil {
			log.Fatal(err)
		}
	}
}

func openSockets(port int) (net.Listener, net.PacketConn, error) {
	TCPsock, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
        return nil, nil, err
    }

	UDPsock, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
	if err!= nil {
		return nil, nil, err
	}

	return TCPsock, UDPsock, nil
}