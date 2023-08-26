package main

import (
	"net/http"
	"time"

	"github.com/habibbushira/ggpoker/p2p"
)

func makeServer(addr, apiAddr string) *p2p.Server {
	cfg := p2p.ServerConfig{
		ListenAddr:    addr,
		Version:       "GGPOKER V0.1-alpha",
		GameVariant:   p2p.TexasHoldem,
		ApiListenAddr: apiAddr,
	}
	server := p2p.NewServer(cfg)
	go server.Start()

	time.Sleep(200 * time.Millisecond)

	return server
}

func main() {
	playerA := makeServer(":3000", ":3001")
	playerB := makeServer(":4000", ":4001")
	playerC := makeServer(":5000", ":5001")
	playerD := makeServer(":6000", ":6001")

	playerB.Connect(playerA.ListenAddr)
	time.Sleep(200 * time.Millisecond)

	playerC.Connect(playerB.ListenAddr)
	time.Sleep(200 * time.Millisecond)

	playerD.Connect(playerC.ListenAddr)

	go func() {
		time.Sleep(time.Second * 1)
		http.Get("http://localhost:3001/ready")

		time.Sleep(time.Second * 1)
		http.Get("http://localhost:4001/ready")

		// time.Sleep(time.Second * 1)
		// http.Get("http://localhost:5001/ready")

		// time.Sleep(time.Second * 1)
		// http.Get("http://localhost:6001/ready")
	}()

	select {}

}
