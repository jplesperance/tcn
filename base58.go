// Copyright 2018 Jesse P Lesperance. All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.
/*
	Transaction Cryptocurrency Network

	Package Main is the only package
	The goal is to maintain a minimal main file
 */
package main

import (
	"bytes"
	"math/big"
)

// The characters we will use for Base58
var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Base58Encode encodes a byte array to Base58
func Base58Encode(input []byte) []byte {
	var result []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}


	if input[0] == 0x00 {
		result = append(result, b58Alphabet[0])
	}

	ReverseBytes(result)

	return result
}

// Base58Decode decodes Base58-encoded data
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)

	for _, b := range input {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()

	if input[0] == b58Alphabet[0] {
		decoded = append([]byte{0x00}, decoded...)
	}

	return decoded
}

// Reverse pairs of bytes, used as part for base58 encoding
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i<j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}