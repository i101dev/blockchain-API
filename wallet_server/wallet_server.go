package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
	"text/template"

	"github.com/i101dev/blockchain-api/blockchain"
	"github.com/i101dev/blockchain-api/utils"
	"github.com/i101dev/blockchain-api/wallet"
)

const tempDir = "templates"

type WalletServer struct {
	port    uint16
	gateway string
}

func NewWalletServer(port uint16, gateway string) *WalletServer {
	return &WalletServer{port, gateway}
}

func (ws *WalletServer) Port() uint16 {
	return ws.port
}

func (ws *WalletServer) Gateway() string {
	return ws.gateway
}

func (ws *WalletServer) Index(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		t, err := template.ParseFiles(path.Join(tempDir, "index.html"))
		if err != nil {
			http.Error(w, "Unable to load template", http.StatusInternalServerError)
			log.Printf("ERROR: Unable to load template: %v", err)
			return
		}
		err = t.Execute(w, "")
		if err != nil {
			http.Error(w, "Unable to execute template", http.StatusInternalServerError)
			log.Printf("ERROR: Unable to execute template: %v", err)
		}
	default:
		log.Printf("ERROR: Invalid HTTP Method")
	}
}

func (ws *WalletServer) Wallet(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		w.Header().Add("Content-Type", "application/json")
		myWallet := wallet.NewWallet()
		m, _ := myWallet.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Panicln("ERROR: Invalid HTTP Method")
	}
}

func (ws *WalletServer) CreateTransaction(w http.ResponseWriter, req *http.Request) {

	switch req.Method {

	case http.MethodPost:

		var txn wallet.WalletTXNRequest

		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&txn)

		if err != nil {
			log.Printf("ERROR decoding wallet transaction: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !txn.Validate() {
			log.Println("ERROR: Missing field(s)")
			w.WriteHeader(http.StatusInternalServerError)
		}

		publicKey := utils.PublicKeyFromString(*txn.SenderPublicKey)
		privateKey := utils.PrivateKeyFromString(*txn.SenderPrivateKey, publicKey)
		value32 := float32(*txn.Value)

		w.Header().Add("Content-Type", "application/json")

		transaction := wallet.NewWalletTransaction(privateKey, publicKey, *txn.SenderBlockchainAddress, *txn.RecipientBlockchainAddress, value32)
		signature := transaction.GenerateSignature()
		signatureStr := signature.String()

		bt := &blockchain.TransactionRequest{
			RecipientBlockchainAddress: txn.RecipientBlockchainAddress,
			SenderBlockchainAddress:    txn.SenderBlockchainAddress,
			SenderPublicKey:            txn.SenderPublicKey,
			Signature:                  &signatureStr,
			Value:                      &value32,
		}

		m, _ := json.Marshal(bt)
		buf := bytes.NewBuffer(m)

		resp, _ := http.Post(ws.Gateway()+"/transactions", "application/json", buf)

		if resp.StatusCode == 201 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Transaction processed successfully"))
			// fmt.Println("\n*** >>> TRANSACTION SUCCESS! <<< ***")
			return
		}

		io.WriteString(w, "Transaction FAILED!")
		fmt.Println("\n*** >>> TRANSACTION FAILED! <<< ***")

	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: Invalid HTTP Method")
	}
}

func (ws *WalletServer) Run() {

	http.HandleFunc("/", ws.Index)
	http.HandleFunc("/wallet", ws.Wallet)
	http.HandleFunc("/transaction", ws.CreateTransaction)

	hostURL := "0.0.0.0:" + strconv.Itoa(int(ws.Port()))

	fmt.Println("Wallet Server is live @:", hostURL)
	log.Fatal(http.ListenAndServe(hostURL, nil))
}
