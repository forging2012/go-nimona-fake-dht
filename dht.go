package dht

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	uuid "github.com/pborman/uuid"

	net "github.com/nimona/go-nimona-net"
	ps "github.com/nimona/go-nimona-peerstore"
)

const (
	// ProtocolID -
	ProtocolID = "/fake-dht/v0"
	// TypeFindPeer -
	TypeFindPeer = "FIND_PEER"
	// TypePing -
	TypePing = "PING"

	// FakeBootstrapPeerID -
	FakeBootstrapPeerID = "5611C2D2-9C06-4376-972F-C538385D79D5"
	//FakeBootstrapAddress -
	FakeBootstrapAddress = "bootstrap.nimona.io:60800"
)

// NewFakeDHT -
func NewFakeDHT(net net.Network, peerstore ps.Peerstore, peer ps.Peer) *FakeDHT {
	dht := &FakeDHT{
		net: net,
		ps:  peerstore,
		pcs: map[string]chan ps.Peer{},
		bp: &ps.BasicPeer{
			ID:        ps.ID(FakeBootstrapPeerID),
			Addresses: []string{FakeBootstrapAddress},
		},
		peer: peer.(*ps.BasicPeer),
	}
	peerstore.Put(dht.bp)
	net.HandleStream(ProtocolID, dht.HandleStream)
	dht.announce()
	return dht
}

// FakeDHT -
type FakeDHT struct {
	net     net.Network
	ps      ps.Peerstore
	pcs     map[string]chan ps.Peer
	bp      *ps.BasicPeer // TODO use ps.Peer
	peer    *ps.BasicPeer // TODO use ps.Peer
	Verbose bool          // TODO replace with proper logger
}

// FakeDHTMessage -
type FakeDHTMessage struct {
	Type       string        `json:"t"`
	Nonce      string        `json:"n"`
	Originator *ps.BasicPeer `json:"o"`
	PeerID     ps.ID         `json:"f"`
	Peer       *ps.BasicPeer `json:"p"`
	Response   bool          `json:"r"`
}

// HandleStream -
func (dht *FakeDHT) HandleStream(protocol string, rwc io.ReadWriteCloser) error {
	r := bufio.NewReader(rwc)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			dht.log("Could not read string", err)
			continue
		}

		dht.log("Got message", line)

		message := &FakeDHTMessage{}
		if err := json.Unmarshal([]byte(line), &message); err != nil {
			dht.log("Could not unmarshal message", err)
			continue
		}

		if message.Nonce == "" {
			dht.log("Missing nonce")
			continue
		}

		// when a peers sends us any message we must add them to our
		// peerstore so we can actually get back to them, and also allow
		// others to look for them.
		// make sure we don't overwrite the bootstrap peer
		if message.Originator.ID != FakeBootstrapPeerID {
			// TODO we should check the peer first somehow
			dht.ps.Put(message.Originator)
		}

		switch message.Type {
		case TypeFindPeer:
			if message.Response {
				if pc, ok := dht.pcs[message.Nonce]; ok {
					// TODO check the peer first somehow
					// make sure we don't overwrite the bootstrap peer
					if message.Peer.ID != FakeBootstrapPeerID {
						dht.ps.Put(message.Peer)
					}
					if message.PeerID == message.Peer.ID {
						pc <- message.Peer
						close(pc) // TODO Erm? Can we close it immediately?
					}
				}
				break
			}

			peer, err := dht.ps.Get(message.PeerID)
			if err != nil {
				// we don't know the peer
				// TODO check for other errors as well
				break
			}

			message.Peer = peer.(*ps.BasicPeer) // TODO check cast
			message.Response = true

			dht.sendMessage(message.Originator.ID, message)

		case TypePing:
		// for now pings are only used to annouce peers
		// so there is nothing to do here

		default:
			dht.log("Missing type")
		}
	}
}

func (dht *FakeDHT) findPeer(id ps.ID, pc chan ps.Peer) {
	// ask the bootstrap node
	message := &FakeDHTMessage{
		Type:       TypeFindPeer,
		Nonce:      uuid.New(),
		Originator: dht.peer,
		PeerID:     id,
	}
	dht.pcs[message.Nonce] = pc
	dht.sendMessage(dht.bp.GetID(), message)
}

func (dht *FakeDHT) announce() {
	// announce our local peer to the bootstrap peer
	message := &FakeDHTMessage{
		Type:       TypePing,
		Nonce:      uuid.New(),
		Originator: dht.peer,
	}
	dht.sendMessage(dht.bp.GetID(), message)
}

func (dht *FakeDHT) sendMessage(id ps.ID, message *FakeDHTMessage) {
	stream, err := dht.net.NewStream(ProtocolID, string(id))
	if err != nil {
		dht.log("Could not get stream", id, err)
		return
	}

	bs, _ := json.Marshal(message)
	dht.log("Sending message", string(bs))
	bs = append(bs, '\n')
	if _, err := stream.Write(bs); err != nil {
		dht.log("Could not write to stream", err)
	}
	if err := stream.Close(); err != nil {
		dht.log("Could not close stream", err)
	}
}

func (dht *FakeDHT) log(a ...interface{}) {
	if dht.Verbose {
		b := append([]interface{}{">", ProtocolID, ">"}, a)
		fmt.Println(b...)
	}
}

// FindPeer -
func (dht *FakeDHT) FindPeer(id ps.ID) (chan ps.Peer, error) {
	cs := make(chan ps.Peer, 1)
	dht.findPeer(id, cs)
	return cs, nil
}
