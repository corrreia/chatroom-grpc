package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var addr = flag.String("addr", "localhost:8421", "the address to connect to")

func main() {
	flag.Parse()

	getCA(*addr, "./certs")

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 1000

	x, y, err := term.GetSize(0)
	if err != nil {
		log.Fatal(err)
	}

	ta.SetWidth(x)
	ta.SetHeight(1)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(x, y-5)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}
	
	case tea.WindowSizeMsg:
		m.textarea.SetWidth(msg.Width) //* i feel like this is a really goofy way to do this but it works
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5 
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

func getCA(address string, path string) {
	os.Mkdir(path, 0777)
	var re = regexp.MustCompile(`(?m)[:]`)
    
	filePath := filepath.Join(  //join the path with the file name, replacing the ":" with "_"
		path, 
		re.ReplaceAllString(
			fmt.Sprintf(
				"ca_cert_%s.pem", 
				address),
				"_"))
	
	//check if the file already exists and return if it does
	_, err := os.Stat(filePath)
    if err == nil {
        return
    }

	// Set up a UDP connection to the server.
	conn, err := net.Dial("udp", address)
	if err != nil {
	    log.Fatalf("did not connect: %v", err)
    }

	conn.Write([]byte("HELLO"))
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) //if the server doesn't respond in 5 seconds, the connection is closed
	
	//buffer for the ca cert
	buffer := make([]byte, 2121) // 2121 is the size of all the certificates

    n, err := conn.Read(buffer)
	if err!= nil {
		log.Fatalf("failed to read: %v", err)
		return
	}

    ioutil.WriteFile(filePath, buffer[:n], 0666)
	if err != nil {
		log.Fatalf("could not write to file: %v", err)
	}
}