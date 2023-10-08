package main

import (
	"time"

	"github.com/habibbushira/ggpoker/p2p"
)

func makeServer(addr, apiAddr string) *p2p.Node {
	cfg := p2p.ServerConfig{
		ListenAddr:    addr,
		Version:       "GGPOKER V0.1-alpha",
		GameVariant:   p2p.TexasHoldem,
		ApiListenAddr: apiAddr,
	}
	server := p2p.NewNode(cfg)
	go server.Start()

	time.Sleep(200 * time.Millisecond)

	return server
}

func main() {
	node1 := makeServer(":3000", ":3001")
	node2 := makeServer(":4000", ":4001")
	node3 := makeServer(":5000", ":5001")
	node4 := makeServer(":6000", ":6001")

	node2.Connect(node1.ListenAddr)
	node3.Connect(node2.ListenAddr)
	node4.Connect(node3.ListenAddr)

	select {}

	// playerB := makeServer(":4000", ":4001")
	// playerC := makeServer(":5000", ":5001")
	// playerD := makeServer(":6000", ":6001")

	// playerB.Connect(playerA.ListenAddr)
	// time.Sleep(200 * time.Millisecond)

	// playerC.Connect(playerB.ListenAddr)
	// time.Sleep(200 * time.Millisecond)

	// playerD.Connect(playerC.ListenAddr)

	// go func() {
	// 	time.Sleep(time.Second * 1)
	// 	http.Get("http://localhost:3001/ready")

	// 	// time.Sleep(time.Second * 1)
	// 	// http.Get("http://localhost:4001/ready")

	// 	time.Sleep(time.Second * 1)
	// 	http.Get("http://localhost:5001/ready")

	// 	time.Sleep(time.Second * 1)
	// 	http.Get("http://localhost:6001/ready")

	// 	//PREFLOP
	// 	// time.Sleep(time.Second * 5)
	// 	// http.Get("http://localhost:4001/fold")

	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:5001/fold")

	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:6001/fold")

	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:3001/fold")

	// 	// //FOLD
	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:4001/fold")

	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:5001/fold")

	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:6001/fold")

	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:3001/fold")

	// 	// //TURN
	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:4001/fold")

	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:5001/fold")

	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:6001/fold")

	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:3001/fold")

	// 	// //RIVER
	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:4001/fold")

	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:5001/fold")

	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:6001/fold")

	// 	// time.Sleep(time.Second * 2)
	// 	// http.Get("http://localhost:3001/fold")
	// }()

	// select {}

}
