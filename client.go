package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	ec "crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	cid "github.com/ipfs/go-cid"
	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	proto "github.com/libp2p/go-libp2p-protocol"
	mh "github.com/multiformats/go-multihash"

	"github.com/hashmatter/p3lib/sphinx"
)

var rendezvousString = "/ipfs-onion/1.0/exampleAB"

var protoDiscovery = proto.ID("/ipfs-onion/1.0/discovery")
var protoPacket = proto.ID("/ipfs-onion/1.0/packet")

var bootstrapPeers = []string{
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
}

func main() {
	// sets up initiator peer. an initiator is libp2p host connected to the IPFS
	// DHT with a dedicated identity (ECDSA key pair) to be used in the onion
	// routing.
	host, privKey, ctx, kad := newOnionClient()

	var relayAddrs [][]byte
	var relayPubKeys []ecdsa.PublicKey

	numRelays := 3
	timeout := 30

	// discovers and connects to numRelays relays
	log.Printf(">> DISCOVER | threshold: %v relays; timeout in %vs\n", numRelays, timeout)
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
				log.Printf(">> CONNECTED | %v", pi.ID)
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
		// get relays pubkey
		for _, r := range relays {
			// builds relayAddrs with same sorting as relayPubKeys
			encPeerID, err := mh.Encode([]byte(r.ID), mh.SHA2_256)
			relayAddrs = append(relayAddrs, encPeerID)
			pis = relays

			stream, err := host.NewStream(context.Background(), r.ID, protoDiscovery)
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

			fmt.Printf(">> RELAY | id:  %v\n pubKey %v\n\n", r.ID, pubKey)
		}

	case <-time.After(time.Second * time.Duration(timeout)):
		fmt.Println("")
		log.Printf(">> TIMEOUT | did not find enough relays, exiting\n")
		os.Exit(1)
	}

	// TODO: change for online libp2p host
	finalAddr := []byte("QmPa4rVE5QoMyuxPEtq2tS1cjBAnpenFkb3rYTNkZsRRVz")

	log.Println(">> CONSTRUCTING onion packet..")
	// initializes onion packet payload
	var payload [256]byte
	copy(payload[:], []byte("example onion routing")[:])

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
	fr := pis[0]
	stream, err := host.NewStream(context.Background(), fr.ID, protoPacket)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = stream.Write(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	stream.Close()

	log.Println(">> SENT | onion packet sent!")
	fmt.Println(packet)
}

func newOnionClient() (host.Host, ecdsa.PrivateKey, context.Context, *dht.IpfsDHT) {
	ctx := context.Background()
	host, err := libp2p.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// join the IPFS DHT for peer discovery by creating a kademlia DHT and
	// connecting to IPFS bootstrap nodes
	log.Println(">> INIT | client kad")

	kad, err := dht.New(ctx, host)
	if err != nil {
		panic(err)
	}

	for _, peerAddr := range bootstrapPeers {
		pAddr, _ := ipfsaddr.ParseString(peerAddr)
		peerinfo, _ := pstore.InfoFromP2pAddr(pAddr.Multiaddr())

		if err = host.Connect(ctx, *peerinfo); err != nil {
			log.Println("ERROR: ", err)
		}
	}

	clientPrivKey, _ := ecdsa.GenerateKey(ec.P256(), rand.Reader)
	return host, *clientPrivKey, ctx, kad
}
