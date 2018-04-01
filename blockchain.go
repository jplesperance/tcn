package main

import (
	"math"
	"bytes"
	"crypto/ecdsa"
	"errors"
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
const targetBits = 21
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

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
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
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte

	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("Error: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

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
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Println("Unable to add new block to blockchain", err)
			return err
		}
		bc.tip = newBlock.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return newBlock
}



// Check if a blockchain exists, if not, generate a genesis block and create blockchain
func NewBlockchain() *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found.  Create on first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)

	if err != nil {
		log.Panic("Unable to open database", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	return &bc

}

func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)

	bci := bc.Iterator()

	for {

		block := bci.Next()

		for _, tx := range block.Transactions {

			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}



/*func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)

	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

	Work:
		for _, tx := range unspentTXs {

			txID := hex.EncodeToString(tx.ID)

			for outIdx, out := range tx.Vout {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
					if accumulated >= amount {
						break Work

					}
				}
			}
		}


		return accumulated, unspentOutputs
}*/

func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction is not found")
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

	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)


	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}

		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}