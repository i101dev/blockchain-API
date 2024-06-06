package main

import (
	"fmt"

	"github.com/i101dev/blockchain-api/blockchain"
	"github.com/i101dev/blockchain-api/wallet"
)

// ---------------------------------------------------
func main() {
	// blockchain := blockchain.NewBlockchain("MINING_REWARD_ADDRESS")

	// blockchain.AddTransaction("A", "B", 1.0)
	// blockchain.AddTransaction("B", "C", 2.0)
	// blockchain.AddTransaction("C", "A", 3.0)
	// blockchain.Mining()

	// blockchain.AddTransaction("R", "G", 200.0)
	// blockchain.AddTransaction("H", "P", 15.0)
	// blockchain.Mining()

	// blockchain.Print()

	// fmt.Printf("C %.1f\n", blockchain.CalculateTotalAmount("C"))
	// fmt.Printf("A %.1f\n", blockchain.CalculateTotalAmount("A"))

	// fmt.Printf("MINING_REWARD_ADDRESS %.1f\n", blockchain.CalculateTotalAmount("MINING_REWARD_ADDRESS"))

	fmt.Println("\n***************")

	miner := wallet.NewWallet()
	sender := wallet.NewWallet()
	recipient := wallet.NewWallet()

	// fmt.Printf("Private key: %s\n", sender.PrivateKeyStr())
	// fmt.Printf("Public key: %s\n", sender.PublicKeyStr())
	// fmt.Printf("Address: %s\n", sender.BlockchainAddress())

	tx1_amount := 13

	// Transaction
	tx1 := wallet.NewWalletTransaction(sender.PrivateKey(), sender.PublicKey(), sender.BlockchainAddress(), recipient.BlockchainAddress(), float32(tx1_amount))

	// Blockchain
	blockchain := blockchain.NewBlockchain(miner.BlockchainAddress())

	isAdded := blockchain.AddTransaction(sender.BlockchainAddress(), recipient.BlockchainAddress(), float32(tx1_amount), sender.PublicKey(), tx1.GenerateSignature())

	fmt.Println("transaction added successfully: ", isAdded)
	// fmt.Printf("signature: %s\n", tx1.GenerateSignature())

	blockchain.Mining()
	blockchain.Print()
	//
	fmt.Printf("*** >>> miner: (%.1f)\n", blockchain.CalculateTotalAmount(miner.BlockchainAddress()))
	fmt.Printf("*** >>> sender: (%.1f)\n", blockchain.CalculateTotalAmount(sender.BlockchainAddress()))
	fmt.Printf("*** >>> recipient: (%.1f)\n", blockchain.CalculateTotalAmount(recipient.BlockchainAddress()))
}
