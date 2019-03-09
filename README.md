## libp2p-onion-routing

This is an example of how to use [p3lib-sphinx](https://github.com/hashmatter/p3lib)
to implement an onion routing protocol on top of a DHT overlay. This example
shows how a client can construct a onion packet which will be forward through 
a secure circuit of libp2p nodes. The circuit nodes (relays) are able to process
the received by "peeling" a layer of the onion and forward the packet to the
next relay. Once the packet has been all processed and forwarded by all relays 
in the circuit, the last relay will have enough information to perform the DHT 
request delegated by the initiator (i.e. the initial node which created the 
onion packet). This *delegation pattern* in combination with provably secure
onion encryption is 

The main goal of the project is to lay the foundations for what could be a
strong privacy preserving routing protocol over DHTs to protect users from 
local passive adversaries that aim at linking DHT lookups and its initiator.

The code in this repository is part of a **highly experimental research project**
to bring [onion routing](https://github.com/gpestana/notes/blob/master/research/metadata_resistant_dht/onion_routing_paper/onion_routing_dht.pdf) to be used on top of IPFS and other P2P networks
The purpose of this code is for demo purposes only. Please bear in mind that the relay discovery
protocol used in the example will defeat the purposes of onion routing security
by leaking to the network which relays are used 

## Further reading

[1] [Privacy preserving lookups with In-DHT Onion Routing](https://github.com/gpestana/notes/blob/master/research/metadata_resistant_dht/onion_routing_paper/onion_routing_dht.pdf/)

[2] [Sphinx: A Compact and Provably Secure Mix Format](https://cypherpunks.ca/~iang/pubs/Sphinx_Oakland09.pdf)

[3] [Using Sphinx to Improve Onion Routing Circuit Construction](https://eprint.iacr.org/2009/628.pdf)

## License

MIT
