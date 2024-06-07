package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
	"text/template"

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
		decoder := json.NewDecoder(req.Body)
		var txn wallet.WalletTXNRequest
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
		value := float32(*txn.Value)

		w.Header().Add("Content-Type", "application/json")

		// fmt.Println(publicKey)
		// fmt.Println(privateKey)
		// fmt.Println("From: ", *txn.SenderBlockchainAddress)
		// fmt.Println("To: ", *txn.RecipientBlockchainAddress)
		// fmt.Printf("Value: %.1f\n", value)

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
