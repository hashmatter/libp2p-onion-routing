package main

import (
	"context"
	"crypto/ecdsa"
	ec "crypto/elliptic"
	"crypto/rand"
	//cid "github.com/ipfs/go-cid"
	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	//inet "github.com/libp2p/go-libp2p-net"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	//mh "github.com/multiformats/go-multihash"
	"log"
)

var bootstrapPeers = []string{
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
}

func newOnionClient() (host.Host, ecdsa.PrivateKey, context.Context, *dht.IpfsDHT) {
	ctx := context.Background()
	host, err := libp2p.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// join the IPFS DHT for peer discovery by creating a kademlia DHT and
	// connecting to IPFS bootstrap nodes
	log.Println(">> inits client kad")

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
