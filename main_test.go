package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"
)

func Test_connect(t *testing.T) {

	t.Run("nil Client", func(t *testing.T) {
		ms := &Client{}
		err := ms.connect(nil, 0, 0)
		if err == nil {
			t.Fatal("Expected error when connecting to a nil Client")
		}
	})

	t.Run("negative latency", func(t *testing.T) {
		ms1 := &Client{}
		ms2 := &Client{}
		err := ms1.connect(ms2, -1*time.Millisecond, 0)
		if err == nil {
			t.Fatal("Expected error when connecting with negative latency")
		}
	})

	t.Run("negative packet loss", func(t *testing.T) {
		ms1 := &Client{}
		ms2 := &Client{}
		err := ms1.connect(ms2, 0, -0.1)
		if err == nil {
			t.Fatal("Expected error when connecting with negative packet loss")
		}
	})

	t.Run("packet loss greater than 1", func(t *testing.T) {
		ms1 := &Client{}
		ms2 := &Client{}
		err := ms1.connect(ms2, 0, 1.1)
		if err == nil {
			t.Fatal("Expected error when connecting with packet loss greater than 1")
		}
	})

	t.Run("successful connection", func(t *testing.T) {
		ms1 := &Client{}
		ms2 := &Client{}
		err := ms1.connect(ms2, 100*time.Millisecond, 0.1)
		if err != nil {
			t.Fatalf("Unexpected error when connecting: %v", err)
		}

		if len(ms1.neighbors) != 1 || ms1.neighbors[0] != ms2 {
			t.Fatal("Client connection not established correctly")
		}

		if len(network.connections) != 1 {
			t.Fatal("Network connection not added")
		}

		conn := network.connections[0]
		if conn.from != ms1 || conn.to != ms2 || conn.latency != 100*time.Millisecond || conn.packetLoss != 0.1 {
			t.Fatal("Network connection properties not set correctly")
		}

	})

}

// captureOutput is a helper function for Test_handleRequest()
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	err := w.Close()
	if err != nil {
		panic(err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

func Test_handleRequest(t *testing.T) {
	ms := &Client{id: 1}

	t.Run("handle GetData request", func(t *testing.T) {
		msg := Message{from: 2, requestType: "GetData"}
		output := captureOutput(func() {
			ms.handleRequest(msg)
		})

		expected := fmt.Sprintf("%sClient 1: Received GetData request from Client 2%s\n", Blue, Reset)
		if output != expected {
			t.Fatalf("Expected output: %q, got: %q", expected, output)
		}
	})

	t.Run("handle UpdateData request", func(t *testing.T) {
		msg := Message{from: 2, requestType: "UpdateData"}
		output := captureOutput(func() {
			ms.handleRequest(msg)
		})

		expected := fmt.Sprintf("%sClient 1: Received UpdateData request from Client 2%s\n", Yellow, Reset)
		if output != expected {
			t.Fatalf("Expected output: %q, got: %q", expected, output)
		}
	})

	t.Run("handle DeleteData request", func(t *testing.T) {
		msg := Message{from: 2, requestType: "DeleteData"}
		output := captureOutput(func() {
			ms.handleRequest(msg)
		})

		expected := fmt.Sprintf("%sClient 1: Received DeleteData request from Client 2%s\n", Magenta, Reset)
		if output != expected {
			t.Fatalf("Expected output: %q, got: %q", expected, output)
		}
	})

	t.Run("handle unknown request type", func(t *testing.T) {
		msg := Message{from: 2, requestType: "InvalidRequest"}
		output := captureOutput(func() {
			ms.handleRequest(msg)
		})

		expected := fmt.Sprintf("%sClient 1: Unknown request type from Client 2%s\n", Red, Reset)
		if output != expected {
			t.Fatalf("Expected output: %q, got: %q", expected, output)
		}
	})
}

func Test_randomDur(t *testing.T) {
	// rand.Seed(seed)
	rand.NewSource(time.Now().UnixNano())

	min := 10
	max := 100

	for i := 0; i < 1000; i++ {
		duration := randomDur(min, max)

		if duration < time.Duration(min)*time.Millisecond || duration >= time.Duration(max)*time.Millisecond {
			t.Fatalf("Expected duration to be between %d and %d but got %v", min, max, duration)
		}
	}

	var sameCount int
	previousDur := randomDur(min, max)
	for i := 0; i < 1000; i++ {
		duration := randomDur(min, max)
		if duration == previousDur {
			sameCount++
		}
		previousDur = duration
	}

	if sameCount > 900 {
		t.Fatal("Expected random duration, but most are same")
	}

}

func Test_Contains(t *testing.T) {
	t.Run("slice contains target string", func(t *testing.T) {
		arr := []string{"GetData", "UpdateData", "DeleteData"}
		str := "DeleteData"

		if !contains(arr, str) {
			t.Fatalf("Expected to find %q in the slice, but it was not found", str)
		}
	})

	t.Run("slice does not contain target string", func(t *testing.T) {
		arr := []string{"GetData", "UpdateData", "DeleteData"}
		str := "grape"

		if contains(arr, str) {
			t.Fatalf("Expected not to find %q in the slice, but it was found", str)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		arr := []string{}
		str := "apple"

		if contains(arr, str) {
			t.Fatalf("Expected not to find %q in the empty slice, but it was found", str)
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		var arr []string
		str := "apple"

		if contains(arr, str) {
			t.Fatalf("Expected not to find %q in the nil slice, but it was found", str)
		}
	})
}
