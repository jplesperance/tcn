package main

import (
	"math/big"
	"fmt"
	"bytes"
	"encoding/binary"
	"log"
	"crypto/sha256"
)

// Proof of work data structure definition
//
// block: pointer to the current block being worked on
// target: the target value the generated hash will be compared to
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// Method to create a new Proof of Work
// Initialize a big int with a value of 1 and shift it left by 256 - targetBits.
// 256 is used as its the length of the SHA-256 hashing algorithm
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

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
	fmt.Printf("maxNonce: %d\n", maxNonce)

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x %d %d", hash, maxNonce-nonce, nonce)
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

// functionality to validate the output of ProofOfWork
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}