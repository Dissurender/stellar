package main

import (
	"math/rand"
	"time"
)

type Network struct {
	connections []*Connection
}

type Connection struct {
	from       *Microservice
	to         *Microservice
	latency    time.Duration
	packetLoss float64
}

func randomDur(min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min)+min) * time.Millisecond
}
