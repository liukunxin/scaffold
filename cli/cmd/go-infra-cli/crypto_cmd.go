package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func runKeygen(args []string) error {
	fs := flag.NewFlagSet("keygen", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	hexKey := hex.EncodeToString(key)
	fmt.Printf("Key (hex): %s\n\n", hexKey)
	fmt.Println("Set as environment variable:")
	fmt.Printf("  export CONFIG_ENCRYPT_KEY=%s\n", hexKey)
	return nil
}

func runEncrypt(args []string) error {
	fs := flag.NewFlagSet("encrypt", flag.ContinueOnError)
	keyHex := fs.String("key", "", "AES-256 key in hex (64 chars)")
	keyEnv := fs.String("key-env", "CONFIG_ENCRYPT_KEY", "environment variable holding the key")
	value := fs.String("value", "", "plaintext value to encrypt")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *value == "" {
		return errors.New("--value is required")
	}

	key, err := resolveKey(*keyHex, *keyEnv)
	if err != nil {
		return err
	}

	encrypted, err := encryptValue(key, *value)
	if err != nil {
		return err
	}
	fmt.Println(encrypted)
	return nil
}

func runDecrypt(args []string) error {
	fs := flag.NewFlagSet("decrypt", flag.ContinueOnError)
	keyHex := fs.String("key", "", "AES-256 key in hex (64 chars)")
	keyEnv := fs.String("key-env", "CONFIG_ENCRYPT_KEY", "environment variable holding the key")
	value := fs.String("value", "", "ENC(...) value to decrypt")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *value == "" {
		return errors.New("--value is required")
	}

	key, err := resolveKey(*keyHex, *keyEnv)
	if err != nil {
		return err
	}

	encoded := *value
	pattern := regexp.MustCompile(`^ENC\(([A-Za-z0-9+/=]+)\)$`)
	if m := pattern.FindStringSubmatch(encoded); len(m) == 2 {
		encoded = m[1]
	}

	decrypted, err := decryptValue(key, encoded)
	if err != nil {
		return err
	}
	fmt.Println(decrypted)
	return nil
}

func resolveKey(keyHex, keyEnv string) ([]byte, error) {
	raw := strings.TrimSpace(keyHex)
	if raw == "" {
		raw = os.Getenv(keyEnv)
	}
	if raw == "" {
		return nil, fmt.Errorf("no key provided: set -key or environment variable %s", keyEnv)
	}
	return parseKeyStr(raw)
}

func parseKeyStr(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) == 64 {
		key, err := hex.DecodeString(raw)
		if err == nil && len(key) == 32 {
			return key, nil
		}
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err == nil && len(key) == 32 {
		return key, nil
	}
	return nil, fmt.Errorf("invalid key format: expect 32-byte hex (64 chars) or base64")
}

func encryptValue(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return "ENC(" + base64.StdEncoding.EncodeToString(ciphertext) + ")", nil
}

func decryptValue(key []byte, encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt failed (wrong key?): %w", err)
	}
	return string(plaintext), nil
}
