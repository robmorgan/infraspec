package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

// KeyPair is a public and private key pair that can be used for SSH access.
type KeyPair struct {
	PublicKey  string
	PrivateKey string
}

// GenerateRSAKeyPair generates an RSA key pair and return the public and private keys.
func GenerateRSAKeyPair(keySize int) (*KeyPair, error) {
	rsaKeyPair, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, err
	}

	// Extract the private key
	keyPemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKeyPair),
	}

	keyPem := string(pem.EncodeToMemory(keyPemBlock))

	// Extract the public key
	sshPubKey, err := ssh.NewPublicKey(rsaKeyPair.Public())
	if err != nil {
		return nil, err
	}

	sshPubKeyBytes := ssh.MarshalAuthorizedKey(sshPubKey)
	sshPubKeyStr := string(sshPubKeyBytes)

	// Return
	return &KeyPair{PublicKey: sshPubKeyStr, PrivateKey: keyPem}, nil
}
