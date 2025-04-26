package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LoadRSAKey(filename string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}

	var parsedKey any
	if block.Type == "RSA PRIVATE KEY" {
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	} else if block.Type == "PRIVATE KEY" {
		parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		if rsaKey, ok := parsedKey.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, fmt.Errorf("not an RSA private key")
	}

	return nil, fmt.Errorf("unsupported key type: %s", block.Type)
}

func main() {
	fmt.Println("Decryption starting...")

	// Load RSA private key
	privateKey, err := LoadRSAKey("private.pem")
	if err != nil {
		panic("Failed to load private key: " + err.Error())
	}

	// Read and decrypt AES key
	encAESKey, err := os.ReadFile("ENCRYPTED_KEY.bin")
	if err != nil {
		panic("Failed to read encrypted AES key: " + err.Error())
	}

	aesKey, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encAESKey)
	if err != nil {
		panic("Failed to decrypt AES key: " + err.Error())
	}

	// Initialize AES cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		panic("Failed to create AES cipher: " + err.Error())
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic("Failed to create GCM: " + err.Error())
	}

	// Walk the encrypted files
	targetDir := "/home/student/Documents/secret-folder"
	filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".enc") {
			return nil
		}

		fmt.Println("Decrypting:", path)

		encryptedData, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Failed to read:", path)
			return nil
		}

		nonceSize := gcm.NonceSize()
		if len(encryptedData) < nonceSize {
			fmt.Println("File too short:", path)
			return nil
		}

		nonce := encryptedData[:nonceSize]
		ciphertext := encryptedData[nonceSize:]
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			fmt.Println("Failed to decrypt:", path)
			return nil
		}

		newPath := strings.TrimSuffix(path, ".enc")
		err = os.WriteFile(newPath, plaintext, 0644)
		if err != nil {
			fmt.Println("Error writing decrypted file:", newPath)
			return nil
		}

		os.Remove(path)
		return nil
	})

	fmt.Println("All decryptable files have been processed.")
}


