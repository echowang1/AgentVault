package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

const envShardEncryptionKey = "SHARD_ENCRYPTION_KEY"

type AES256GCMEncryptor struct {
	key []byte
}

func NewAES256GCMEncryptor(key []byte) (*AES256GCMEncryptor, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKeySize
	}
	copied := make([]byte, 32)
	copy(copied, key)
	return &AES256GCMEncryptor{key: copied}, nil
}

func NewAES256GCMEncryptorFromEnv() (*AES256GCMEncryptor, error) {
	raw := os.Getenv(envShardEncryptionKey)
	if raw == "" {
		return nil, fmt.Errorf("%s is required", envShardEncryptionKey)
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", envShardEncryptionKey, err)
	}
	return NewAES256GCMEncryptor(key)
}

func (e *AES256GCMEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (e *AES256GCMEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidCiphertext
	}
	nonce := ciphertext[:nonceSize]
	enc := ciphertext[nonceSize:]
	out, err := gcm.Open(nil, nonce, enc, nil)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}
	return out, nil
}
