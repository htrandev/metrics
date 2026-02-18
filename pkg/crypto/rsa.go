package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func loadFile(filename string) ([]byte, error) {
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("crypto: read file: %w", err)
	}

	keyPemBlock, _ := pem.Decode(keyBytes)
	if keyPemBlock == nil {
		return nil, fmt.Errorf("crypto: key not found")
	}

	return keyPemBlock.Bytes, nil
}

// PrivateKey returns private key from given file.
func PrivateKey(filename string) (*rsa.PrivateKey, error) {
	block, err := loadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("crypto: load private key file: %w", err)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: parse private key: %w", err)
	}

	return privateKey, nil
}

// PublicKey returns private key from given file.
func PublicKey(filename string) (*rsa.PublicKey, error) {
	block, err := loadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("crypto: load public key file: %w", err)
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: parse public key: %w", err)
	}

	return publicKey, nil
}

// Encrypt encrypts payload with public key.
func Encrypt(key *rsa.PublicKey, payload []byte) ([]byte, error) {
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, key, payload, nil)
	if err != nil {
		return nil, fmt.Errorf("crypto: encrypt payload: %w", err)
	}
	return encrypted, nil
}

// Decrypt decrypts payload with private key.
func Decrypt(key *rsa.PrivateKey, payload []byte) ([]byte, error) {
	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, payload, nil)
	if err != nil {
		return nil, fmt.Errorf("crypto: decrypt payload: %w", err)
	}
	return decrypted, nil
}
