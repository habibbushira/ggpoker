package p2p

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type PlayersReady struct {
	mu         sync.RWMutex
	recvStatus map[string]bool
}

func NewPlayersReady() *PlayersReady {
	return &PlayersReady{
		recvStatus: make(map[string]bool),
	}
}

func (pr *PlayersReady) addRecvStatus(from string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.recvStatus[from] = true
}

func (pr *PlayersReady) len() int {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	return len(pr.recvStatus)
}

func (pr *PlayersReady) clear() {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.recvStatus = make(map[string]bool)
}

type Game struct {
	listenAddr   string
	broadcastTo  chan BroadcastTo
	playersReady *PlayersReady
	playersList  PlayersList

	// currentStatus shoud be automatically accessable
	currentStatus GameStatus
}

func NewGame(addr string, bc chan BroadcastTo) *Game {
	g := &Game{
		listenAddr:    addr,
		broadcastTo:   bc,
		playersReady:  NewPlayersReady(),
		currentStatus: GameStatusConnected,
	}

	go g.loop()

	return g
}

func (g *Game) SetReady() {
	g.playersReady.addRecvStatus(g.listenAddr)
	g.sendToPlayers(MessageReady{})

	g.setStatus(GameStatusReady)
}

func (g *Game) setStatus(s GameStatus) {
	if g.currentStatus != s {
		atomic.StoreInt32((*int32)(&g.currentStatus), (int32)(s))
	}
}

func (g *Game) sendToPlayers(payload any) {
	g.broadcastTo <- BroadcastTo{
		To:      g.playersList,
		Payload: payload,
	}

	fmt.Printf("sendign to players payload: %v, players: %v", payload, g.playersList)
}

func (g *Game) AddPlayer(from string) {
	g.playersList = append(g.playersList, from)
	sort.Sort(g.playersList)
	g.playersReady.addRecvStatus(from)
}

func (g *Game) loop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		<-ticker.C
		fmt.Printf("players connected: we: %s, status: %s\n", g.listenAddr, g.currentStatus)
		fmt.Printf("%s\n", g.playersList)
	}
}

// type GameState struct {
// 	listenAddr  string
// 	broadcastch chan BroadcastTo
// 	isDealer    bool       // atomic accessable
// 	gameStatus  GameStatus // atomic accessable

// 	playersLock sync.RWMutex
// 	players     map[string]*Player
// 	playersList PlayersList
// }

// func NewGameState(addr string, broadcastch chan BroadcastTo) *GameState {
// 	g := &GameState{
// 		listenAddr:  addr,
// 		broadcastch: broadcastch,
// 		isDealer:    false,
// 		gameStatus:  GameStatusWaiting,
// 		players:     make(map[string]*Player),
// 	}

// 	g.AddPlayer(addr, GameStatusWaiting)

// 	go g.loop()
// 	return g
// }

// // TODO: (@habib) Check other read and write occurences of the GameStatus
// func (g *GameState) SetStatus(s GameStatus) {
// 	if g.gameStatus != s {
// 		atomic.StoreInt32((*int32)(&g.gameStatus), (int32)(s))
// 		g.SetPlayerStatus(g.listenAddr, s)
// 	}
// }

// func (g *GameState) playersWaitingForCards() int {
// 	totalPlayers := 0
// 	for _, player := range g.playersList {
// 		if player.Status == GameStatusWaiting {
// 			totalPlayers++
// 		}
// 	}

// 	return totalPlayers
// }

// func (g *GameState) CheckNeedDealCards() {
// 	playersWaiting := g.playersWaitingForCards()

// 	if playersWaiting == len(g.players) &&
// 		g.isDealer &&
// 		g.gameStatus == GameStatusWaiting {
// 		fmt.Printf("need to deal cards %s\n", g.listenAddr)

// 		g.InitiateShuffleAndDeal()
// 	}
// }

// func (g *GameState) GetPlayersWithStatus(s GameStatus) []string {
// 	players := []string{}
// 	for addr, player := range g.players {
// 		if player.Status == s {
// 			players = append(players, addr)
// 		}
// 	}

// 	return players

// }

// // index of our own position on the table
// func (g *GameState) getPositionOnTable() int {
// 	for index, player := range g.playersList {
// 		if g.listenAddr == player.ListenAddr {
// 			return index
// 		}
// 	}

