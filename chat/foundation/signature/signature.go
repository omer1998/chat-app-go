package signature

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

type SigRSV struct {
	V *big.Int
	R *big.Int
	S *big.Int
}

func Sign(value any, pvKey *ecdsa.PrivateKey) (*SigRSV, error) {
	// hashing data
	dataHash := hash(value)

	sigBytes, err := crypto.Sign(dataHash, pvKey)
	if err != nil {
		return nil, fmt.Errorf("error creating signature: %w", err)
	}
	//here we need to convert this signature byte into R S V value
	return toSigRSV(sigBytes), nil
}

// GetAddressFromSignature get the public address from the signature
func GetAddressFromSignature(value any, sigRSV *SigRSV) (string, error) {
	dataHash := hash(value)

	pubKey, err := crypto.SigToPub(dataHash, toSigBytes(sigRSV))
	if err != nil {
		return "", fmt.Errorf("err sigToPub: %w", err)
	}
	address := crypto.PubkeyToAddress(*pubKey)
	return address.String(), nil
}
func hash(value any) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	hashData := crypto.Keccak256(data)
	return hashData

}

func toSigRSV(signature []byte) *SigRSV {

	r := big.NewInt(0).SetBytes(signature[0:32])
	s := big.NewInt(0).SetBytes(signature[32:64])
	v := big.NewInt(0).SetBytes([]byte{signature[64]})

	return &SigRSV{
		R: r,
		S: s,
		V: v,
	}

}

func toSigBytes(sigRSV *SigRSV) []byte {
	sigBytes := make([]byte, crypto.SignatureLength)

	rBytes := make([]byte, 32)
	sigRSV.R.FillBytes(rBytes)
	copy(sigBytes, rBytes)

	sBytes := make([]byte, 32)
	sigRSV.S.FillBytes(sBytes)
	copy(sigBytes[32:], sBytes)

	sigBytes[64] = byte(sigRSV.V.Uint64())
	return sigBytes
}
