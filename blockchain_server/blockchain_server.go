package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/i101dev/blockchain-api/blockchain"
	"github.com/i101dev/blockchain-api/wallet"
)

var cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

type BlockchainServer struct {
	port uint16
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{port}
}

func (bcs *BlockchainServer) GetChain(w http.ResponseWriter, req *http.Request) {
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

func (bcs *BlockchainServer) Run() {

	http.HandleFunc("/", bcs.GetChain)

	hostURL := "0.0.0.0:" + strconv.Itoa(int(bcs.Port()))

	fmt.Println("Server is live @:", hostURL)
	log.Fatal(http.ListenAndServe(hostURL, nil))
}
