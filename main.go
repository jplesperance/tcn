package main

import (
	"bytes"
	"crypto/sha256"
	"time"
	"fmt"
	"math/big"
	"encoding/binary"
	"log"
	"math"
)


// define the maximum value of nonce
var maxNonce = math.MaxInt64

// define the difficulty of mining
// Picked an arbitrary number, the goal is to have a target less than 256 bits in memory

// Todo: Implement a difficulty adjusting algorithm
const targetBits = 24

// Struct for out blocks
//
// Timestamp: the timestamp for when the block was created
// Data: actual information contained in the block
// PrevBlockHash: the hash of the previous block
// Hash: the hash of the current block
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// Blockchaing struct only needs an array to hold ordered hashes
type Blockchain struct {
	blocks []*Block
}

// Proof of work data structure definition
//
// block: pointer to the current block being worked on
// target: the target value the generated hash will be compared to
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// Create a new block, populate the fields and return it to the calling method
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// Create and add a new block to the blockchain
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

// A function for generating a Genesis block, needed as the first block in a
// blockchain
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// A function to generate a Genesis block and create a new blockchain
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}

// Method to create a new Proof of Work
// Initialize a big int with a value of 1 and shift it left by 256 - targetBits.
// 256 is used as its the length of the SHA-256 hashing algorithm
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(156-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

// Prepare the data for hashing
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		}, []byte{},
	)
	return data
}

// Implement the core of the ProofOfWork functionality
// Initialized the data, hashes the data, validates the hash
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}

	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Convert integer into hexidecimal value
func IntToHex(n int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, n)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// Meat of the app
func main() {
	bc := NewBlockchain()

	bc.AddBlock("Send 1 btc to Ivan")
	bc.AddBlock("Send 2 btc to Igor")

	for _, block := range bc.blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}
