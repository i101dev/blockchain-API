package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/i101dev/blockchain-api/utils"
)

type Wallet struct {
	privateKey        *ecdsa.PrivateKey
	publicKey         *ecdsa.PublicKey
	blockchainAddress string
}

func NewWallet() *Wallet {

	w := new(Wallet)
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	w.privateKey = privateKey
	w.publicKey = &w.privateKey.PublicKey

	h2 := sha256.New()
	h2.Write(w.publicKey.X.Bytes())
	h2.Write(w.publicKey.Y.Bytes())
	digest2 := h2.Sum(nil)

	h3 := sha256.New()
	h3.Write(digest2)
	digest3 := h3.Sum(nil)

	vd4 := make([]byte, 21)
	vd4[0] = 0x00
	copy(vd4[1:], digest3[:])

	h5 := sha256.New()
	h5.Write(vd4)
	digest5 := h5.Sum(nil)

	h6 := sha256.New()
	h6.Write(digest5)
	digest6 := h6.Sum(nil)

	chkSum := digest6[:4]

	dc8 := make([]byte, 25)
	copy(dc8[:21], vd4[:])
	copy(dc8[21:], chkSum[:])

	address := base58.Encode(dc8)
	w.blockchainAddress = address

	return w
}

func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

func (w *Wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x", w.privateKey.D.Bytes())
}

func (w *Wallet) PublicKey() *ecdsa.PublicKey {
	return w.publicKey
}

func (w *Wallet) PublicKeyStr() string {
	return fmt.Sprintf("%064x%064x", w.publicKey.X.Bytes(), w.publicKey.Y.Bytes())
}

func (w *Wallet) BlockchainAddress() string {
	return w.blockchainAddress
}

func (w *Wallet) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PrivateKey string `json:"private_key"`
		PublicKey  string `json:"public_key"`
		Address    string `json:"address"`
	}{
		PublicKey:  w.PublicKeyStr(),
		PrivateKey: w.PrivateKeyStr(),
		Address:    w.BlockchainAddress(),
	})
}

// -------------------------------------------------
type WalletTXN struct {
	senderPrivateKey           *ecdsa.PrivateKey
	senderPublicKey            *ecdsa.PublicKey
	senderBlockchainAddress    string
	recipientBlockchainAddress string
	value                      float32
}

func NewWalletTransaction(privKey *ecdsa.PrivateKey, pubKey *ecdsa.PublicKey, sender string, recipient string, value float32) *WalletTXN {
	return &WalletTXN{
		senderPrivateKey:           privKey,
		senderPublicKey:            pubKey,
		senderBlockchainAddress:    sender,
		recipientBlockchainAddress: recipient,
		value:                      value,
	}
}

func (wt *WalletTXN) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value     float32 `json:"value"`
	}{
		Sender:    wt.senderBlockchainAddress,
		Recipient: wt.recipientBlockchainAddress,
		Value:     wt.value,
	})
}

func (wt *WalletTXN) GenerateSignature() *utils.Signature {

	m, _ := json.Marshal(wt)
	hash := sha256.Sum256([]byte(m))

	r, s, err := ecdsa.Sign(rand.Reader, wt.senderPrivateKey, hash[:])

	if err != nil {
		fmt.Printf("%s : %+v", strings.Repeat("-", 30), err)
		panic(err)
	}

	return &utils.Signature{R: r, S: s}
}

// -------------------------------------------------
type WalletTXNRequest struct {
	SenderPrivateKey           *string  `json:"sender_private_key"`
	SenderBlockchainAddress    *string  `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string  `json:"recipient_blockchain_address"`
	SenderPublicKey            *string  `json:"sender_public_key"`
	Value                      *float32 `json:"value"`
}

func (tr *WalletTXNRequest) Validate() bool {

	if tr.SenderPrivateKey == nil {
		log.Panicln("SenderPrivateKey MISSING")
		return false
	}
	if tr.SenderBlockchainAddress == nil {
		log.Panicln("SenderBlockchainAddress MISSING")
		return false
	}
	if tr.RecipientBlockchainAddress == nil {
		log.Panicln("RecipientBlockchainAddress MISSING")
		return false
	}
	if tr.SenderPublicKey == nil {
		log.Panicln("SenderPublicKey MISSING")
		return false
	}
	if tr.Value == nil {
		log.Panicln("Value MISSING")
		return false
	}

	if tr.SenderPrivateKey == nil ||
		tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Value == nil {
		return false
	}
	return true
}
