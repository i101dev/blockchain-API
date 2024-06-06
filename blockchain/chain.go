package blockchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/i101dev/blockchain-api/utils"
)

// ------------------------------------------------------------------
const (
	MINING_DIFFICULTY = 3
	MINING_SENDER     = "THE BLOCKCHAIN"
	MINING_REWARD     = 100.0
)

// ------------------------------------------------------------------

type Blockchain struct {
	transactionPool   []*Transaction // equivalent to `memPool`
	chain             []*Block
	blockchainAddress string
}

func NewBlockchain(blockchainAddress string) *Blockchain {
	b := &Block{}
	bc := new(Blockchain)
	bc.blockchainAddress = blockchainAddress
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
	fmt.Printf("%s\n", strings.Repeat("*", 89))
}

func (bc *Blockchain) AddTransaction(sender string, recipient string, value float32, senderPublicKey *ecdsa.PublicKey, sig *utils.Signature) bool {

	txn := NewTransaction(sender, recipient, value)

	if sender == MINING_SENDER {
		bc.transactionPool = append(bc.transactionPool, txn)
		return true
	}

	if bc.VerifyTransactionSignature(senderPublicKey, sig, txn) {
		bc.transactionPool = append(bc.transactionPool, txn)
		return true
	}

	fmt.Println("failed to verify transaction signature")

	return false
}

func (bc *Blockchain) VerifyTransactionSignature(senderPublicKey *ecdsa.PublicKey, sig *utils.Signature, txn *Transaction) bool {
	m, _ := json.Marshal(txn)
	hash := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublicKey, hash[:], sig.R, sig.S)
}

func (bc *Blockchain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.transactionPool {
		newTx := NewTransaction(t.senderBlockchainAddress, t.recipientBlockchainAddress, t.value)
		transactions = append(transactions, newTx)
	}
	return transactions
}

func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, transactions []*Transaction, difficulty int) bool {

	zeros := strings.Repeat("0", difficulty)

	guessBlock := Block{
		nonce:        nonce,
		timestamp:    0,
		previousHash: previousHash,
		transactions: transactions,
	}
	guessHashStr := fmt.Sprintf("%x", guessBlock.Hash())

	return guessHashStr[:difficulty] == zeros
}

func (bc *Blockchain) ProofOfWork() int {

	transactions := bc.CopyTransactionPool()
	previousHash := bc.LastBlock().Hash()

	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULTY) {
		nonce += 1
	}

	return nonce
}

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*Transaction{}
	return b
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) Mining() bool {
	bc.AddTransaction(MINING_SENDER, bc.blockchainAddress, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, previousHash)
	// log.Println("action=mining, status=success")
	return true
}

func (bc *Blockchain) CalculateTotalAmount(blockchainAddress string) float32 {

	var totalAmount float32 = 0.0

	for _, b := range bc.chain {
		for _, t := range b.transactions {

			if t.recipientBlockchainAddress == blockchainAddress {
				totalAmount += t.value
			}

			if t.senderBlockchainAddress == blockchainAddress {
				totalAmount -= t.value
			}
		}
	}

	return totalAmount
}

// -------------------------------------------------------------------------

type Transaction struct {
	senderBlockchainAddress    string
	recipientBlockchainAddress string
	value                      float32
}

func NewTransaction(sender string, recipient string, value float32) *Transaction {
	return &Transaction{sender, recipient, value}
}

func (t *Transaction) Print() {
	fmt.Printf("\n	%s", strings.Repeat("-", 40))
	fmt.Printf("\n	> sender address: %s", t.senderBlockchainAddress)
	fmt.Printf("\n	> recipient address: %s", t.recipientBlockchainAddress)
	fmt.Printf("\n	> transaction value: %.1f", t.value)
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value     float32 `json:"value"`
	}{
		Sender:    t.senderBlockchainAddress,
		Recipient: t.recipientBlockchainAddress,
		Value:     t.value,
	})
}
