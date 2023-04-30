package main

import (
	"fmt"
	"github.com/Dissurender/stellar"
	"math/rand"
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

var network Network

//
//
//

func main() {
	// rand.Seed(seed)
	rand.NewSource(time.Now().UnixNano())

	microserviceCount := 10
	microservices := make([]*Microservice, microserviceCount)

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

}
