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

### Frame
The [frame](frame) folder contains the code implementing the RESP protocol. Not everything is implemented for now.
Also, we rely on version 3 at this time. Here is how this section is organized:

- [frame.go](frame/frame.go) contains the high-level abstractions about frames and methods not tied to a specific frame
object.
- There is a file for each frame type (with its test file). For instance, RESP simple string is implemented in 
[string.go](frame/sstring.go) and [string_test.go](frame/sstring_test.go) files.

### Command
This section is related to commands implementation. Each command is implemented in its own file along with an 
accompanying test file. We follow *Go command line pattern*. The files [command.go](command/command.go) provides
interface abstraction about commands. In fact, each command has to respect a structure which is enforced by this 
interface.
To add a new command, implement the Command interface and update the factory method.
Each new command should have its own file.

### Database
The database is the storage backend for the server.
It is split into two parts: the storage layer and the eviction policy layer.
We use Go Map as storage.
We support multiple evictions policies which share a common interface.
To achieve this modularity, we separated the storage and eviction structures.
They are totally independent (each has its own state).
The logic for GET, SET,
DEL commands are already defined in [cache.go](db/cache.go)
so adding a new policy is as simple as implementing the Eviction interface.

### Server
This is where we implement server logic: like spawning a new server, listening to connections, processing commands and 
responding to clients. We heavily rely on a Go concurrency model.

### Concurrency
Each connection is handled by a goroutine which shares the database with others.
So we need to add some synchronization to avoid race condition.
Here is how our implementation performs against a real redis server in a mackbook air M2.

<u>Benchmarks</u>

Real Redis server
````commandline
➜  gcache git:(dev/shared-map) ✗ redis-benchmark -t set,get -n 10000000 -r 1000 -c 101 -q
SET: 157114.12 requests per second, p50=0.335 msec                    
GET: 153857.98 requests per second, p50=0.335 msec                    
````

Gcache server
````commandline
➜  gcache git:(dev/shared-map) ✗ redis-benchmark -t set,get -n 10000000 -r 1000 -c 101 -q
WARNING: Could not fetch server CONFIG
SET: 155513.73 requests per second, p50=0.327 msec                    
GET: 156047.62 requests per second, p50=0.327 msec  
````