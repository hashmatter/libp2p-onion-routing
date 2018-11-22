# ipfs-onion

Please bear in mind that ipfs-onion is a **highly experimental project**.
	
ipfs-onion is experiment on using [in-DHT onion routing](https://github.com/gpestana/notes/blob/master/research/metadata_resistant_dht/onion_routing_paper/onion_routing_dht.pdf) on
IPFS. The main goal of the project is to lie the foundations for what could be 
privacy preserving routing protocol over IPFS that uses in-DHT onion routing, 
so that a local, passive adversary cannot link a network request back to its 
initiator.

ipfs-onion consists of primitives for clients to wrap kad requests in
multiple encryption layers and primitives for servers (relays) to decrypt,
verify and forward onion routing packets. It also defines a multistream protocol
`/in-dht-onion/*` for peers to agree on the protocol being used.

## Protocol

Peers using ipfs-routing encode streams using the multistream protocol 
`/in-dht-onion/1.0`.

The onion header and packet construction uses an [adapted](https://eprint.iacr.org/2009/628.pdf) [sphinx packet format](https://cypherpunks.ca/~iang/pubs/Sphinx_Oakland09.pdf).

### Relay peer discovery

> TBD

### Onion circuit building

> TBD

### API

```golang

import (
	io "github.com/hashmatter/ipfs-onion"
)

relayDirLocation := "QmSAR9Zw6bvVqMt35uBfnETaWkmxhZ6mWyQeRK8Eem721p" // for service discovery. in this location there is a directory list with onion relays and their PKs

path, _ := io.BuildCircuit(relayDirLocation) // fetches list of available relays and builds the onion path

packet, header, next, _ := io.BuildCircuit(circuit, data)

_ := io.send(packet, header, next)
```

## Further reading

[1] [Privacy preserving lookups with In-DHT Onion Routing](https://github.com/gpestana/notes/blob/master/research/metadata_resistant_dht/onion_routing_paper/onion_routing_dht.pdf/) on IPFS.

[2] [Sphinx: A Compact and Provably Secure Mix Format](https://cypherpunks.ca/~iang/pubs/Sphinx_Oakland09.pdf)

[3] [Using Sphinx to Improve Onion Routing Circuit Construction](https://eprint.iacr.org/2009/628.pdf)

## License

MIT
