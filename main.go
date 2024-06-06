package main

import (
	"fmt"

	"github.com/i101dev/blockchain-api/blockchain"
)

// ---------------------------------------------------
func main() {
	blockchain := blockchain.NewBlockchain("MINING_REWARD_ADDRESS")

	blockchain.AddTransaction("A", "B", 1.0)
	blockchain.AddTransaction("B", "C", 2.0)
	blockchain.AddTransaction("C", "A", 3.0)
	blockchain.Mining()

	blockchain.AddTransaction("R", "G", 200.0)
	blockchain.AddTransaction("H", "P", 15.0)
	blockchain.Mining()

	blockchain.Print()

	fmt.Printf("C %.1f\n", blockchain.CalculateTotalAmount("C"))
	fmt.Printf("A %.1f\n", blockchain.CalculateTotalAmount("A"))
	fmt.Printf("MINING_REWARD_ADDRESS %.1f\n", blockchain.CalculateTotalAmount("MINING_REWARD_ADDRESS"))
}
