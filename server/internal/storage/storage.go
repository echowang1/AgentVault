package storage

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidKeySize    = errors.New("invalid key size: must be 32 bytes")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrNotFound          = errors.New("record not found")
)

// ShardStorage stores encrypted shard 2 payloads.
type ShardStorage interface {
	Store(ctx context.Context, id string, shard2 []byte) error
	Load(ctx context.Context, id string) ([]byte, error)
	Exists(ctx context.Context, id string) (bool, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]string, error)
}

// WalletInfo describes wallet metadata persisted by the service.
type WalletInfo struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	PublicKey string    `json:"public_key"`
	Shard2ID  string    `json:"shard2_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WalletStorage stores wallet metadata.
type WalletStorage interface {
	Create(ctx context.Context, info *WalletInfo) error
	GetByAddress(ctx context.Context, address string) (*WalletInfo, error)
	GetByID(ctx context.Context, id string) (*WalletInfo, error)
	Update(ctx context.Context, info *WalletInfo) error
	Delete(ctx context.Context, id string) error
}

// Encryptor encrypts/decrypts shard payloads.
type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}
