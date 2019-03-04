package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	ec "crypto/elliptic"
	"crypto/rand"
	sphinx "github.com/hashmatter/p3lib/sphinx"
	cid "github.com/ipfs/go-cid"
	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-net"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	proto "github.com/libp2p/go-libp2p-protocol"
	mh "github.com/multiformats/go-multihash"
	"log"
	"os"
	"os/signal"
	"time"
)

var protoId = proto.ID("/ipfs-onion/1.0/")

var rendezvousString = "/ipfs-onion/1.0/example01"

var bootstrapPeers = []string{
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
}

type OnionRelay struct {
	host host.Host
	ctx  *sphinx.RelayerCtx
	rdv  string
}

func main() {
	// sets up an onion relay. is a libp2p hosts, part of the IPFS DHT (for
	// service discovery) and with a dedicated ECDSA key pair, which is its
	// identity as a relayer. The relay registers itself as an IPFS provider of a
	// predefined string, so that an initiator can discover them.
	r, pub := newOnionRelayer()
	log.Printf(">> %v \n", r.host.Addrs(), pub.X)
	log.Printf(">> %v \n", pub.X)
	log.Printf(">> %v \n", r.host.ID())

	// keeps connection on until SIGINT (ctrl+c)
	handleExit(r.host)
	select {}
}

func handleOnionPacket(stream inet.Stream) {
	log.Println(">> handling new packet")
	buf := bufio.NewReader(stream)
	b, err := buf.ReadBytes('\n')
	if err != nil {
		log.Fatal(err)
	}
	// do something: verify, decrypt, forward
	log.Println("%v", b)
}

func handleRelayDiscovery(stream inet.Stream) {
	log.Println(">> handling new relay discovery")
}

func newOnionRelayer() (*OnionRelay, ecdsa.PublicKey) {
	// relay is a libp2p host
	ctx := context.Background()
	host, err := libp2p.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// sets handler for incoming onion routing packets sent through protocol
	// /ipfs-onion/1.0/
	host.SetStreamHandler(protoId, handleOnionPacket)

	// sets handler for incoming onion relay discovery requests. the relay will
	// answer to this packets with its ECDSA public key
	host.SetStreamHandler(protoId, handleRelayDiscovery)

	// join the IPFS DHT for peer discovery by creating a kademlia DHT and
	// connecting to IPFS bootstrap nodes
	log.Println(">> inits relayer kad")

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

	// register relay by becoming a provider a pre-agreed string (redevouz point)
	v1b := cid.V1Builder{Codec: cid.Raw, MhType: mh.SHA2_256}
	rendezvousPoint, _ := v1b.Sum([]byte(rendezvousString))

	log.Printf(">> registering as relayer at [%v] (%v)\n",
		rendezvousString, rendezvousPoint)

	tctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := kad.Provide(tctx, rendezvousPoint, true); err != nil {
		panic(err)
	}

	// onion relay identity (pub key) must be detatched from host identity, so
	// create an ephemeral identity for relayer
	log.Printf(">> setting up relayer context\n")

	relayPrivKey, _ := ecdsa.GenerateKey(ec.P256(), rand.Reader)
	relayContext := sphinx.NewRelayerCtx(relayPrivKey)

	return &OnionRelay{
		host: host,
		ctx:  relayContext,
		rdv:  rendezvousString,
	}, relayPrivKey.PublicKey
}

func handleExit(h host.Host) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			log.Println(">> shutting host down....")
			err := h.Close()
			if err != nil {
				log.Println(err)
			}
			log.Println(">> done, bye")
			os.Exit(1)
		}
	}()
}
