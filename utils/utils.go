package utils

import (
	"math/rand"
	"time"
)

func RandomDuration(min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min)+min) * time.Millisecond
}

func Contains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}
