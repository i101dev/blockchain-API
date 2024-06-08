package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/i101dev/blockchain-api/utils"
)

// ------------------------------------------------------------------
const (
	MINING_DIFFICULTY = 3
	MINING_REWARD     = 100.0
	MINING_SENDER     = "THE BLOCKCHAIN"
	MINING_TIMER      = 20

	BLOCKCHAIN_PORT_RANGE_START      = 5000
	BLOCKCHAIN_PORT_RANGE_END        = 5003
	NEIGHBOR_IP_RANGE_START          = 0
	NEIGHBOR_IP_RANGE_END            = 1
	BLOCKCHIN_NEIGHBOR_SYNC_TIME_SEC = 20
)

// ------------------------------------------------------------------

type Blockchain struct {
	mux               sync.Mutex
	transactionPool   []*Transaction // equivalent to `memPool`
	chain             []*Block
	blockchainAddress string
	port              uint16

	muxNeighbors sync.Mutex
	peers        []string
}

func NewBlockchain(blockchainAddress string, port uint16) *Blockchain {

	b := &Block{}
	bc := new(Blockchain)

	bc.port = port
	bc.blockchainAddress = blockchainAddress

	bc.CreateBlock(0, b.Hash())

	return bc
}

func (bc *Blockchain) Run() {
	// bc.StartMining()
	bc.StartSyncPeers()
	bc.ResolveConflicts()
}

func (bc *Blockchain) SetNeighbors() {
	bc.peers = utils.FindNeighbors(
		utils.GetHost(), bc.port,
		NEIGHBOR_IP_RANGE_START, NEIGHBOR_IP_RANGE_END,
		BLOCKCHAIN_PORT_RANGE_START, BLOCKCHAIN_PORT_RANGE_END)
	log.Printf("%v", bc.peers)
}

func (bc *Blockchain) SyncNeighbors() {
	bc.muxNeighbors.Lock()
	defer bc.muxNeighbors.Unlock()
	bc.SetNeighbors()
}

func (bc *Blockchain) StartSyncPeers() {
	bc.SyncNeighbors()
	_ = time.AfterFunc(time.Second*BLOCKCHIN_NEIGHBOR_SYNC_TIME_SEC, bc.StartSyncPeers)
}

func (bc *Blockchain) Chain() []*Block {
	return bc.chain
}

func (bc *Blockchain) TransactionPool() []*Transaction {
	return bc.transactionPool
}

func (bc *Blockchain) ClearTransactionPool() {
	bc.transactionPool = bc.transactionPool[:0]
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*Block `json:"blocks"`
	}{
		Blocks: bc.chain,
	})
}

