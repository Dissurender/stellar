package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ANSI codes for fmt
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

type Client struct {
	id        int
	cType     string
	outbox    chan Message
	inbox     chan Message
	neighbors []*Client
}

type Message struct {
	requestId   int
	requestType string
	from        int
	to          int
	data        string
	sendAt      time.Time
	latency     time.Duration
}

type Network struct {
	connections []*Connection
}

type Connection struct {
	from       *Client
	to         *Client
	latency    time.Duration
	packetLoss float64
}

// connect() adds a node to the network via .neighbors
func (c *Client) connect(other *Client, latency time.Duration, packetLoss float64) error {
	if other == nil {
		return fmt.Errorf("%scannot connect to a nil Client%s", Red, Reset)
	}
	if latency < 0 {
		return fmt.Errorf(fmt.Sprintf("%slatency must be non-negative%s", Red, Reset))
	}
	if packetLoss < 0 || packetLoss > 1 {
		return fmt.Errorf(fmt.Sprintf("%spacket loss is out of bounds%s", Red, Reset))
	}

	c.neighbors = append(c.neighbors, other)
	network.connections = append(network.connections, &Connection{
		from:       c,
		to:         other,
		latency:    latency,
		packetLoss: packetLoss,
	})

	return nil
}

// send() is a simple method for throwing messages to other clients
func (c *Client) send(msg Message) {
	c.outbox <- msg
}

// run() processes queued up messages a clients has
func (c *Client) run(wg *sync.WaitGroup) {
	defer wg.Done()

	for msg := range c.inbox {
		c.handleRequest(msg)
	}
}

var requestTypes = []string{"GetData", "UpdateData", "DeleteData"}

func (c *Client) handleRequest(msg Message) {
	switch msg.requestType {
	case "GetData":
		fmt.Printf("%sClient %d: Received GetData request from Client %d%s\n", Blue, c.id, msg.from, Reset)
	case "UpdateData":
		fmt.Printf("%sClient %d: Received UpdateData request from Client %d%s\n", Yellow, c.id, msg.from, Reset)
	case "DeleteData":
		fmt.Printf("%sClient %d: Received DeleteData request from Client %d%s\n", Magenta, c.id, msg.from, Reset)
	default:
		fmt.Printf("%sClient %d: Unknown request type from Client %d%s\n", Red, c.id, msg.from, Reset)
	}
}

// helper func to determine latency for messages
func randomDur(min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min)+min) * time.Millisecond
}

// initialize psuedo global values for use throughout the program
var network Network

var ClientCount = 10
var Clients = make([]*Client, ClientCount)

func main() {
	// rand.Seed(seed)
	rand.NewSource(time.Now().UnixNano())

	for i := 0; i < ClientCount; i++ {
		Clients[i] = &Client{
			id:        i,
			cType:     "client",
			outbox:    make(chan Message),
			inbox:     make(chan Message),
			neighbors: []*Client{},
		}
	}

	// add a bit of attributes that would apply in real situation
	for i := 0; i < ClientCount; i++ {
		for j := i + 1; j < ClientCount; j++ {
			latency := randomDur(10, 100)
			packetLoss := rand.Float64() * 0.1
			err := Clients[i].connect(Clients[j], latency, packetLoss)
			if err != nil {
				fmt.Printf("%sError connecting Client %d and Client %d:%s %v\n", Red, i, j, Reset, err)
			}
		}
	}

	// create wait groups for the different counts needed for routines
	var wg sync.WaitGroup
	var netWg sync.WaitGroup
	var sendWg sync.WaitGroup

	for _, Client := range Clients {
		wg.Add(1)
		go Client.run(&wg)
	}

	// Simulate sending messages
	for i := 0; i < ClientCount; i++ {
		sendWg.Add(1)
		go func(sender int) {
			defer sendWg.Done()
			for requestId, neighbor := range Clients[sender].neighbors {

				requestType := requestTypes[rand.Intn(len(requestTypes))]

				Clients[sender].send(Message{
					requestId:   requestId,
					requestType: requestType,
					from:        sender,
					to:          neighbor.id,
					data:        fmt.Sprintf("Request from Client %d!", sender),
					sendAt:      time.Now(),
					latency:     randomDur(10, 100),
				})
			}
		}(i)
	}

	// Simulate the network passing messages between Clients
	for i := 0; i < ClientCount; i++ {
		netWg.Add(1)
		go func(ClientIndex int) {
			defer netWg.Done()
			for msg := range Clients[ClientIndex].outbox {
				for _, conn := range network.connections {
					if conn.from.id == msg.from && conn.to.id == msg.to {
						// Simulate latency
						time.Sleep(conn.latency)

						// Simulate packet loss
						if rand.Float64() < conn.packetLoss {
							fmt.Printf("%sPacket loss: Client %d -> Client %d%s\n", Red, msg.from, msg.to, Reset)
							continue
						}

						Clients[msg.to].inbox <- msg
					}
				}
			}
		}(i)
	}

	sendWg.Wait()

	runCLI()

	// tie off the open channels
	for i := 0; i < ClientCount; i++ {
		close(Clients[i].outbox)
		close(Clients[i].inbox)
	}

}

func runCLI() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%sEnter a request with the format: <from> <to> <requestType> (e.g., 0 1 GetData)%s\n", Green, Reset)
		fmt.Printf("%sOr type 'exit' to quit: %s\n", Green, Reset)

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("%sError reading input: %s%s\n", Red, err, Reset)
			continue
		}

		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}

		parts := strings.Split(input, " ")
		if len(parts) != 3 {
			fmt.Printf("%sInvalid input format, please try again.%s\n", Red, Reset)
			continue
		}

		from, err := strconv.Atoi(parts[0])
		if err != nil || from < 0 || from >= ClientCount {
			fmt.Printf("%sInvalid 'from' Client ID, please try again.%s\n", Red, Reset)
			continue
		}

		to, err := strconv.Atoi(parts[1])
		if err != nil || to < 0 || to >= ClientCount {
			fmt.Printf("%sInvalid 'to' Client ID, please try again.%s\n", Red, Reset)
			continue
		}

		requestType := parts[2]
		if !contains(requestTypes, requestType) {
			fmt.Printf("%sInvalid request type. Allowed types: %v%s\n", Red, requestTypes, Reset)
			continue
		}

		requestID := rand.Int()
		Clients[from].send(Message{
			requestId:   requestID,
			requestType: requestType,
			from:        from,
			to:          to,
			data:        fmt.Sprintf("%sRequest from Client %d!%s", Magenta, from, Reset),
			sendAt:      time.Now(),
			latency:     randomDur(10, 100),
		})
	}
}

// contains() is a helper function for runCLI()

func contains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}
