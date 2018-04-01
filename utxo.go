package main

import (
	"encoding/hex"
	"log"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

type UTXOSet struct {
	Blockchain *Blockchain
}

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	upspentOutputs := make(map[string][]int)
}
