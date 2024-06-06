package main

import "github.com/i101dev/blockchain-api/blockchain"

// ---------------------------------------------------
func main() {
	blockchain := blockchain.NewBlockchain()

	previousHash := blockchain.LastBlock().Hash()
	blockchain.CreateBlock(2, previousHash)

	previousHash = blockchain.LastBlock().Hash()
	blockchain.CreateBlock(4, previousHash)

	previousHash = blockchain.LastBlock().Hash()
	blockchain.CreateBlock(6, previousHash)

	blockchain.Print()
}
