package main

import (
	"fmt"
	"sync"
	"time"
)

type Microservice struct {
	id        int
	outbox    chan Message
	inbox     chan Message
	neighbors []*Microservice
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

func (ms *Microservice) run(wg *sync.WaitGroup) {
	for msg := range ms.inbox {
		ms.handleRequest(msg)
	}
}

func (ms *Microservice) send(msg Message) {
	ms.outbox <- msg
}
