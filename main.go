package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/gob"
	"github.com/hashmatter/p3lib/sphinx"
	cid "github.com/ipfs/go-cid"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	proto "github.com/libp2p/go-libp2p-protocol"
	ma "github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
	"log"
	"sync"
)

var rendezvousString = "/ipfs-onion/1.0/example01"
var protoId = proto.ID("/ipfs-onion/1.0/")

func main() {
	// sets up initiator peer. an initiator is libp2p host connected to the IPFS
	// DHT with a dedicated identity (ECDSA key pair) to be used in the onion
	// routing.
	host, privKey, ctx, kad := newOnionClient()

	numRelays := 2
	var relayAddrs []ma.Multiaddr
	var relayPubKeys []ecdsa.PublicKey

	var wg sync.WaitGroup
	wg.Add(numRelays)

	// discovers and connects to numRelays relays
	log.Println(">> discovering relays")
	v1b := cid.V1Builder{Codec: cid.Raw, MhType: mh.SHA2_256}
	rendezvousPoint, _ := v1b.Sum([]byte(rendezvousString))
	pinfos, err := kad.FindProviders(ctx, rendezvousPoint)
	if err != nil {
		log.Fatal(err)
	}

	for _, pi := range pinfos {
		// tries to connect to relays
		go func(pi pstore.PeerInfo) {
			err := host.Connect(ctx, pi)
			// connection was successfull
			if err == nil {
				defer wg.Done()
				log.Printf(">> connected to %v", pi)
			} else {
				log.Println(err)
			}
		}(pi)

	}
	wg.Wait()

	// got all peers

	// TODO: change for libp2p host
	finalAddr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/udp/1234")

	// initializes onion packet payload
	var payload [256]byte
	copy(payload[:], []byte("example onion routing")[:])

	// builds onion packet
	packet, err :=
		sphinx.NewPacket(&privKey, relayPubKeys, finalAddr, relayAddrs, payload)
	if err != nil {
		log.Fatal(err)
	}

	// encodes onion packet and wires it to the first relay in the circuit
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(packet)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(host)
	log.Println(ctx)
	log.Println(packet)
}
