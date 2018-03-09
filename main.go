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
	"strconv"
	"encoding/gob"

	"github.com/boltdb/bolt"
)

// define the maximum value of nonce
var maxNonce = math.MaxInt64

// define the difficulty of mining
// Picked an arbitrary number, the goal is to have a target less than 256 bits in memory

// Todo: Implement a difficulty adjusting algorithm
const targetBits = 24
const blocksBucket = "blocks"
const dbFile = "blockchain.db"

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
	tip []byte
	db  *bolt.DB
}

// Proof of work data structure definition
//
// block: pointer to the current block being worked on
// target: the target value the generated hash will be compared to
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

type BlockchainIterator struct {
	currentHash []byte
	db *bolt.DB
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})
	if err != nil {
		log.Print("Unbale to retrieve the blockchain from the database")

	}

	i.currentHash = block.PrevBlockHash

	return block
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
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("1"))

		return nil
	})

	if err != nil {
		log.Panic("Unable to retieve blockchain from the database", err)
	}

	newBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Println("Error updating block", err)
			return err
		}
		err = b.Put([]byte("1"), newBlock.Hash)
		if err != nil {
			log.Println("Unable to add new block to blockchain", err)
			return err
		}
		bc.tip = newBlock.Hash

		return nil
	})
}

// A function for generating a Genesis block, needed as the first block in a
// blockchain
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// Check if a blockchain exists, if not, generate a genesis block and create blockchain
func NewBlockchain() *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal("Unable to open database", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Fatal("Unable to create blocksBucket", err)
				return err
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Fatal("Unable to add Genesis block to database", err)
				return err
			}
			err = b.Put([]byte("1"), genesis.Hash)
			if err != nil {
				log.Println("Unable to update the genesis hash to the database", err)
				return err
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("1"))
		}
		return nil
	})

	bc := Blockchain{tip, db}
	return &bc

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

// Meat of the app
func main() {
	bc := NewBlockchain()

	bc.AddBlock("Send 1 btc to Ivan")
	bc.AddBlock("Send 2 btc to Igor")

	for _, block := range bc.tip {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("POW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
