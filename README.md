# Repo for system developed in _Distributed Systems_ lecture in SS22

by Alexander Breiter & Lukas Schandl

Building and Usage:
0. Install Go, see [docs](https://golang.org/)

1. `make build` on command line

2. `bin/server` to start server component with defaults (`bin/server -h` for help)

3. `bin/client` to start client component with defaults (`bin/client -h` for help)

4. The client will end itself. Kill the server using ctrl-c.

5. `make clean` will remove binaries to start from scratch.

