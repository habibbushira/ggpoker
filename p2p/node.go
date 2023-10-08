package p2p

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/habibbushira/ggpoker/proto"
	"google.golang.org/grpc"
)

type Node struct {
	ServerConfig

	peerLock sync.RWMutex
	peers    map[proto.GossipClient]*proto.Version

	proto.UnimplementedGossipServer
}

func NewNode(cfg ServerConfig) *Node {
	return &Node{
		ServerConfig: cfg,
		peers:        make(map[proto.GossipClient]*proto.Version),
	}
}

func (n *Node) Handshake(ctx context.Context, version *proto.Version) (*proto.Version, error) {
	client, err := makeGrpcClientConn(version.ListenAddr)
	if err != nil {
		return nil, err
	}

	n.addPeer(client, version)

	return n.getVersion(), nil
}

func (n *Node) addPeer(c proto.GossipClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	n.peers[c] = v

	fmt.Println(v.PeerList)

	fmt.Printf("new player connected: we=%s, remote=%s\n", n.ListenAddr, v.ListenAddr)
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    "GGPOKER-0.0.1",
		ListenAddr: n.ListenAddr,
		PeerList:   n.getPeerList(),
	}
}

func (n *Node) getPeerList() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()

	var (
		list = make([]string, len(n.peers))
		i    = 0
	)

	for _, v := range n.peers {
		list[i] = v.ListenAddr
	}

	return list
}

func (n *Node) Connect(addr string) error {
	client, err := makeGrpcClientConn(addr)
	if err != nil {
		return err
	}

	hs, err := client.Handshake(context.TODO(), n.getVersion())
	if err != nil {
		return nil
	}

	n.addPeer(client, hs)
	return nil

}

func (n *Node) Start() error {
	grpcServer := grpc.NewServer()
	proto.RegisterGossipServer(grpcServer, n)

	ln, err := net.Listen("tcp", n.ListenAddr)
	if err != nil {
		return err
	}

	fmt.Printf("started new poker game server: port=%s, variant=%s, maxPlayers=%d\n", n.ListenAddr, n.GameVariant, n.MaxPlayers)

	return grpcServer.Serve(ln)
}

func makeGrpcClientConn(addr string) (proto.GossipClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := proto.NewGossipClient(conn)
	return client, nil
}
