# Candango Go Peasant

Peasant Protocol: A Contract for Controlling Agents

The Peasant protocol is a high-level abstraction designed to facilitate
communication between agents (peasants) and central entities (bastions). It
does not impose specific implementation details, security requirements, or
redundancy levels, but instead establishes a minimal contract for what must be
implemented.

In this protocol, agents are referred to as peasants, while central entities
are called bastions. The relationship between a bastion and peasant can be
either stateful or stateless. In a stateful scenario, bastions must implement
a session control mechanism, requiring peasants to perform "knocks" (similar
to knocking on a door) to request permission or establish a valid session. In a
stateless scenario, the concept of knocking is ignored.

The Peasant protocol mandates the implementation of nonce generation,
consumption and validation on both the peasant and bastion sides.
Additionally, bastions must provide a directory list of available resources
that peasants can consume.

## Support

GoPeasat is one of
[Candango Open Source Group](http://www.candango.org/projects/)
initiatives. Available under the
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).

This site and all documentation are licensed under
[Creative Commons 3.0](http://creativecommons.org/licenses/by/3.0/).

Copyright © 2024 Flavio Garcia
