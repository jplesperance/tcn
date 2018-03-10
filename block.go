package main

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
)

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

// Create a new block, populate the fields and return it to the calling method
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// use encoding/gob to encode the block and return as a byte array
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Fatal("Block encoding failed: ", err)
	}

	return result.Bytes()
}

// Take the byte array, decode the block and return the struct
func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))

	err := decoder.Decode(&block)
	if err != nil {
		log.Fatal("Failed to decode block: ", err)
	}

	return &block
}

