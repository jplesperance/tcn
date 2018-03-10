package main

import (
	"math"
	"github.com/boltdb/bolt"
	"log"
)

// define the maximum value of nonce
var maxNonce = math.MaxInt64

// define the difficulty of mining
// Picked an arbitrary number, the goal is to have a target less than 256 bits in memory

// Todo: Implement a difficulty adjusting algorithm
const targetBits = 20
const blocksBucket = "blocks"
const dbFile = "blockchain.db"

// Blockchaing struct only needs an array to hold ordered hashes
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
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
		b := tx.Bucket([]byte(blocksBucket))
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
