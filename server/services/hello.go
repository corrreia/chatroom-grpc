package services

import (
	"io/ioutil"
	"log"
	"net"
	"strings"
)

func StartHelloServer(socket net.PacketConn, caPath string) (error) {
	log.Println("Starting Hello Server")

	// read ca certificate
	cacert, err := ioutil.ReadFile(caPath)
	if err!= nil { return err }

	log.Println("CA certificate loaded")

	for {
		buffer := make([]byte, 5) // HELLO
		n, addr, err := socket.ReadFrom(buffer)
		if err!= nil { return err }

		log.Printf("Client connected from %s\n", addr.String())

		if strings.Contains(string(buffer[:n]), "HELLO") {
            log.Println("Sending CA certificate")
            _, err = socket.WriteTo([]byte(string(cacert)), addr)
            if err!= nil { 
				log.Println(err) //this should not return an error but if it does, it should not stop the server
			}
		}
	}
}