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

func (ws *WalletServer) WalletAmount(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodGet:

		blockchainAddress := req.URL.Query().Get("blockchain_address")
		endpoint := fmt.Sprintf("%s/amount", ws.Gateway())

		// _ = blockchainAddress
		// _ = endpoint

		client := &http.Client{}
		bcsReq, _ := http.NewRequest("GET", endpoint, nil)

		// _ = client
		// _ = bcsReq

		q := bcsReq.URL.Query()
		q.Add("blockchain_address", blockchainAddress)
		bcsReq.URL.RawQuery = q.Encode()

		bcsResp, err := client.Do(bcsReq)
		if err != nil {
			log.Printf("%+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")

		if bcsResp.StatusCode == 200 {

			var amt blockchain.AmountResponse
			decoder := json.NewDecoder(bcsResp.Body)
			err := decoder.Decode(&amt)

			if err != nil {
				log.Printf("%+v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			m, _ := json.Marshal(struct {
				Message string  `json:"message"`
				Amount  float32 `json:"amount"`
			}{
				Message: "success",
				Amount:  amt.Amount,
			})

			io.WriteString(w, string(m[:]))

		} else {

			m, _ := json.Marshal(struct {
				Message string `json:"message"`
			}{
				Message: "QUERY FAILED",
			})

			io.WriteString(w, string(m[:]))
		}

	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (ws *WalletServer) Run() {

	http.HandleFunc("/", ws.Index)
	http.HandleFunc("/wallet", ws.Wallet)
	http.HandleFunc("/wallet/amount", ws.WalletAmount)
	http.HandleFunc("/transaction", ws.CreateTransaction)

	hostURL := "0.0.0.0:" + strconv.Itoa(int(ws.Port()))

	fmt.Println("Wallet Server is live @:", hostURL)
	log.Fatal(http.ListenAndServe(hostURL, nil))
}
