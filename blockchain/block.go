package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Block struct {
	nonce        int
	previousHash [32]byte
	timestamp    int64
	transactions []*Transaction
}

func NewBlock(nonce int, previousHash [32]byte, transactions []*Transaction) *Block {
	b := new(Block)
	b.nonce = nonce
	b.timestamp = time.Now().UnixNano()
	b.previousHash = previousHash
	b.transactions = transactions
	return b
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp int64 `json:"timestamp"`
		Nonce     int   `json:"nonce"`
		// ThisHash     string         `json:"this_hash"`
		PreviousHash string         `json:"previous_hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp: b.timestamp,
		Nonce:     b.nonce,
		// ThisHash:     fmt.Sprintf("%x", b.Hash()),
		PreviousHash: fmt.Sprintf("%x", b.previousHash),
		Transactions: b.transactions,
	})
}

func (b *Block) UnmarshalJSON(data []byte) error {

	var previousHash string

	v := &struct {
		Timestamp    *int64          `json:"timestamp"`
		Nonce        *int            `json:"nonce"`
		PreviousHash *string         `json:"previous_hash"`
		Transactions *[]*Transaction `json:"transactions"`
	}{
		Timestamp:    &b.timestamp,
		Nonce:        &b.nonce,
		PreviousHash: &previousHash,
		Transactions: &b.transactions,
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	ph, _ := hex.DecodeString(*v.PreviousHash)
	copy(b.previousHash[:], ph[:32])

	return nil
}

func (b *Block) Print() {
	fmt.Printf("> timestamp		%d\n", b.timestamp)
	fmt.Printf("> nonce			%d\n", b.nonce)
	fmt.Printf("> previousHash		%x\n", b.previousHash)
	fmt.Println("\n### Transactions:")
	for _, t := range b.transactions {
		t.Print()
	}
	fmt.Println()
	fmt.Println()
	fmt.Printf("\n%s", strings.Repeat("=", 89))
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256(m)
}

func (b *Block) Transactions() []*Transaction {
	return b.transactions
}

func (b *Block) PreviousHash() [32]byte {
	return b.previousHash
}

func (b *Block) Timestamp() int64 {
	return b.timestamp
}

func (b *Block) Nonce() int {
	return b.nonce
}
