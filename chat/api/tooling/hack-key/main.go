package main

import (
	"fmt"
	"log"
	"math/big"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/omer1998/chat-app-go.git/chat/foundation/signature"
)

const ethID = 27

type IncMessage struct {
	ToId string `json:"toId"`
	// From User   `json:"from"`
	Msg string `json:"msg"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

}

func run() error {
	basePath := "C:/Users/master/Desktop/chat-app-go/chat/zarf/client"
	filePath := filepath.Join(basePath, "key.ecdsa")

	pvKey, err := crypto.LoadECDSA(filePath)
	if err != nil {
		pvKey, err = crypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("error generating key: %w", err)
		}

		if err := crypto.SaveECDSA(filePath, pvKey); err != nil {
			return fmt.Errorf("error saving private key to file: %w", err)
		}
	}

	//data to be send by the client (which the message and its metadat)
	data := struct {
		ToId   string `json:"toId"`
		Msg    string `json:"msg"`
		FromId string `json:"fromId"`
		Nonce  int    `json:"nonce"`
	}{
		ToId:   "098u84309583",
		Msg:    "hello world",
		FromId: crypto.PubkeyToAddress(pvKey.PublicKey).String(),
		Nonce:  time.Now().Nanosecond(),
	}

	sigRSV, err := signature.Sign(data, pvKey)
	if err != nil {
		return err
	}

	// this sigRSV will get delivered to the server in a way or another
	// in the server we will be able to retrieve the public address from the data signature

	publicAddress, err := signature.GetAddressFromSignature(data, sigRSV)
	if err != nil {
		return err
	}

	fmt.Println("public key address from private key: ", crypto.PubkeyToAddress(pvKey.PublicKey).String())
	fmt.Println("public key address after signature: ", publicAddress)

	return nil
}
func toSignatureBytes(r, v, s *big.Int) []byte {
	sigBytes := make([]byte, crypto.SignatureLength)

	rBytes := make([]byte, 32)
	r.FillBytes(rBytes)
	copy(sigBytes, rBytes)
	sBytes := make([]byte, 32)
	s.FillBytes(sBytes)
	copy(sigBytes[32:], sBytes)

	sigBytes[64] = byte(v.Uint64() - ethID)

	return sigBytes
}

// toSignatureValues converts the signature into the r, s, v values.
func toSignatureValues(sig []byte) (v, r, s *big.Int) {
	r = big.NewInt(0).SetBytes(sig[:32])
	s = big.NewInt(0).SetBytes(sig[32:64])
	v = big.NewInt(0).SetBytes([]byte{sig[64] + ethID})

	return v, r, s
}
