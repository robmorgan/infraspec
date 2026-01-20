package ec2

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/google/uuid"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"golang.org/x/crypto/ssh"
)

func (s *EC2Service) createKeyPair(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	keyName, ok := params["KeyName"].(string)
	if !ok || keyName == "" {
		return s.errorResponse(400, "MissingParameter", "KeyName is required"), nil
	}

	// Check if key pair already exists
	var existing KeyPairInfo
	if err := s.state.Get(fmt.Sprintf("ec2:key-pairs:%s", keyName), &existing); err == nil {
		return s.errorResponse(400, "InvalidKeyPair.Duplicate", fmt.Sprintf("The keypair '%s' already exists", keyName)), nil
	}

	// Generate RSA key pair (4096 bits for security best practices)
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to generate key pair"), nil
	}

	// Encode private key to PEM
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Generate public key fingerprint
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to generate public key"), nil
	}
	fingerprint := ssh.FingerprintLegacyMD5(publicKey)

	keyPairId := fmt.Sprintf("key-%s", uuid.New().String()[:8])

	keyPairInfo := KeyPairInfo{
		KeyPairId:      &keyPairId,
		KeyName:        &keyName,
		KeyFingerprint: &fingerprint,
		KeyType:        KeyType("rsa"),
	}

	if err := s.state.Set(fmt.Sprintf("ec2:key-pairs:%s", keyName), &keyPairInfo); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store key pair"), nil
	}

	return s.createKeyPairResponse(keyPairId, keyName, fingerprint, string(privateKeyPEM))
}

func (s *EC2Service) describeKeyPairs(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	keyNames := s.parseKeyNames(params)

	var keyPairs []KeyPairInfo

	if len(keyNames) > 0 {
		for _, keyName := range keyNames {
			var keyPair KeyPairInfo
			if err := s.state.Get(fmt.Sprintf("ec2:key-pairs:%s", keyName), &keyPair); err != nil {
				return s.errorResponse(400, "InvalidKeyPair.NotFound", fmt.Sprintf("The key pair '%s' does not exist", keyName)), nil
			}
			keyPairs = append(keyPairs, keyPair)
		}
	} else {
		keys, err := s.state.List("ec2:key-pairs:")
		if err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to list key pairs"), nil
		}

		for _, key := range keys {
			var keyPair KeyPairInfo
			if err := s.state.Get(key, &keyPair); err == nil {
				keyPairs = append(keyPairs, keyPair)
			}
		}
	}

	return s.describeKeyPairsResponse(keyPairs)
}

func (s *EC2Service) deleteKeyPair(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	keyName, ok := params["KeyName"].(string)
	if !ok || keyName == "" {
		return s.errorResponse(400, "MissingParameter", "KeyName is required"), nil
	}

	var keyPair KeyPairInfo
	if err := s.state.Get(fmt.Sprintf("ec2:key-pairs:%s", keyName), &keyPair); err != nil {
		// AWS doesn't return an error if key pair doesn't exist
		return s.deleteKeyPairResponse()
	}

	s.state.Delete(fmt.Sprintf("ec2:key-pairs:%s", keyName))

	return s.deleteKeyPairResponse()
}

func (s *EC2Service) importKeyPair(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	keyName, ok := params["KeyName"].(string)
	if !ok || keyName == "" {
		return s.errorResponse(400, "MissingParameter", "KeyName is required"), nil
	}

	publicKeyMaterial, ok := params["PublicKeyMaterial"].(string)
	if !ok || publicKeyMaterial == "" {
		return s.errorResponse(400, "MissingParameter", "PublicKeyMaterial is required"), nil
	}

	// Check if key pair already exists
	var existingImport KeyPairInfo
	if err := s.state.Get(fmt.Sprintf("ec2:key-pairs:%s", keyName), &existingImport); err == nil {
		return s.errorResponse(400, "InvalidKeyPair.Duplicate", fmt.Sprintf("The keypair '%s' already exists", keyName)), nil
	}

	keyPairId := fmt.Sprintf("key-%s", uuid.New().String()[:8])
	fingerprint := fmt.Sprintf("%s:mock:fingerprint", keyPairId[:8])

	keyPairInfo := KeyPairInfo{
		KeyPairId:      &keyPairId,
		KeyName:        &keyName,
		KeyFingerprint: &fingerprint,
		KeyType:        KeyType("rsa"),
	}

	if err := s.state.Set(fmt.Sprintf("ec2:key-pairs:%s", keyName), &keyPairInfo); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store key pair"), nil
	}

	return s.importKeyPairResponse(keyPairId, keyName, fingerprint)
}
