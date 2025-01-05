package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
)

func Encrypt(plaintext, key []byte) (string, error) {
	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		slog.Info(err.Error())
		return "", err
	}

	// Create GCM (Galois/Counter Mode) for encryption
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		slog.Info(err.Error())
		return "", err
	}

	// Create a nonce
	nonce := make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		slog.Info(err.Error())
		return "", err
	}

	// Encrypt the data
	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)

	// Combine nonce and ciphertext
	finalData := append(nonce, ciphertext...)

	// Encode the encrypted data in base64 for easier storage or transmission
	return base64.StdEncoding.EncodeToString(finalData), nil
}

// Decrypt decrypts the base64-encoded ciphertext using AES-GCM with the provided key.
func Decrypt(encryptedData string, key []byte) (string, error) {
	// Decode the base64-encoded ciphertext
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		slog.Info(err.Error())
		return "", err
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		slog.Info(err.Error())
		return "", err
	}

	// Create GCM for decryption
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		slog.Info(err.Error())
		return "", err
	}

	// Extract nonce and ciphertext
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		slog.Info(err.Error())
		return "", err
	}

	return string(plaintext), nil
}
