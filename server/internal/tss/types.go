package tss

import (
	"context"
	"errors"
)

var (
	ErrKeyGenFailed      = errors.New("key generation failed")
	ErrInvalidPartyCount = errors.New("party count must be 2")
	ErrContextCanceled   = errors.New("operation canceled")
)

// KeyGenResult is the public result returned to caller after a successful 2-of-2 key generation.
type KeyGenResult struct {
	Address   string
	Shard1    string
	Shard2ID  string
	PublicKey string
}

// KeyGenerateProgress represents a key generation progress update.
type KeyGenerateProgress struct {
	Step    string
	Percent int
}

// ProgressCallback receives key generation progress updates.
type ProgressCallback func(progress KeyGenerateProgress)

// KeyGenerator generates 2-of-2 ECDSA key shares.
type KeyGenerator interface {
	GenerateKey(ctx context.Context) (*KeyGenResult, error)
	GenerateKeyWithProgress(ctx context.Context, callback ProgressCallback) (*KeyGenResult, error)
}
