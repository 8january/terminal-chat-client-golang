package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
	"golang.org/x/term"
)

var (
	labelColor    = color.New(color.FgHiGreen).SprintFunc()
	outputColor   = color.New(color.FgRed).SprintFunc()
	receivedColor = color.New(color.FgYellow).SprintFunc()
	addr          = flag.String("addr", "localhost:8080", "http service address")
	server        = flag.String("server", "default", "Server to connect to: 'local' or 'render'")
)

type Client struct {
	conn      *websocket.Conn
	send      chan []byte
	receive   chan Message
	done      chan struct{}
	interrupt chan os.Signal
}

type UI struct {
	app          *tview.Application
	messagesView *tview.TextView
	inputField   *tview.InputField
	sendMessage  chan string
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Room string `json:"room"`
}

type Message struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Room    string `json:"room"`
	Content string `json:"content"`
}

func NewUser(_name string, _room string) *User {
	id := uuid.New().String()
	return &User{
		ID: id, Name: _name, Room: _room,
	}
}

func NewClient(addr string) (*Client, error) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	var uri url.URL
	switch *server {
	case "render":
		uri = url.URL{Scheme: "wss", Host: "terminal-chat-server-golang.onrender.com", Path: "/ws"}
	default:
		uri = url.URL{Scheme: "ws", Host: addr, Path: "/"}
	}

	conn, _, err := websocket.DefaultDialer.Dial(uri.String(), nil)
	if err != nil {
		log.Println("Dial error:", err)
		return nil, err
	}

	return &Client{
		conn:      conn,
		send:      make(chan []byte),
		receive:   make(chan Message),
		done:      make(chan struct{}),
		interrupt: interrupt,
	}, nil

}

func (c *Client) run() {
	go c.readPump()
	go c.writePump()

	for {
		select {
		case <-c.done:
			return
		case <-c.interrupt:
			log.Println("Interrupting app")
			err := c.conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			if err != nil {
				log.Println("Write close error:", err)
			}
			os.Exit(0)
		}
	}
}

func (c *Client) readPump() {
	defer c.conn.Close()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("JSON unmarshalling error:", err)
		}

		c.receive <- msg
	}
}

func (c *Client) writePump() {
	for {
		select {
		case message := <-c.send:
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Write error:", err)
				return
			}
		case <-c.done:
			return
		}
	}
}

func NewUI() *UI {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	app := tview.NewApplication()
	messagesView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)

	messagesView.
		SetBorder(false).
		SetBackgroundColor(tcell.ColorDefault)

	inputField := tview.NewInputField().
		SetLabel("> ").SetFieldBackgroundColor(tcell.ColorDefault)

	inputField.
		SetBorder(false)

	sendMessage := make(chan string)

	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			message := inputField.GetText()
			if message != "" {
				fmt.Fprintf(messagesView, "[red]You: [white]%s\n", message)
				sendMessage <- message
				inputField.SetText("")
				messagesView.ScrollToEnd()
			}
		}
	})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(messagesView, 0, 9, false).
		AddItem(inputField, 0, 1, true)

	app.
		SetRoot(flex, true).
		SetFocus(inputField)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTAB {
			if app.GetFocus() == inputField {
				app.SetFocus(messagesView)
			} else {
				app.SetFocus(inputField)
			}
		}
		return event
	})

	return &UI{
		app:          app,
		messagesView: messagesView,
		inputField:   inputField,
		sendMessage:  sendMessage,
	}
}

func (ui *UI) run() {
	if err := ui.app.Run(); err != nil {
		panic(err)
	}
}

func (ui *UI) displayMessage(msg Message) {
	ui.app.QueueUpdateDraw(func() {
		switch msg.Type {
		case "server_message":
			fmt.Fprintf(ui.messagesView, "[yellow]%s\n", msg.Content)
		case "user_message":
			fmt.Fprintf(ui.messagesView, "[green]%s: [white]%s\n", msg.Name, msg.Content)
		default:
			fmt.Fprintf(ui.messagesView, "[white]%s\n", msg.Content)
		}
		ui.messagesView.ScrollToEnd()

	})

}

func getUserInfo() *User {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println("Error getting terminal size:", err)
		return nil
	}

	clearTerm()
	centerSpace(height)
	printCentered(labelColor("Who are you?: "), width)

	name := input()

	printCentered(labelColor("Room: "), width)

	room := input()
	clearTerm()

	return NewUser(name, room)
}

func handleMessages(client *Client, ui *UI, user *User) {
	for {
		select {
		case msg := <-client.receive:
			ui.displayMessage(msg)
		case text := <-ui.sendMessage:
			message := Message{
				ID:      user.ID,
				Type:    "user_message",
				Name:    user.Name,
				Room:    user.Room,
				Content: text,
			}

			jsonMessage, err := json.Marshal(message)
			if err != nil {
				log.Println("JSON marshalling error:", err)
			}

			client.send <- jsonMessage
		case <-client.done:
			return
		}
	}

}

func main() {
	flag.Usage = printHelp

	help := flag.Bool("help", false, "Show help message")
	flag.BoolVar(help, "h", false, "Show help message (shorthand)")

	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	user := getUserInfo()
	client, err := NewClient(*addr)
	if err != nil {
		log.Println("New client error:", err)
	}
	defer client.conn.Close()

	ui := NewUI()
	log.Println("Connected to room", user.Room)
	log.Println("Press ctrl^c to close application")

	go client.run()
	go ui.run()

	handleMessages(client, ui, user)
}

func printHelp() {
	fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
	fmt.Println("Options:")
	fmt.Println("  -server string")
	fmt.Println("        Server to connect to: 'local' or 'render' (default \"local\")")
	fmt.Println("  -addr string")
	fmt.Println("        HTTP service address (default \"localhost:8080\")")
	fmt.Println("  -h, --help")
	fmt.Println("        Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  Connect to local server:")
	fmt.Printf("    %s\n", os.Args[0])
	fmt.Println("  Connect to Render server:")
	fmt.Printf("    %s -server render\n", os.Args[0])
	fmt.Println("  Connect to custom address:")
	fmt.Printf("    %s -addr example.com:8080\n", os.Args[0])
}

func printCentered(message string, totalWidth int) {
	padding := (totalWidth - len(message)) / 2
	fmt.Print(strings.Repeat(" ", padding) + message)
}

func centerSpace(height int) {
	for i := 0; i < height/2-1; i++ {
		fmt.Println()
	}
}

func clearTerm() {
	fmt.Print("\033[H\033[2J")
}

func input() string {
	reader := bufio.NewReader(os.Stdin)
	data, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Reading error:", err)
		return ""
	}

	return strings.ReplaceAll(strings.TrimSpace(data), " ", "_")
}
