package main

import (
	"encoding/hex"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const rawHash string = "00010966776006953D5567439E5E39F86A0D273BEED61967F6"
const encodedHash string = "16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM"

func TestBase58Encode(t *testing.T) {

	hash, err := hex.DecodeString(rawHash)
	if err != nil {
		log.Fatal(err)
	}

	encoded := Base58Encode(hash)
	assert.Equal(t, encodedHash, string(encoded))

}

func TestBase58Decode(t *testing.T) {
	decoded := Base58Decode([]byte(encodedHash))
	assert.Equal(t, strings.ToLower(rawHash), hex.EncodeToString(decoded))
}