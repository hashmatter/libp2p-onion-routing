package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	ec "crypto/elliptic"
	"encoding/gob"
	"fmt"
	"github.com/hashmatter/p3lib/sphinx"
	cid "github.com/ipfs/go-cid"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	proto "github.com/libp2p/go-libp2p-protocol"
	ma "github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var rendezvousString = "/ipfs-onion/1.0/example02"
var protoId = proto.ID("/ipfs-onion/1.0/")

func main() {
	// sets up initiator peer. an initiator is libp2p host connected to the IPFS
	// DHT with a dedicated identity (ECDSA key pair) to be used in the onion
	// routing.
	host, privKey, ctx, kad := newOnionClient()

	var relayAddrs []ma.Multiaddr
	var relayPubKeys []ecdsa.PublicKey

	numRelays := 3
	timeout := 10

	// discovers and connects to numRelays relays
	log.Printf(">> discovering at least %v relays (timeout in %vs)\n", numRelays, timeout)
	v1b := cid.V1Builder{Codec: cid.Raw, MhType: mh.SHA2_256}
	rendezvousPoint, _ := v1b.Sum([]byte(rendezvousString))

	c := make(chan []pstore.PeerInfo)
	var pis []pstore.PeerInfo

	// relay discovery
	go func(pis []pstore.PeerInfo) {
		for pi := range kad.FindProvidersAsync(ctx, rendezvousPoint, 50) {
			err := host.Connect(ctx, pi)
			if err == nil {
				// connection was successfull
				log.Printf(">> connected to relay %v", pi.ID)
				pis = append(pis, pi)
				if len(pis) >= numRelays {
					c <- pis
				}
			} else {
				fmt.Printf(".")
			}
		}
	}(pis)

	select {
	case relays := <-c:
		log.Printf(">> getting relays pubkeys")
		// get relays pubkey
		for _, r := range relays {
			// builds relayAddrs with same sorting as relayPubKeys
			relayAddrs = append(relayAddrs, selectAddr(r.Addrs))

			stream, err := host.NewStream(context.Background(), r.ID, protoId)
			if err != nil {
				log.Fatalln(err)
			}
			out, err := ioutil.ReadAll(stream)
			if err != nil {
				log.Fatalln(err)
			}
			curve := ec.P256()
			x, y := ec.Unmarshal(curve, out)
			pubKey := ecdsa.PublicKey{Curve: curve, X: x, Y: y}
			relayPubKeys = append(relayPubKeys, pubKey)
			fmt.Printf("pubkey parsed: %v\n", pubKey)
		}

	case <-time.After(time.Second * time.Duration(timeout)):
		log.Printf(">> timeout. did not find enough relays, exiting\n")
		os.Exit(1)
	}

	// TODO: change for libp2p host
	finalAddr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/udp/1234")

	fmt.Println("---- BUILD ONION PACKET with:")
	fmt.Println(relayAddrs)
	fmt.Println(relayPubKeys)

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

	// forward packet to first relay:
	log.Println(packet)
}

// super hack for selecting a multiaddress that is either ip4 or ip6 (and not
// /p2p-circuit (NOT SAFE)
func selectAddr(addrs []ma.Multiaddr) ma.Multiaddr {
	var res ma.Multiaddr
	for _, a := range addrs {
		if a.String()[1] == "i"[0] {
			res = a
			break
		}
	}
	return res
}
