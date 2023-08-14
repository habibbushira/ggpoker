package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
)

type GameVariant uint8

func (gv GameVariant) String() string {
	switch gv {
	case TexasHoldem:
		return "TEXAS HOLDEM"
	case Other:
		return "OTHER"
	default:
		return "UNKNOWN"
	}
}

const (
	TexasHoldem GameVariant = iota
	Other
)

type ServerConfig struct {
	ListenAddr  string
	Version     string
	GameVariant GameVariant
}

type Server struct {
	ServerConfig

	transport   *TCPTransport
	peerLock    sync.RWMutex
	mu          sync.RWMutex
	peers       map[net.Addr]*Peer
	addPeer     chan *Peer
	delPeer     chan *Peer
	msgCh       chan *Message
	broadcastch chan any
	gameState   *GameState
}

func NewServer(cfg ServerConfig) *Server {
	s := &Server{
		ServerConfig: cfg,
		peers:        make(map[net.Addr]*Peer),
		addPeer:      make(chan *Peer, 10),
		delPeer:      make(chan *Peer),
		msgCh:        make(chan *Message),
		broadcastch:  make(chan any),
	}

	s.gameState = NewGameState(s.ListenAddr, s.broadcastch)

	if s.ListenAddr == ":3000" {
		s.gameState.isDealer = true
	}

	tr := NewTCPTransport(s.ListenAddr)
	s.transport = tr

	tr.AddPeer = s.addPeer
	tr.DelPeer = s.delPeer

	return s
}

func (s *Server) Start() {
	go s.loop()

	fmt.Printf("started new game: port %s variant %v\n", s.ListenAddr, s.GameVariant)

	go func() {
		msg := <-s.broadcastch
		fmt.Println("broadcasting to all peers")
		if err := s.Broadcast(msg); err != nil {
			fmt.Println("broadcast error ", err)
		}
	}()
	s.transport.ListenAndAccept()

}

func (s *Server) sendPeerList(p *Peer) error {
	peerList := MessagePeerList{
		Peers: []string{},
	}

	peers := s.Peers()

	for i := 0; i < len(peers); i++ {
		if peers[i] != p.listenAddr {
			peerList.Peers = append(peerList.Peers, peers[i])
		}
	}

	// for _, peer := range s.peers {
	// 	peerList.Peers = append(peerList.Peers, peer.listenAddr)
	// }

	if len(peerList.Peers) == 0 {
		return nil
	}

	msg := NewMessage(s.ListenAddr, peerList)

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	return p.Send(buf.Bytes())
}

func (s *Server) AddPeer(p *Peer) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.conn.RemoteAddr()] = p
}

func (s *Server) Peers() []string {
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()

	peers := make([]string, len(s.peers))
	it := 0
	for _, peer := range s.peers {
		peers[it] = peer.listenAddr
		it++
	}

	return peers

}

func (s *Server) SendHandshake(p *Peer) error {
	hs := &Handshake{
		GameVariant: s.GameVariant,
		Version:     s.Version,
		GameStatus:  s.gameState.gameStatus,
		ListenAddr:  s.ListenAddr,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hs); err != nil {
		return err
	}

	return p.Send(buf.Bytes())
}

func (s *Server) isInPeerList(addr string) bool {
	peers := s.Peers()
	for i := 0; i < len(peers); i++ {
		if peers[i] == addr {
			return true
		}
	}

	return false
}

func (s *Server) Connect(addr string) error {
	if s.isInPeerList(addr) {
		return nil
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	peer := &Peer{
		conn:     conn,
		outbound: true,
	}

	s.addPeer <- peer

	return s.SendHandshake(peer)
}

func (s *Server) loop() {
	for {
		select {
		case peer := <-s.addPeer:
			if err := s.handleNewPeer(peer); err != nil {
				fmt.Printf("handle peer error: %s", err)
			}
		case peer := <-s.delPeer:
			peer.conn.Close()
			delete(s.peers, peer.conn.RemoteAddr())
			fmt.Printf("player disconnected %s\n", peer.conn.RemoteAddr())

		case msg := <-s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				fmt.Println("handle msg error ", err)
			}
		}
	}
}

func (s *Server) handleNewPeer(peer *Peer) error {
	hs, err := s.handshake(peer)
	if err != nil {
		peer.conn.Close()
		delete(s.peers, peer.conn.RemoteAddr())
		return fmt.Errorf("handshake with incoming player failed: %s\n", err)
	}

	// NOTE: this readloop always needs to start after the handshake
	go peer.ReadLoop(s.msgCh)

	if !peer.outbound {
		if err := s.SendHandshake(peer); err != nil {
			peer.conn.Close()
			delete(s.peers, peer.conn.RemoteAddr())
			return fmt.Errorf("failed to send handshake with peer: %s", err)
		}

		go func() {
			if err := s.sendPeerList(peer); err != nil {
				fmt.Printf("peerlist error: %s\n", err)
			}
		}()
	}

	fmt.Printf("handshake successful: new player connected peer: %s, gamestatus: %s, listenAddr: %s we: %s\n", peer.conn.RemoteAddr(), hs.GameStatus, peer.listenAddr, s.ListenAddr)

	// s.peers[peer.conn.RemoteAddr()] = peer

	s.AddPeer(peer)

	s.gameState.AddPlayer(peer.listenAddr, hs.GameStatus)

	return nil
}

func (s *Server) Broadcast(payload any) error {
	msg := NewMessage(s.ListenAddr, payload)

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		go func(peer *Peer) {
			if err := peer.Send(buf.Bytes()); err != nil {
				fmt.Println("broadcast to peer error: ", err)
			}

			fmt.Printf("broadcast we: %s, peer: %s", s.ListenAddr, peer.listenAddr)
		}(peer)
	}

	return nil
}

func (s *Server) handshake(p *Peer) (*Handshake, error) {
	hs := &Handshake{}
	if err := gob.NewDecoder(p.conn).Decode(hs); err != nil {
		return nil, err
	}

	if s.GameVariant != hs.GameVariant {
		return nil, fmt.Errorf("gamevariant doesn't match %s", hs.GameVariant)
	}
	if s.Version != hs.Version {
		return nil, fmt.Errorf("version doesn't match %s", hs.Version)
	}

	p.listenAddr = hs.ListenAddr

	return hs, nil
}

func (s *Server) handleMessage(msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessagePeerList:
		return s.handlePeerList(v)
	case MessageCards:
		fmt.Printf("%s: recieved msg from (%s) -> %+v\n", s.ListenAddr, msg.From, v)
	}
	return nil
}

// TODO goroutine
func (s *Server) handlePeerList(l MessagePeerList) error {
	// fmt.Printf("recieved peerlist message we %s list %v \n", s.ListenAddr, l.Peers)
	for i := 0; i < len(l.Peers); i++ {
		if err := s.Connect(l.Peers[i]); err != nil {
			fmt.Println("failed to dial peer: ", err)
			continue
		}
	}

	return nil
}

func init() {
	gob.Register(MessagePeerList{})
	gob.Register(MessageCards{})
}
