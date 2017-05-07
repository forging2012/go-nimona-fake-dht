# Fake DHT

This is a fake discovery service for our network.  
Is it just a quick hack until something better comes along and __SHOULD
NOT BE USED__. Ever. 

It has a hardcoded bootstrap peer to which all other fake dhts will connect,
announce themselves and query for peers. There is no query propagation to
other peers and peer information cannot leak from the bootstrap node as it will
only respond to exact peer id requests unline most other dhts.

It will also not verify that the peer id or address is valid, nor that it
belongs to the peer that announced it. Basically we do not care about security
at all.