# DESIGN

This document describes the design of gcache server. It is a couple of notes about the organization of the code which
can improve its understanding and the reasoning behind some decisions.

## RESP (Redis framing protocol)
This is a client server application. Which means there is a need to transfer data over a network protocol. Data 
traversing networks need to be serialized and deserialized at reception. It is an agreement between the client and the 
server about the meanings of the data exchanged. We decided to implement the [RESP](https://redis.io/docs/reference/protocol-spec/) protocol which is a simple, yet
powerful serialization protocol. The second reason we chose this protocol is that we are building a kind of Redis
clone and wanted existing Redis clients to be compatible with our server.

## Code organization

### Frame folder
The [frame](frame) folder contains the code implementing the RESP protocol. Not everything is implemented for now.
Also, we rely on version 3 at this time. Here is how this section is organized:

- [frame.go](frame/frame.go) contains the high-level abstractions about frames and methods not tied to a specific frame
object.
- There is a file for each frame type (with its test file). For instance, RESP simple string is implemented in 
[string.go](frame/sstring.go) and [string_test.go](frame/sstring_test.go) files.

### Commands folder
This section is related to commands implementation. Each command is implemented in its own file along with an 
accompanying test file. We follow *Go command line pattern*. The files [command.go](command/command.go) provides
interface abstraction about commands. In fact, each command has to respect a structure which is enforced by this 
interface.
To add a new command, implement the Command interface and update the factory method.
Each new command should have its own file.

## The server folder
This is where we implement server logic: like spawning a new server, listening to connections, processing commands and 
responding to clients. We heavily rely on a Go concurrency model.