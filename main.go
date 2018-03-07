package main

import (
	"bytes"
	"strconv"
	"crypto/sha256"
	"time"
	"fmt"
)


// Struct for out blocks
//
// Timestamp: the timestamp for when the block was created
// Data: actual information contained in the block
// PrevBlockHash: the hash of the previous block
// Hash: the hash of the current block
type Block struct {
	Timestamp int64
	Data []byte
	PrevBlockHash []byte
	Hash []byte
}

// Blockchaing struct only needs an array to hold ordered hashes
type Blockchain struct {
	blocks []*Block
}

// This method concatenates the block fields and calculates a SHA-256 hash
// on the concatenated combination
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

// Create a new block, populate the fields and return it to the calling method
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}}
	block.SetHash()
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