package p2p

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

type NetAddr string

func (n NetAddr) String() string  { return string(n) }
func (n NetAddr) Network() string { return "tcp" }

type Peer struct {
	conn       net.Conn
	outbound   bool
	listenAddr string
}

func (p *Peer) Send(b []byte) error {
	_, err := p.conn.Write(b)
	return err
}

func (p *Peer) ReadLoop(msgCh chan *Message) {
	for {
		// new make a pointer; like &Message
		msg := new(Message)
		if err := gob.NewDecoder(p.conn).Decode(msg); err != nil {
			fmt.Printf("decode message error: %s", err)
			break
		}

		msgCh <- msg
	}

	// TODO(@habibbushira): unregister this peer
	p.conn.Close()

}

type TCPTransport struct {
	listenAddr string
	listener   net.Listener
	AddPeer    chan *Peer
	DelPeer    chan *Peer
}

func NewTCPTransport(addr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: addr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	t.listener = ln

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}

		peer := &Peer{
			conn:     conn,
			outbound: false,
		}

		t.AddPeer <- peer
	}
}