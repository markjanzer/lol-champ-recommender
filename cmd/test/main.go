package main

import (
	"fmt"
)

func modifyInt(i *int) {
	localI := *i
	fmt.Println(localI)
	localI = 5
}

func main() {
	// fmt.Println(2 * time.Minute / 100)

	myInt := 3
	modifyInt(&myInt)
	fmt.Println(myInt)
}
