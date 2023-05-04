package main

import "time"

type Message struct {
	requestId   int
	requestType string
	from        int
	to          int
	data        string
	sendAt      time.Time
	latency     time.Duration
}
