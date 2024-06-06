package blockchain

import (
	"fmt"
	"strings"
)

type Blockchain struct {
	transactionPool []string // equivalent to `memPool`
	chain           []*Block
}

func NewBlockchain() *Blockchain {
	b := &Block{}
	bc := new(Blockchain)
	bc.CreateBlock(0, b.Hash())
	return bc
}

func (bc *Blockchain) Print() {
	fmt.Println("--------------")
	fmt.Println("| Blockchain |")
	fmt.Println("--------------")
	for i, block := range bc.chain {
		fmt.Println()
		fmt.Printf("%s Block %d %s\n", strings.Repeat("=", 15), i, strings.Repeat("=", 65))
		block.Print()
	}
	fmt.Println()
	fmt.Printf("%s\n", strings.Repeat("*", 70))
}

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash)
	bc.chain = append(bc.chain, b)
	return b
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}
