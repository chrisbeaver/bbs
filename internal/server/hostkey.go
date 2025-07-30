package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// GenerateHostKey creates a new RSA host key and saves it to file
func GenerateHostKey(filename string) (ssh.Signer, error) {
	// Check if key file already exists
	if _, err := os.Stat(filename); err == nil {
		// Load existing key
		keyBytes, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read existing host key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse existing host key: %w", err)
		}

		return signer, nil
	}

	// Generate new RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Convert to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Save to file
	keyFile, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create host key file: %w", err)
	}
	defer keyFile.Close()

	// Set restrictive permissions
	if err := keyFile.Chmod(0600); err != nil {
		return nil, fmt.Errorf("failed to set host key permissions: %w", err)
	}

	if err := pem.Encode(keyFile, privateKeyPEM); err != nil {
		return nil, fmt.Errorf("failed to write host key: %w", err)
	}

	// Create SSH signer
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH signer: %w", err)
	}

	return signer, nil
}
