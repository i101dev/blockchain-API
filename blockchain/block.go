package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type Block struct {
	nonce        int
	previousHash [32]byte
	timestamp    int64
	transactions []string
}

func NewBlock(nonce int, previousHash [32]byte) *Block {
	b := new(Block)
	b.nonce = nonce
	b.timestamp = time.Now().UnixNano()
	b.previousHash = previousHash
	return b
}

func (b *Block) Print() {
	// fmt.Println("--------------------------------------------")
	fmt.Printf("> timestamp		%d\n", b.timestamp)
	fmt.Printf("> nonce			%d\n", b.nonce)
	fmt.Printf("> previousHash		%x\n", b.previousHash)
	fmt.Printf("> transactions		%s\n", b.transactions)
	// fmt.Println("--------------------------------------------")
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp    int64    `json:"timestamp"`
		Nonce        int      `json:"nonce"`
		PreviousHash [32]byte `json:"previous_hash"`
		Transactions []string `json:"transactions"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		PreviousHash: b.previousHash,
		Transactions: b.transactions,
	})
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	fmt.Println(string(m))
	return sha256.Sum256(m)
}
