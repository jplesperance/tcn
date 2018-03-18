# BittyCoin
BittyCoin is a simple blockchain and cryptocurrency implemented in Go.  BittyCoin supports wallets, block mining, transactions and a distributed network.  

## Motivation
I decided to start this project as a way to gain a better understanding for blockchain technology and implementation.  BittyCoin is loosely based on the BitCoin Specification.  

## Getting Started
These instructions will help you get BittyCoin installed and gat the initial blockchain created and mine the Genesis block.

### Installing

There are 2 methods to install BittyCoin:

Install with go get
```bash
go get -u bitbucket.org/lesperanceio/bittycoin
```

Clone the reop and install
```bash
cd $GOPATH/src/bitbucket.org/lesperanceio
git clone https://bitbucket.org/lesperanceio/bittycoin.git 
cd bittycoin
go install -v
```

## Usage

### Create the blockchain
When creating the blockchain, you must specify an address, that address will recieve credit(coins) for the mining of the Genesis block
```bash
bittycoin createblockchain -address <wallet-address>
```

### Getting the balance of a wallet
```bash
bittycoin getbalance -address <wallet-address>
```

### Send coins to another wallet
```bash
bittycoin send -to <destination-wallet> -from <source-wallet> -amount <amount of coins to send>
```

### Print all the blocks of the blockchain
```bash
bittycoin printchain
```

## Built With
* [BoltDB](https://github.com/boltdb/bolt) - BoltDB: embedded k/v database for Go

## Contributing
Please read [CONTRIBUTING.md](https://bitbucket.org/lesperanceio/bittycoin/CONTRIBUTING.md)

## Versioning
We use [SemVer]() for versioning. For the versions available, see the [tags on this repository]().