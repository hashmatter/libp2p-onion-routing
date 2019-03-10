## libp2p-onion-routing

`libp2p-onion-routing` demonstrates how to use onion routing for a
strong privacy preserving routing protocol to be used over DHTs and other
decentralized networks. The onion routing aims at protecting users from 
local passive adversaries that spoof DHT requests to link lookups to its initiators.

![usage-example](https://media.giphy.com/media/xT0xetJEkloDtbVHSU/giphy.gif)

Check the [recorded presentation](https://media.giphy.com/media/xT0xetJEkloDtbVHSU/giphy.gif) 
where we go through the code and explain how use delegated lookups in 
libp2p with onion routing and [p3lib-sphinx](https://github.com/hashmatter/p3lib).

### How?

The code in this repo shows how a DHT client can use the library
[p3lib-sphinx](https://github.com/hashmatter/p3lib) to construct a onion packet 
which and forward it through a secure circuit of libp2p nodes.

The circuit nodes (relays) are able to process
the received by "peeling" a layer of the onion and forward the packet to the
next relay. Once the packet has been all processed and forwarded by all relays 
in the circuit, the last relay will have enough information to perform the DHT 
request delegated by the initiator (i.e. the initial node which created the 
onion packet). The primitives for relays to process onion packets are also
implemented by [p3lib-sphinx](https://github.com/hashmatter/p3lib).

This *delegation pattern* in combination with provably secure
onion encryption is similar to what is used by other anonymous P2P networks,
such as [Tor](https://torproject.org).

### Open problems

- [ ] Implementation of SURBs in [p3lib-sphinx](https://github.com/hashmatter/p3lib)

SURBs (Single Use Reply Blocks) are used to allow the exit node to send a
response to the initiator through the established secure path with expected
security properties. The initiator prepares an onion packet for the exit relay
to fill with the results and send back. 
If you'd like to help with the SURB implementation in p3lib-sphinx, check
[p3lib-sphinx](https://github.com/hashmatter/p3lib).

- [ ] Scalable and anonymous relay selection with partial network view

- [ ] Measure and understand entropy requirements for security

There are many more open problems to achieve practical metadata resistance in
distributed hash tables and P2P networks. If you are interested in discussing
and working on these problems, check what [hashmatter](https://hashmatter.com)
has been working on and reach out!

### Please note!

The code in this repository is part of an experimental research project to 
implement practical privacy preserving network protocols and primitives. The
code in this repo is **highly experimental** and for demo purposes only, do not
use it in production.

Please bear in mind that the relay discovery protocol used in this repo will 
defeat the purposes of onion routing security by the identity of the initiator
to all the relays in the circuit. This is not safe and should not be used in
production.

## Further reading

[1] [Privacy preserving lookups with In-DHT Onion Routing](https://github.com/gpestana/notes/blob/master/research/metadata_resistant_dht/onion_routing_paper/onion_routing_dht.pdf/)

[2] [Sphinx: A Compact and Provably Secure Mix Format](https://cypherpunks.ca/~iang/pubs/Sphinx_Oakland09.pdf)

[3] [Using Sphinx to Improve Onion Routing Circuit Construction](https://eprint.iacr.org/2009/628.pdf)

## License

MIT
