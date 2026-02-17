package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// Genetare creates two files: private.pem and public.pem
func Generate(path string) error {
	if path == "" {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current dir: %w", err)
		}
		path = dir
	}

	privatePath := filepath.Join(path, "private.pem")
	publicPath := filepath.Join(path, "public.pem")

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generating private key: %w", err)
	}

	privateFile, err := os.Create(privatePath)
	if err != nil {
		return fmt.Errorf("create private key file: %w", err)
	}
	defer privateFile.Close()
	
	privatePEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privateFile, privatePEM); err != nil {
		return fmt.Errorf("encode private key: %w", err)
	}
	
	publicFile, err := os.Create(publicPath)
	if err != nil {
		return fmt.Errorf("create public key file: %w", err)
	}
	defer publicFile.Close()

	publicBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	publicPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicBytes,
	}
	if err := pem.Encode(publicFile, publicPEM); err != nil {
		return fmt.Errorf("encode public key: %w", err)
	}

	fmt.Printf("Key successfully created:\n\tpublic: [%s],\n\tprivate: [%s] \n", publicPath, privatePath)
	return nil
}
