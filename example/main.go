package main

import (
	"fmt"

	dht "github.com/nimona/go-nimona-dht"
	fdht "github.com/nimona/go-nimona-fake-dht"
	net "github.com/nimona/go-nimona-net"
	ps "github.com/nimona/go-nimona-peerstore"
)

func main() {
	pb := &ps.BasicPeer{
		ID:        "5611C2D2-9C06-4376-972F-C538385D79D5",
		Addresses: []string{"localhost:60800"},
	}

	p1 := &ps.BasicPeer{
		ID:        "PEER-ONE",
		Addresses: []string{"localhost:60801"},
	}

	p2 := &ps.BasicPeer{
		ID:        "PEER-TWO",
		Addresses: []string{"localhost:60802"},
	}

	newDHT(pb)
	dht1 := newDHT(p1)
	dht2 := newDHT(p2)

	cs, _ := dht1.FindPeer(p2.ID)
	fmt.Println("Found p2 from p1", <-cs)

	cs, _ = dht2.FindPeer(p1.ID)
	fmt.Println("Found p1 from p2", <-cs)

}

func newDHT(peer *ps.BasicPeer) dht.DHT {
	peerstore := ps.New()
	network, _ := net.NewTCPNetwork(peer, peerstore)
	fakedht := fdht.NewFakeDHT(network, peerstore, peer)
	network.HandleStream(fdht.ProtocolID, fakedht.HandleStream)
	return fakedht
}
