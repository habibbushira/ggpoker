package main

import (
	"github.com/habibbushira/ggpoker/p2p"
)

func main() {
	// rand.New(rand.NewSource(time.Now().UnixNano()))
	// for i := 0; i < 10; i++ {
	// 	d := deck.New()
	// 	fmt.Println(d)
	// 	fmt.Println()
	// }

	cfg := p2p.ServerConfig{
		ListenAddr: ":3000",
	}
	server := p2p.NewServer(cfg)

	server.Start()

}
