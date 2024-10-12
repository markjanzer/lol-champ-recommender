package main

import (
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	// fmt.Println(2 * time.Minute / 100)
	fmt.Println(rate.Every(2 * time.Minute / 100))
}
