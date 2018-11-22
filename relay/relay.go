package relay

import (
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	_ "github.com/lightningnetwork/lightning-onion"
)

type Relayer interface {
	// registers relay in the network
	Register(string) error

	// decrypts packet received
	Decrypt() //TODO: signature

	// forwards new packet to next hop
	Forward() //TODO: signature

	// exit relay: performs request
	DoRequest() //TODO: signature

	// verifies packet and header receiced
	verifyHeader() // TODO: signature
}

type Relay struct {
	host            *host.Host
	kad             *dht.IpfsDHT
	rendezvousPoint string
}

func New(host *host.Host, kad *dht.IpfsDHT, rendezvousPoint string) *Relay {
	return &Relay{
		host,
		kad,
		rendezvousPoint,
	}
}

// implements register usign a rendezvous point in the DHT
func (*Relay) Register() error {
	return nil
}
