# Transactional Cryptocurrency Network

Transactional Cryptocurrency Network, TCN, is a simple blockchain and cryptocurrency implemented in Go.  TCN supports wallets, block mining, transactions and a distributed network.  

## Motivation

I decided to start this project as a way to gain a better understanding for blockchain technology and implementation.  TCN is loosely based on the BitCoin Specification.  

## Getting Started

These instructions will help you get TCN installed and gat the initial blockchain created and mine the Genesis block.

### Installing

There are 2 methods to install TCN:

Install with go get

```bash
go get -u github.com/jplesperance/tcn
```

Clone the reop and install

```bash
cd $GOPATH/src/github.com/jplesperance
git clone https://github.com/jplesperance/tcn
cd tcn
go install -v
```

## Usage

### Create the blockchain

When creating the blockchain, you must specify an address, that address will recieve credit(coins) for the mining of the Genesis block

```bash
tcn createblockchain -address <wallet-address>
```

### Getting the balance of a wallet

```bash
tcn getbalance -address <wallet-address>
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

Please read [CONTRIBUTING.md](https://github.com/jplesperance/tcn/src/master/CONTRIBUTING.md)

## Versioning

We use [SemVer](http://server.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/jplesperance/tcn/tags).

## Authors

* **Jesse P Lesperance** = *Initial Work* - [jplesperance](https://lesperance.io)

