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

type Microservice struct {
	id        int
	outbox    chan Message
	inbox     chan Message
	neighbors []*Microservice
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
	from       *Microservice
	to         *Microservice
	latency    time.Duration
	packetLoss float64
}

func (ms *Microservice) connect(other *Microservice, latency time.Duration, packetLoss float64) error {
	if other == nil {
		return fmt.Errorf("%scannot connect to a nil Microservice%s", Red, Reset)
	}
	if latency < 0 {
		return fmt.Errorf(fmt.Sprintf("%slatency must be non-negative%s", Red, Reset))
	}
	if packetLoss < 0 || packetLoss > 1 {
		return fmt.Errorf(fmt.Sprintf("%spacket loss is out of bounds%s", Red, Reset))
	}

	ms.neighbors = append(ms.neighbors, other)
	network.connections = append(network.connections, &Connection{
		from:       ms,
		to:         other,
		latency:    latency,
		packetLoss: packetLoss,
	})

	return nil
}

func (ms *Microservice) send(msg Message) {
	ms.outbox <- msg
}

func (ms *Microservice) run(wg *sync.WaitGroup) {
	defer wg.Done()

	for msg := range ms.inbox {
		ms.handleRequest(msg)
	}
}

var requestTypes = []string{"GetData", "UpdateData", "DeleteData"}

func (ms *Microservice) handleRequest(msg Message) {
	switch msg.requestType {
	case "GetData":
		fmt.Printf("%sMicroservice %d: Received GetData request from Microservice %d%s\n", Blue, ms.id, msg.from, Reset)
	case "UpdateData":
		fmt.Printf("%sMicroservice %d: Received UpdateData request from Microservice %d%s\n", Yellow, ms.id, msg.from, Reset)
	case "DeleteData":
		fmt.Printf("%sMicroservice %d: Received DeleteData request from Microservice %d%s\n", Magenta, ms.id, msg.from, Reset)
	default:
		fmt.Printf("%sMicroservice %d: Unknown request type from Microservice %d%s\n", Red, ms.id, msg.from, Reset)
	}
}

func randomDur(min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min)+min) * time.Millisecond
}

var network Network

var microserviceCount = 10
var microservices = make([]*Microservice, microserviceCount)

func main() {
	// rand.Seed(seed)
	rand.NewSource(time.Now().UnixNano())

	for i := 0; i < microserviceCount; i++ {
		microservices[i] = &Microservice{
			id:        i,
			outbox:    make(chan Message),
			inbox:     make(chan Message),
			neighbors: []*Microservice{},
		}
	}

	for i := 0; i < microserviceCount; i++ {
		for j := i + 1; j < microserviceCount; j++ {
			latency := randomDur(10, 100)
			packetLoss := rand.Float64() * 0.1
			err := microservices[i].connect(microservices[j], latency, packetLoss)
			if err != nil {
				fmt.Printf("%sError connecting Microservice %d and Microservice %d:%s %v\n", Red, i, j, Reset, err)
			}
		}
	}

	// create wait groups for the different counts needed for routines
	var wg sync.WaitGroup
	var netWg sync.WaitGroup
	var sendWg sync.WaitGroup

	for _, microservice := range microservices {
		wg.Add(1)
		go microservice.run(&wg)
	}

	// Simulate sending messages
	for i := 0; i < microserviceCount; i++ {
		sendWg.Add(1)
		go func(sender int) {
			defer sendWg.Done()
			for requestId, neighbor := range microservices[sender].neighbors {
				// Choose a random request type
				requestType := requestTypes[rand.Intn(len(requestTypes))]

				microservices[sender].send(Message{
					requestId:   requestId,
					requestType: requestType,
					from:        sender,
					to:          neighbor.id,
					data:        fmt.Sprintf("Request from Microservice %d!", sender),
					sendAt:      time.Now(),
					latency:     randomDur(10, 100),
				})
			}
		}(i)
	}

	// Simulate the network passing messages between Microservices
	for i := 0; i < microserviceCount; i++ {
		netWg.Add(1)
		go func(MicroserviceIndex int) {
			defer netWg.Done()
			for msg := range microservices[MicroserviceIndex].outbox {
				for _, conn := range network.connections {
					if conn.from.id == msg.from && conn.to.id == msg.to {
						// Simulate latency
						time.Sleep(conn.latency)

						// Simulate packet loss
						if rand.Float64() < conn.packetLoss {
							fmt.Printf("%sPacket loss: Microservice %d -> Microservice %d%s\n", Red, msg.from, msg.to, Reset)
							continue
						}

						microservices[msg.to].inbox <- msg
					}
				}
			}
		}(i)
	}

	sendWg.Wait()

	runCLI()

	for i := 0; i < microserviceCount; i++ {
		close(microservices[i].outbox)
		close(microservices[i].inbox)
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
		if err != nil || from < 0 || from >= microserviceCount {
			fmt.Printf("%sInvalid 'from' microservice ID, please try again.%s\n", Red, Reset)
			continue
		}

		to, err := strconv.Atoi(parts[1])
		if err != nil || to < 0 || to >= microserviceCount {
			fmt.Printf("%sInvalid 'to' microservices ID, please try again.%s\n", Red, Reset)
			continue
		}

		requestType := parts[2]
		if !contains(requestTypes, requestType) {
			fmt.Printf("%sInvalid request type. Allowed types: %v%s\n", Red, requestTypes, Reset)
			continue
		}

		requestID := rand.Int()
		microservices[from].send(Message{
			requestId:   requestID,
			requestType: requestType,
			from:        from,
			to:          to,
			data:        fmt.Sprintf("%sRequest from microservices %d!%s", Magenta, from, Reset),
			sendAt:      time.Now(),
			latency:     randomDur(10, 100),
		})
	}
}

func contains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}
