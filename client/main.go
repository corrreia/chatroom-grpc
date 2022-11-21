package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"google.golang.org/grpc/credentials"
)

type model int

type tickMsg time.Time

func main() {
	// get server address from command line
	serverAddr := flag.String("server", "localhost:8421", "server address")
	serverPass := flag.String("password", "", "server password")
	certPath := flag.String("cert_path", "./cert/", "path to ca_cert.pem")
	flag.Parse()

	//read cacert in certPath+ca_cert.pem
	CAcert, err := os.ReadFile(*(certPath) + "ca_cert.pem")
	if err != nil {
		fmt.Println("Error reading ca_cert.pem")
		return
	}

	//split serverAddr into host and port
	host, _, err := net.SplitHostPort(*serverAddr)
	if err != nil {
		fmt.Println("Error splitting server address")
		return
	}
	creds, err := credentials.NewClientTLSFromFile(string(CAcert), host)
	if err != nil {
		fmt.Println("Error reading ca_cert.pem")
		return
	}
	
	p := tea.NewProgram(model(5), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tick(), tea.EnterAltScreen)
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
		
	case tickMsg:
		m--
		if m <= 0 {
			return m, tea.Quit
		}
		return m, tick()
	}
	
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("\n\n     Hi. This program will exit in %d seconds...", m)
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

