package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/i101dev/blockchain-api/blockchain"
	"github.com/i101dev/blockchain-api/utils"
	"github.com/i101dev/blockchain-api/wallet"
)

var cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

type BlockchainServer struct {
	port uint16
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{port}
}

func (bcs *BlockchainServer) Port() uint16 {
	return bcs.port
}

func (bcs *BlockchainServer) GetBlockchain() *blockchain.Blockchain {

	ID := "blockchain"

	bc, ok := cache[ID]

	if !ok {
		minerWallet := wallet.NewWallet()
		bc = blockchain.NewBlockchain(minerWallet.BlockchainAddress(), bcs.Port())
		cache[ID] = bc
		log.Printf("\nAddress: %s", minerWallet.BlockchainAddress())
	}

	return bc
}

func (bcs *BlockchainServer) GetChainData(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockchain()
		m, _ := bc.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		log.Printf("ERROR: Invalid HTTP Method")
	}
}

func (bcs *BlockchainServer) Transactions(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")

		bc := bcs.GetBlockchain()
		transactions := bc.TransactionPool()

		m, _ := json.Marshal(struct {
			Transactions []*blockchain.Transaction `json:"transactions"`
			Length       int                       `json:"length"`
		}{
			Transactions: transactions,
			Length:       len(transactions),
		})

		io.WriteString(w, string(m[:]))

	case http.MethodPost:

		var txn blockchain.TransactionRequest

		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&txn)

		if err != nil {
			log.Printf("ERROR: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// fmt.Printf("%+v", txn)

		if !txn.Validate() {
			log.Println("ERROR: Missing field(s)")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		publicKey := utils.PublicKeyFromString(*txn.SenderPublicKey)
		signature := utils.SignatureFromString(*txn.Signature)
		bc := bcs.GetBlockchain()
		// _ = publicKey
		// _ = signature
		// _ = bc

		isCreated := bc.CreateTransaction(*txn.SenderBlockchainAddress, *txn.RecipientBlockchainAddress, *txn.Value, publicKey, signature)

		w.Header().Add("Content-Type", "application/json")

		if !isCreated {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusCreated)
		}

	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Mine(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		isMined := bc.Mining()

		if !isMined {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		w.Header().Add("Content-Type", "application/json")

	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) StartMine(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		bc.StartMining()
		w.Header().Add("Content-Type", "application/json")

	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Amount(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodGet:

		blockchainAddress := req.URL.Query().Get("blockchain_address")
		amount := bcs.GetBlockchain().CalculateTotalAmount(blockchainAddress)

		ar := &blockchain.AmountResponse{Amount: amount}
		m, _ := ar.MarshalJSON()

		w.Header().Add("Content-Type", "app")
		io.WriteString(w, string(m[:]))

	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Run() {

	http.HandleFunc("/", bcs.GetChainData)
	http.HandleFunc("/transactions", bcs.Transactions)
	http.HandleFunc("/mine", bcs.Mine)
	http.HandleFunc("/mine/start", bcs.StartMine)
	http.HandleFunc("/amount", bcs.Amount)

	hostURL := "0.0.0.0:" + strconv.Itoa(int(bcs.Port()))

	fmt.Println("Blockchain Server is live @:", hostURL)
	log.Fatal(http.ListenAndServe(hostURL, nil))
}
