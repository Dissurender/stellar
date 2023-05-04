package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func runCli(microservices []*Microservice, requestTypes []string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\nEnter a request in the format: <from> <to> <requestType> (e.g., 0 1 GetData)")
		fmt.Print("Or type 'exit' to quit: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}

		parts := strings.Split(input, " ")
		if len(parts) != 3 {
			fmt.Println("Invalid input format. Please try again.")
			continue
		}

		from, err := strconv.Atoi(parts[0])
		if err != nil || from < 0 || from >= len(microservices) {
			fmt.Println("Invalid 'from' microservice ID. Please try again.")
			continue
		}

		to, err := strconv.Atoi(parts[1])
		if err != nil || to < 0 || to >= len(microservices) {
			fmt.Println("Invalid 'to' microservice ID. Please try again.")
			continue
		}

		requestType := parts[2]
		if !contains(requestTypes, requestType) {
			fmt.Printf("Invalid request type. Allowed types: %v\n", requestTypes)
			continue
		}

		requestId := rand.Int()
		microservices[from].send(Message{
			requestId:   requestId,
			requestType: requestType,
			from:        from,
			to:          to,
			data:        fmt.Sprintf("Request from Microservice %d!", from),
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
