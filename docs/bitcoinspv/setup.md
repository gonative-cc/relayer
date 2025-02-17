# Bitcoin-SPV Setup

## Install Dependencies

- Docker
    1. Download Docker Desktop from [https://www.docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop)
    2. Install Docker Desktop
    3. Start Docker Desktop
    4. Verify installation by running: `docker --version`

- Golang
    1. Download Go from [https://golang.org/dl/](https://golang.org/dl/)
    2. Install Go by following instructions for your OS
    3. Add Go to your PATH
    4. Verify installation by running: `go version`
    5. Set GOPATH environment variable

## Run bitcoind node and bitcoin-lightclient

- Setup the docker containers for `bitcoind` node and `bitcoin-lightclient`, refer to [../../contrib/bitcoin-mock.md](../../contrib/bitcoin-mock.md) for instructions.
