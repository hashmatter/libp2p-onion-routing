package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	ec "crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	cid "github.com/ipfs/go-cid"
	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	proto "github.com/libp2p/go-libp2p-protocol"
	mh "github.com/multiformats/go-multihash"

	"github.com/hashmatter/p3lib/sphinx"
)

// rendezvous point for relay to register as provider
var rendezvousString = "/ipfs-onion/1.0/exampleAB"

var protoDiscovery = proto.ID("/ipfs-onion/1.0/discovery")
var protoPacket = proto.ID("/ipfs-onion/1.0/packet")

type OnionRelay struct {
	host host.Host
	ctx  *sphinx.RelayerCtx
	rdv  string
}

func main() {
	// sets up an onion relay. a relay is a libp2p host, part of the IPFS DHT (for
	// service discovery) and with a dedicated ECDSA key pair. the ECDSA key pair
	// is its identity as a relayer and it should be different from its host
	// identity. the relay "registers" itself as a relay by adding itself as a
	// provider of the predefined rendezvous point in IPFS.
	r, pub := newOnionRelayer()

	log.Printf(">> %v\n%v\n", r.host.ID().Pretty(), pub)

	// keeps connection on until SIGINT (ctrl+c)
	handleExit(r.host)
	select {}
}

func newOnionRelayer() (*OnionRelay, ecdsa.PublicKey) {
	// relay is a libp2p host
	ctx := context.Background()
	host, err := libp2p.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// join the IPFS DHT for peer discovery by creating a kademlia DHT and
	// connecting to IPFS bootstrap nodes
	log.Println(">> INIT | relayer kad")

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

	log.Printf(">> REGISTER | relayer at [%v] (%v)\n",
		rendezvousString, rendezvousPoint)

	tctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := kad.Provide(tctx, rendezvousPoint, true); err != nil {
		panic(err)
	}

	// onion relay identity (pub key) must be detatched from host identity, so
	// create an ephemeral identity for relayer
	log.Printf(">> SETTING UP | relayer context\n")

	relayPrivKey, _ := ecdsa.GenerateKey(ec.P256(), rand.Reader)
	relayContext := sphinx.NewRelayerCtx(relayPrivKey)

	// sets handler for incoming onion routing packets sent through protocol
	// /ipfs-onion/1.0/
	host.SetStreamHandler(protoDiscovery, func(stream inet.Stream) {
		handleDiscovery(relayPrivKey.PublicKey, stream)
	})

	// handles a new onion packet
	host.SetStreamHandler(protoPacket, func(stream inet.Stream) {
		handlePacket(ctx, relayContext, kad, host, stream)
	})

	// sets handler for incoming onion relay discovery requests. the relay will
	// answer to this packets with its ECDSA public key

	return &OnionRelay{
		host: host,
		ctx:  relayContext,
		rdv:  rendezvousString,
	}, relayPrivKey.PublicKey
}

// handles a new onion packet stream byte
func handlePacket(ctx context.Context, relayContext *sphinx.RelayerCtx,
	kad *dht.IpfsDHT, host host.Host, stream inet.Stream) {

	out, err := ioutil.ReadAll(stream)
	if err != nil {
		log.Fatal(err)
	}

	var packet sphinx.Packet
	r := bytes.NewReader(out)

	dec := gob.NewDecoder(r)
	err = dec.Decode(&packet)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("RECEIVED | onion packet")
	log.Println(packet)

	na, nextPacket, err := relayContext.ProcessPacket(&packet)
	if err != nil {
		log.Fatal("ERR | process packet: ", err)
	}

	log.Println("\nPROCESSED | onion packet")
	log.Println(nextPacket)

	// checks if it is exit relayer
	if nextPacket.IsLast() {
		if err != nil {
			log.Println("ERR | ", err)
			log.Println(na)
		}
		log.Println("LAST PACKET | exit relayer information: \n")
		log.Println("Payload: ", string(nextPacket.Payload[:]))
		return
	}

	// get peerinfo based on address host.FindPeer()
	nextPid, err := nextRelayID(na[:])
	if err != nil {
		log.Fatal(err)
	}

	log.Println("NEXT RELAY | find peer", nextPid)
	pinfo, err := kad.FindPeer(ctx, nextPid)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("NEXT RELAY | connect", nextPid.Pretty())
	err = host.Connect(ctx, pinfo)
	if err != nil {
		log.Println(nextPid)
		log.Fatal("connect to host: ", err)
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(nextPacket)
	if err != nil {
		log.Fatal("encode packet ", err)
		log.Fatal(err)
	}

	// forwasds packet to next relay.
	log.Println("FORWARD | next relay", nextPid)
	stream, err = host.NewStream(context.Background(), nextPid, protoPacket)
	if err != nil {
		log.Fatal("create out stream ", err)
		log.Fatalln(err)
	}
	_, err = stream.Write(buf.Bytes())
	if err != nil {
		log.Fatal("write to stream ", err)
		log.Fatal(err)
	}
	stream.Close()
	log.Println("FORWARD | successful")
}

func nextRelayID(na []byte) (peer.ID, error) {
	np, err := mh.Decode(na[:36])
	if err != nil {
		return "", err
	}

	npid, err := mh.Cast(np.Digest[:])
	if err != nil {
		return "", err
	}

	nextPid, err := peer.IDB58Decode(npid.B58String())
	if err != nil {
		return "", err
	}
	return nextPid, nil
}

func handleDiscovery(pk ecdsa.PublicKey, stream inet.Stream) {
	log.Println(">> DISCOVER | request")
	encPubkey := ec.Marshal(pk.Curve, pk.X, pk.Y)
	_, err := stream.Write(encPubkey)
	if err != nil {
		log.Println(err)
	}
	stream.Close()
}

func handleExit(h host.Host) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			log.Println(">> SHUTTING down....")
			err := h.Close()
			if err != nil {
				log.Println(err)
			}
			log.Println(">> done, bye")
			os.Exit(1)
		}
	}()
}

var bootstrapPeers = []string{
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
}
