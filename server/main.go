package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/corrreia/chatroom-grpc/server/console"
	"github.com/corrreia/chatroom-grpc/server/grpc"
	"github.com/corrreia/chatroom-grpc/utils"
)

// flags: -flag <default value>
// -port 8421
// -password ""
// -max_clients 10
// -debug false
// -log_file "server.log"
// -cert_path "./cert/"
// -cert_domain "localhost"

func main() {
	// parse flags
	port := flag.Int("port", 8421, "port to listen on")
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

	// read ca certificate
	caCert, err := ioutil.ReadFile(utils.Path("/ca_cert.pem"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("CA certificate loaded")

	//start hello server
	go grpc.HelloServer(*port+1, string(caCert))

	//create a thread and start server
	go grpc.StartServer(*port, *password, *maxClients+1, utils.Path("/server_cert.pem"), utils.Path("/server_key.pem"))  //maxClients+1 because the server cosoles counts as a client

	//start console
	console.StartConsole()
}