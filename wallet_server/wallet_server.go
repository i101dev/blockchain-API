package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"text/template"
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

func (ws *WalletServer) Run() {

	http.HandleFunc("/", ws.Index)

	hostURL := "0.0.0.0:" + strconv.Itoa(int(ws.Port()))

	fmt.Println("Wallet Server is live @:", hostURL)
	log.Fatal(http.ListenAndServe(hostURL, nil))
}
