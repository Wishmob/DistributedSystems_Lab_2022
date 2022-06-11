# Repo for system developed in _Distributed Systems_ lecture in SS22

by Alexander Breiter & Lukas Schandl


![System Architecture](/img/VSArchitecture.png)

## Building and Usage:
0. Install docker, see [docs](https://www.docker.com/get-started/)

1. start docker

2. `make start` on command line in root dir

3. be amazed

## Scaling

It is possible to start the system with multiple sensors at once.
Do `make start5` `make start25` `make start50` `make start100` for an equivalent amount of concurrent sensors
