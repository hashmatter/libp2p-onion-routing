package main

import (
	"context"
	io "github.com/hashmatter/ipfs-onion"
	_ "github.com/hashmatter/ipfs-onion/client"
	relay "github.com/hashmatter/ipfs-onion/relay"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-net"
	"log"
)

var rendezvousPoint = "/ipfs-onion/1.0/relay_example"
var bootstrapPeers = []string{
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
}

func handleStream(stream inet.Stream) {
	log.Println("handling new stream..")
}

func main() {
	// creates relay as libp2p host
	ctxRel := context.Background()
	peer, err := libp2p.New(ctxRel)
	if err != nil {
		log.Fatal(err)
	}

	peer.SetStreamHandler(io.STREAM_HANDLER, handleStream)
	kad, err := dht.New(ctxRel, peer)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("bootstrapping relay..")
	err = bootstrapDht(peer)
	if err != nil {
		log.Fatal(err)
	}

	rel := relay.New(&peer, kad, rendezvousPoint)
	log.Println("announcing relay..")

	err = rel.Register()
	if err != nil {
		log.Fatal(err)
	}
}

func bootstrapDht(peer host.Host) error {
	return nil
}
