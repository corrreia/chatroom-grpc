package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/corrreia/chatroom-grpc/cert"
	"github.com/corrreia/chatroom-grpc/server/grpc"
	"github.com/corrreia/chatroom-grpc/server/console"
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
	certPath := flag.String("cert_path", "./cert/", "path to where to store/get/generate certificates")
	certDomain := flag.String("cert_domain", "localhost", "domain to generate certificates for if they don't exist")
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

	keyPem, certPem, err := handleCerts(*certPath, *certDomain)
	if err != nil {
		log.Fatal(err)
	}

	//create a thread and start server
	go grpc.StartServer(*port, *password, *maxClients+1, keyPem, certPem)  //maxClients+1 because the server cosoles counts as a client

	//start console
	console.StartConsole()
}

func handleCerts(certPath string, certDomain string) (string, string, error) {
	log.Println("Handling certificates...")
	if _, err := os.Stat(certPath + "ca_key.pem"); errors.Is(err, os.ErrNotExist) {
		log.Println("Generating CA certificates...")
		err := cert.GenerateCACertKey(certPath)
		if err != nil {
			return "", "", err
		}

		log.Println("Generating server certificates...")
		err = cert.GenerateServerCertKey(certDomain, certPath, certPath)
		if err != nil {
			return "", "", err
		}
	} else{
		log.Println("Certificates found")
	}
	//read files and return server key and cert
	keyPem, err := os.ReadFile(certPath + "server_key.pem")
	if err != nil {
		return "", "", err
	}
	certPem, err := os.ReadFile(certPath + "server_cert.pem")
	if err != nil {
		return "", "", err
	}
	return string(keyPem), string(certPem), nil
}