func (bc *Blockchain) UnmarshalJSON(data []byte) error {

	v := &struct {
		Blocks []*Block `json:"chain"`
	}{
		Blocks: bc.chain,
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	return nil
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

func (bc *Blockchain) CreateTransaction(sender string, recipient string, value float32, senderPublicKey *ecdsa.PublicKey, sig *utils.Signature) bool {

	isTransacted := bc.AddTransaction(sender, recipient, value, senderPublicKey, sig)

	if isTransacted {
		for _, n := range bc.peers {

			pubKeyStr := fmt.Sprintf("%064x%064x", senderPublicKey.X.Bytes(), senderPublicKey.Y.Bytes())
			sigStr := sig.String()

			bt := &TransactionRequest{&sender, &recipient, &pubKeyStr, &sigStr, &value}

			m, _ := json.Marshal(bt)
			buf := bytes.NewBuffer(m)

			endpoint := fmt.Sprintf("http://%s/transactions", n)
			client := &http.Client{}

			req, _ := http.NewRequest("PUT", endpoint, buf)
			resp, _ := client.Do(req)

			log.Printf("%+v", resp)
		}
	}

	return isTransacted
}

func (bc *Blockchain) AddTransaction(sender string, recipient string, value float32, senderPublicKey *ecdsa.PublicKey, sig *utils.Signature) bool {

	txn := NewTransaction(sender, recipient, value)

	if sender == MINING_SENDER {
		bc.transactionPool = append(bc.transactionPool, txn)
		return true
	}

	if bc.VerifyTransactionSignature(senderPublicKey, sig, txn) {

		// if bc.CalculateTotalAmount(sender) < value {
		// 	log.Println("insufficient funds")
		// 	return false
		// }

		bc.transactionPool = append(bc.transactionPool, txn)
		return true
	}

	// fmt.Println("failed to verify transaction signature")

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

	for _, p := range bc.peers {
		endpoint := fmt.Sprintf("http://%s/transactions", p)
		client := &http.Client{}

		req, _ := http.NewRequest("DELETE", endpoint, nil)
		resp, _ := client.Do(req)

		log.Printf("%+v", resp)
	}

	return b
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) Mining() bool {

	bc.mux.Lock()
	defer bc.mux.Unlock()

	if len(bc.transactionPool) == 0 {
		fmt.Println("\nNo transactions - Mining skipped")
		return false
	}

	fmt.Println("\nMining NOW!")

	bc.AddTransaction(MINING_SENDER, bc.blockchainAddress, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, previousHash)
	log.Println("action=mining, status=success")

	for _, p := range bc.peers {
		endpoint := fmt.Sprintf("http://%s/consensus", p)
		client := &http.Client{}
		req, _ := http.NewRequest("PUT", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%+v", resp)
	}

	return true
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	_ = time.AfterFunc(time.Second*MINING_TIMER, bc.StartMining)
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

func (bc *Blockchain) ValidChain(chain []*Block) bool {

	preBlock := chain[0]
	currentIndex := 1

	for currentIndex < len(chain) {

		b := chain[currentIndex]

		if b.previousHash != preBlock.Hash() {
			return false
		}

		if !bc.ValidProof(b.Nonce(), b.PreviousHash(), b.Transactions(), MINING_DIFFICULTY) {
			return false
		}

		preBlock = b
		currentIndex += 1
	}

	return true
}

func (bc *Blockchain) ResolveConflicts() bool {

	var longestChain []*Block = nil
	maxLength := len(bc.chain)

	fmt.Println("\nResolving conflicts...")

	for _, p := range bc.peers {

		endpoint := fmt.Sprintf("http://%s/chain", p)
		resp, _ := http.Get(endpoint)

		if resp.StatusCode == 200 {
			var bcResp Blockchain
			decoder := json.NewDecoder(resp.Body)
			_ = decoder.Decode(&bcResp)

			chain := bcResp.Chain()

			if len(chain) > maxLength && bc.ValidChain(chain) {
				maxLength = len(chain)
				longestChain = chain
			}
		}
	}

	if longestChain != nil {
		bc.chain = longestChain
		log.Printf("Resovle confilicts replaced")
		return true
	}

	log.Printf("Resovle conflicts not replaced")
	return false
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
	fmt.Printf("\n	%s", strings.Repeat("-", 55))
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

func (t *Transaction) UnmarshalJSON(data []byte) error {

	v := &struct {
		Sender    *string  `json:"sender_blockchain_address"`
		Recipient *string  `json:"recipient_blockchain_address"`
		Value     *float32 `json:"value"`
	}{
		Sender:    &t.senderBlockchainAddress,
		Recipient: &t.recipientBlockchainAddress,
		Value:     &t.value,
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	return nil
}

// -------------------------------------------------------------------------

type TransactionRequest struct {
	SenderBlockchainAddress    *string  `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string  `json:"recipient_blockchain_address"`
	SenderPublicKey            *string  `json:"sender_public_key"`
	Signature                  *string  `json:"signature"`
	Value                      *float32 `json:"value"`
}

// func (t *TransactionRequest) Print() {
// 	fmt.Printf("\n	%s", strings.Repeat("-", 55))
// 	fmt.Printf("\n	> SenderBlockchainAddress: %s", *t.SenderBlockchainAddress)
// 	fmt.Printf("\n	> RecipientBlockchainAddress: %s", *t.RecipientBlockchainAddress)
// 	fmt.Printf("\n	> SenderPublicKey: %s", *t.SenderPublicKey)
// 	fmt.Printf("\n	> Signature: %s", *t.Signature)
// 	// fmt.Printf("\n	> value: %.1f", t.Value)
// }

func (tr *TransactionRequest) Validate() bool {

	if tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Signature == nil ||
		tr.Value == nil {
		return false
	}

	// if tr.Value == nil {
	// 	log.Println("Value MISSING")
	// 	return false
	// }
	// if tr.SenderBlockchainAddress == nil {
	// 	log.Println("SenderBlockchainAddress MISSING")
	// 	return false
	// }
	// if tr.RecipientBlockchainAddress == nil {
	// 	log.Println("RecipientBlockchainAddress MISSING")
	// 	return false
	// }
	// if tr.SenderPublicKey == nil {
	// 	log.Println("SenderPublicKey MISSING")
	// 	return false
	// }
	// if tr.Signature == nil {
	// 	log.Println("Signature MISSING")
	// 	return false
	// }

	return true
}

// -------------------------------------------------------------------------

type AmountResponse struct {
	Amount float32 `json:"amount"`
}

func (ar *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount float32 `json:"amount"`
	}{
		Amount: ar.Amount,
	})
}