// 	return -1
// }

// func (g *GameState) getPrevPositionOnTable() int {
// 	ourPosition := g.getPositionOnTable()

// 	// if we are in the first position of the table
// 	// we should return the last index of the player lists
// 	if ourPosition == 0 {
// 		return len(g.playersList) - 1
// 	}

// 	return ourPosition - 1
// }

// // getNextPosition return the index of the next player in teh PlayersList.
// func (g *GameState) getNextPositionOnTable() int {
// 	ourPosition := g.getPositionOnTable()
// 	if ourPosition == -1 {
// 		panic("getNextPositionOnTable errror")
// 	}

// 	if ourPosition == len(g.playersList)-1 {
// 		return 0
// 	}

// 	return ourPosition + 1

// }

// func (g *GameState) ShuffleAndEncrypt(from string, deck [][]byte) error {
// 	g.SetPlayerStatus(from, GameStatusShuffleAndDeal)

// 	prevPlayer := g.playersList[g.getPrevPositionOnTable()]

// 	if g.isDealer && from == prevPlayer.ListenAddr {
// 		fmt.Println("********************************************")
// 		fmt.Println("end shuffle round flip", from, g.listenAddr)
// 		fmt.Println("********************************************")
// 		return nil
// 	}

// 	dealToPlayer := g.playersList[g.getNextPositionOnTable()]
// 	//encyrption and shuffle

// 	g.SetStatus(GameStatusShuffleAndDeal)
// 	g.SendToPlayer(dealToPlayer.ListenAddr, MessageEncDeck{Deck: [][]byte{}})

// 	fmt.Printf("%s => %s\n", g.listenAddr, GameStatusShuffleAndDeal)
// 	return nil
// }

// // InitiateShuffleAndDeal is only used for the "real" dealer. The actual "button player"
// func (g *GameState) InitiateShuffleAndDeal() {
// 	dealToPlayer := g.playersList[g.getNextPositionOnTable()]

// 	g.SetStatus(GameStatusShuffleAndDeal)
// 	g.SendToPlayer(dealToPlayer.ListenAddr, MessageEncDeck{Deck: [][]byte{}})
// }

// func (g *GameState) SendToPlayer(addr string, payload any) {
// 	g.broadcastch <- BroadcastTo{
// 		To:      []string{addr},
// 		Payload: payload,
// 	}

// 	fmt.Printf("sending payload: %v to player: %s\n", payload, addr)
// }

// func (g *GameState) SendToPlayersWithStatus(payload any, s GameStatus) {
// 	players := g.GetPlayersWithStatus(s)

// 	g.broadcastch <- BroadcastTo{
// 		To:      players,
// 		Payload: payload,
// 	}

// 	fmt.Printf("sendign to players payload: %v, players: %v", payload, players)
// }

// func (g *GameState) SetPlayerStatus(addr string, status GameStatus) {
// 	player, ok := g.players[addr]
// 	if !ok {
// 		panic("player could not be found")
// 	}

// 	player.Status = status

// 	g.CheckNeedDealCards()
// }

// func (g *GameState) AddPlayer(addr string, status GameStatus) {
// 	g.playersLock.Lock()
// 	defer g.playersLock.Unlock()

// 	// if status == GameStatusWaiting {
// 	// 	g.AddPlayerWaiting()
// 	// }

// 	player := &Player{
// 		ListenAddr: addr,
// 	}
// 	g.players[addr] = player
// 	g.playersList = append(g.playersList, player)
// 	sort.Sort(g.playersList)

// 	// Seth the player status also when we add the player
// 	g.SetPlayerStatus(addr, status)

// 	fmt.Printf("new player joined: addr - %s, status - %s\n", addr, status)
// }

// func (g *GameState) loop() {
// 	ticker := time.NewTicker(time.Second * 5)
// 	for {
// 		<-ticker.C
// 		fmt.Printf("players connected: we: %s, status: %s\n", g.listenAddr, g.gameStatus)
// 		fmt.Printf("%s\n", g.playersList)
// 	}
// }

type PlayersList []string

func (list PlayersList) Len() int { return len(list) }
func (list PlayersList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
func (list PlayersList) Less(i, j int) bool {
	port1, _ := strconv.Atoi(list[i][1:])
	port2, _ := strconv.Atoi(list[j][1:])

	return port1 < port2
}
