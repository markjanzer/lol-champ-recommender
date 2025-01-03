package main

import (
	"fmt"
	"math"
)

func wilsonScore(wins, games float64) float64 {
	if games == 0 {
		return 0
	}
	z := 1.96 // 95% confidence
	phat := wins / games
	return (phat + z*z/(2*games) - z*math.Sqrt((phat*(1-phat)+z*z/(4*games))/games)) / (1 + z*z/games)
}

func main() {
	fmt.Println("wilsonScore(5, 9):", wilsonScore(5, 9))
	fmt.Println("wilsonScore(10, 10):", wilsonScore(10, 10))
	fmt.Println("wilsonScore(10, 100):", wilsonScore(10, 100))
}
