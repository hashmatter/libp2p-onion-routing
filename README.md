# ipfs-onion

Please bear in mind that this is a **highly experimental research project**.
	
*ipfs-onion* is an experiment to bring [in-DHT onion routing](https://github.com/gpestana/notes/blob/master/research/metadata_resistant_dht/onion_routing_paper/onion_routing_dht.pdf) to
IPFS. 

The main goal of the project is to lay the foundations for what could be a
string privacy preserving routing protocol over IPFS that uses in-DHT onion routing, 
so that a local passive adversary cannot link network requests back to the 
initiator.

`ipfs-onion` library consists of primitives for clients to wrap kad requests in
multiple encryption layers and for servers (relays) to decrypt,
verify and forward onion routing packets. It defines a new multistream protocol
`/in-dht-onion/1.0` and handles relay register and discovery on IPFS.

## Protocol

Peers using ipfs-routing encode streams using the multistream protocol 
`/in-dht-onion/1.0`.

The onion header and packet construction uses an [adapted](https://eprint.iacr.org/2009/628.pdf) [sphinx packet format](https://cypherpunks.ca/~iang/pubs/Sphinx_Oakland09.pdf).

### Relay registration and discovery

> TBD

### Onion circuit building

> TBD

### API

**1. Relay**

```golang

import (
	relay "github.com/hashmatter/ipfs-onion/relay"
)

// ...
```

**2. Client**

```golang

import (
 client "github.com/hashmatter/ipfs-onion/client"
)

// ...
```


## Further reading

[1] [Privacy preserving lookups with In-DHT Onion Routing](https://github.com/gpestana/notes/blob/master/research/metadata_resistant_dht/onion_routing_paper/onion_routing_dht.pdf/)

[2] [Sphinx: A Compact and Provably Secure Mix Format](https://cypherpunks.ca/~iang/pubs/Sphinx_Oakland09.pdf)

[3] [Using Sphinx to Improve Onion Routing Circuit Construction](https://eprint.iacr.org/2009/628.pdf)

## License

MIT
