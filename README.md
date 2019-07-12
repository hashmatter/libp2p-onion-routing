## libp2p-onion-routing

`libp2p-onion-routing` demonstrates how to use onion routing for a
strong privacy preserving routing protocol to be used over DHTs and other
decentralized networks. The onion routing aims at protecting users from 
local passive adversaries that spoof DHT requests to link lookups to its initiators.

We will release a video tutorial  
where we go through the code and explain how use delegated lookups in 
libp2p with onion routing and [p3lib-sphinx](https://github.com/hashmatter/p3lib).

### How?

[p3lib-sphinx](https://github.com/hashmatter/p3lib) is used to construct and
process the onion packet.

**1) *Initiator* selects set of relay nodes and gets its public keys and addresses**

```go
// discovers and connects to numRelays relays
pinfo, relayAddrs, relayPubKeys, err := selectRelays(ctx, numRelays, timeout,
host, kad)
```

In production, the `selectRelays()` should select a set of relays anonymously,
i.e., no one should be able to infer which relays were selected.

**2) *Initiator* constructs the onion packet**

```go
// the payload consist of information for the exit relay (last relay in the
// circuit) to perform the a DHT lookup. in this example, the initiator
// delegates the lookup of the content addressed file 'QmSAR...K8Eem721p'
var payload [256]byte
copy(payload[:], []byte("GET QmSAR9Zw6bvVqMt35uBfnETaWkmxhZ6mWyQeRK8Eem721p")[:])

// the final address is not relevant in this context since the exit relay will
// perform a network request (DHT lookup) rather than
// connecting to a specific peer.
finalAddr := []byte("")

// uses p3lib-sphinx library to construct an onion routing packet
packet, err := sphinx.NewPacket(&privKey, relayPubKeys, finalAddr, relayAddrs, payload)
```

**3) *Initiator* encodes and forwards the onion packet to first relay**

```go
// encodes onion packet and wires it to the first relay in the circuit
var buf bytes.Buffer
enc := gob.NewEncoder(&buf)
enc.Encode(packet)

// forward packet to first relay. the relay that the packet is relayed must
// map to the first relay in the input when constructing the onion packet
firstRelay := pinfo[0]
stream, _ := host.NewStream(context.Background(), firstRelay.ID, protoPacket)
_, err = stream.Write(buf.Bytes())
stream.Close()
```

**4) *Relay* receives the onion packet, processes it and gets the address
of next relay**

```go
var packet sphinx.Packet
out, _ := ioutil.ReadAll(stream)

r := bytes.NewReader(out)

dec := gob.NewDecoder(r)
dec.Decode(&packet)

nextAddr, nextPacket, err := relayContext.ProcessPacket(&packet)
```

**5) Step 5 repeats until last relay processes the packet**

**6) After processing the packet, the *last Relay* performs DHT lookup set by
initiator**

```go
if nextPacket.IsLast() {
	log.Println("LAST PACKET | exit relayer information: \n")

	performDelegatedRequest(nextPacket.Payload[:])
	return
}
```

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

### Why?

Because we all want security, privacy and decentralized networks to go 
mainstream! Also check [In Pursuit of Private DHTs](https://www.gpestana.com/blog/in-pursuit-of-private-dhts/)!

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

[0] [Privacy preserving DHTs](https://github.com/gpestana/notes/issues/8)

[1] [Privacy preserving lookups with In-DHT Onion Routing](https://github.com/gpestana/notes/blob/master/research/metadata_resistant_dht/onion_routing_paper/onion_routing_dht.pdf/)

[2] [Sphinx: A Compact and Provably Secure Mix Format](https://cypherpunks.ca/~iang/pubs/Sphinx_Oakland09.pdf)

[3] [Using Sphinx to Improve Onion Routing Circuit Construction](https://eprint.iacr.org/2009/628.pdf)

## License

MIT
