package p2p

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type PlayerAction byte

const (
	PlayerActionFold PlayerAction = iota + 1
	PlayerActionCheck
	PlayerActionBet
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

// func (pr *PlayersReady) haveRecv(from string) bool {
// 	pr.mu.Lock()
// 	defer pr.mu.Unlock()

// 	_, ok := pr.recvStatus[from]
// 	return ok
// }

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
	listenAddr  string
	broadcastTo chan BroadcastTo

	playersReady *PlayersReady
	playersList  PlayersList

	// currentStatus and currentDealer shoud be automatically accessable
	currentStatus GameStatus
	// NOTE: this will be -1 when the game is in a bootstrapped state
	currentDealer int32
	// currentPlayerTurn should be atomically accessable
	currentPlayerTurn int32
}

func NewGame(addr string, bc chan BroadcastTo) *Game {
	g := &Game{
		listenAddr:    addr,
		broadcastTo:   bc,
		playersReady:  NewPlayersReady(),
		currentStatus: GameStatusConnected,
		currentDealer: -1,
	}

	g.playersList = append(g.playersList, addr)

	go g.loop()

	return g
}

func (g *Game) Fold() {
	g.setStatus(GameStatusFolded)

	g.sendToPlayers(MessagePlayerAction{
		Action:            PlayerActionFold,
		CurrentGameStatus: g.currentStatus,
	}, g.getOtherPlayers()...)
}

func (g *Game) ShuffleAndEncrypt(from string, deck [][]byte) error {

	prevPlayer := g.playersList[g.getPrevPositionOnTable()]

	if from != prevPlayer {
		return fmt.Errorf("recieved encrypted deck from the wrong player (%s) should be (%s)", from, prevPlayer)
	}

	_, isDealer := g.getCurrentDealerAddr()

	if isDealer && from == prevPlayer {
		g.setStatus(GameStatusPreFlop)
		g.sendToPlayers(MessagePreFlop{}, g.getOtherPlayers()...)
		return nil
	}

	dealToPlayer := g.playersList[g.getNextPositionOnTable()]
	// dealToPlayer := g.getNextReadyPlayer(g.getNextPositionOnTable())
	//encyrption and shuffle

	g.setStatus(GameStatusDealing)
	g.sendToPlayers(MessageEncDeck{Deck: [][]byte{}}, dealToPlayer)

	fmt.Printf("received cards and going to shuffle: we %s, from %s, dealing %s\n", g.listenAddr, from, dealToPlayer)
	return nil
}

// InitiateShuffleAndDeal is only used for the "real" dealer. The actual "button player"
func (g *Game) InitiateShuffleAndDeal() {
	dealToPlayer := g.playersList[g.getNextPositionOnTable()]

	g.setStatus(GameStatusDealing)

	g.sendToPlayers(MessageEncDeck{Deck: [][]byte{}}, dealToPlayer)

	fmt.Printf("Dealing cards: we %s, to %s\n", g.listenAddr, dealToPlayer)
}

func (g *Game) SetPlayerReady(from string) {
	g.playersReady.addRecvStatus(from)
	fmt.Printf("setting player satus to ready we: %s, player: %s\n", g.listenAddr, from)

	// Check if the round can be started
	if g.playersReady.len() < 2 {
		return
	}

	// In this case we have engough players, hence, the round can be started
	// g.playersReady.clear()

	if _, ok := g.getCurrentDealerAddr(); ok {
		g.InitiateShuffleAndDeal()
	}
}

func (g *Game) SetReady() {
	g.playersReady.addRecvStatus(g.listenAddr)
	g.sendToPlayers(MessageReady{}, g.getOtherPlayers()...)

	g.setStatus(GameStatusReady)
}

func (g *Game) getCurrentDealerAddr() (string, bool) {
	currentDealer := g.playersList[0]
	if g.currentDealer > -1 {
		currentDealer = g.playersList[g.currentDealer]
	}

	return currentDealer, g.listenAddr == currentDealer
}

func (g *Game) SetStatus(s GameStatus) {
	g.setStatus(s)
}

func (g *Game) setStatus(s GameStatus) {
	if g.currentStatus != s {
		atomic.StoreInt32((*int32)(&g.currentStatus), (int32)(s))
	}
}

func (g *Game) sendToPlayers(payload any, addr ...string) {
	g.broadcastTo <- BroadcastTo{
		To:      addr,
		Payload: payload,
	}

	fmt.Printf("sending to players payload: %v, players: %v, we %s\n", payload, addr, g.listenAddr)
}

func (g *Game) AddPlayer(from string) {
	g.playersList = append(g.playersList, from)
	sort.Sort(g.playersList)
}

func (g *Game) loop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		<-ticker.C

		currentDealer, _ := g.getCurrentDealerAddr()
		fmt.Printf("players connected: we: %s, status: %s, dealer %s\n", g.listenAddr, g.currentStatus, currentDealer)
		fmt.Printf("playersList: %s\n", g.playersList)
	}
}

// index of our own position on the table
func (g *Game) getPositionOnTable() int {
	for index, player := range g.playersList {
		if g.listenAddr == player {
			return index
		}
	}

	return -1
}

func (g *Game) getPrevPositionOnTable() int {
	ourPosition := g.getPositionOnTable()

	// if we are in the first position of the table
	// we should return the last index of the player lists
	if ourPosition == 0 {
		return len(g.playersList) - 1
	}

	return ourPosition - 1
}

// getNextPosition return the index of the next player in teh PlayersList.
func (g *Game) getNextPositionOnTable() int {
	ourPosition := g.getPositionOnTable()
	if ourPosition == -1 {
		panic("getNextPositionOnTable errror")
	}

	if ourPosition == len(g.playersList)-1 {
		return 0
	}

	return ourPosition + 1

}

func (g *Game) getOtherPlayers() []string {
	players := []string{}

	for _, addr := range g.playersList {
		if addr == g.listenAddr {
			continue
		}
		players = append(players, addr)
	}
	return players
}

// func (g *Game) getNextReadyPlayer(nextPos int) string {
// 	if nextPos >= len(g.playersList) {
// 		return ""
// 	}
// 	nextPlayer := g.playersList[nextPos]

// 	if g.playersReady.haveRecv(nextPlayer) {
// 		return nextPlayer
// 	}

// 	return g.getNextReadyPlayer(nextPos + 1)
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
