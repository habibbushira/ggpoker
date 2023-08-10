package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/habibbushira/ggpocker/deck"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 10; i++ {
		d := deck.New()
		fmt.Println(d)
		fmt.Println()
	}
}
