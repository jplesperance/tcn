package main

import (
	"math"
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
	"os"
	"fmt"
)

// define the maximum value of nonce
var maxNonce = math.MaxInt64

// define the difficulty of mining
// Picked an arbitrary number, the goal is to have a target less than 256 bits in memory

// Todo: Implement a difficulty adjusting algorithm
const targetBits = 20
const blocksBucket = "blocks"
const dbFile = "blockchain.db"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// Blockchaing implements interactions with a DB
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
	fmt.Println("Next()")
	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		fmt.Println("Get the hash of the encoded block")
		encodedBlock := b.Get(i.currentHash)
		fmt.Println("decod block")
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
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("1"))

		return nil
	})

	if err != nil {
		log.Panic("Unable to retieve blockchain from the database", err)
	}

	newBlock := NewBlock(transactions, lastHash)

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
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// Check if a blockchain exists, if not, generate a genesis block and create blockchain
func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		log.Fatal("No existing blockchain was found.")
		fmt.Println("No existing blockchain found.  Create on first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	fmt.Println("database opened")
	if err != nil {
		log.Fatal("Unable to open database", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		fmt.Println("updating database")
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("1"))
		return nil
	})
	fmt.Println("Not sure where from here")
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	return &bc

}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	fmt.Println("Initiate Iterator")
	bci := bc.Iterator()

	for {
		fmt.Println("Get next Block")
		block := bci.Next()

		for _, tx := range block.Transactions {
			fmt.Println("Encode transaction id to string")
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				fmt.Println("check if can be unlocked")
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	fmt.Println("Find unspent transactions")
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

	Work:
		for _, tx := range unspentTXs {
			fmt.Println("Encode transaction IDs to a string")
			txID := hex.EncodeToString(tx.ID)

			for outIdx, out := range tx.Vout {
				if out.CanBeUnlockedWith(address) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
					if accumulated >= amount {
						break Work

					}
				}
			}
		}


		return accumulated, unspentOutputs
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		fmt.Println("Creating Genesis block")
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}
		fmt.Println("Storing genesis block in database")
		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		tip = genesis.Hash

		return nil
	})
	fmt.Println("Interesting")
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}