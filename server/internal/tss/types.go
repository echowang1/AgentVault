package tss

import (
	"context"
	"errors"

	tsskeygen "github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
)

var (
	ErrKeyGenFailed      = errors.New("key generation failed")
	ErrInvalidPartyCount = errors.New("party count must be 2")
	ErrContextCanceled   = errors.New("operation canceled")

	ErrInvalidHash      = errors.New("message hash must be 32 bytes")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrShardNotFound    = errors.New("shard not found")
	ErrShardMismatch    = errors.New("shards mismatch")
	ErrSignFailed       = errors.New("sign failed")
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

// SignRequest is a signing request from caller.
type SignRequest struct {
	Address     string
	MessageHash string
	Shard1      string
	Shard2ID    string
}

// Signature is an ethereum style ECDSA signature.
type Signature struct {
	R             string
	S             string
	V             uint8
	FullSignature string
}

// SignProgress represents a signing progress update.
type SignProgress struct {
	Step    string
	Percent int
}

// SignProgressCallback receives signing progress updates.
type SignProgressCallback func(progress SignProgress)

// Signer signs messages with 2-of-2 key shares.
type Signer interface {
	Sign(ctx context.Context, req *SignRequest) (*Signature, error)
	SignWithProgress(ctx context.Context, req *SignRequest, callback SignProgressCallback) (*Signature, error)
	SignBatch(ctx context.Context, reqs []*SignRequest) ([]*Signature, error)
}

// ShardStorage loads encrypted shard2 payload by id.
type ShardStorage interface {
	LoadShard2(ctx context.Context, shard2ID string) ([]byte, error)
}

// ShareResolver resolves in-memory TSS save data by cache id.
type ShareResolver interface {
	LoadSaveData(ctx context.Context, cacheID string) (tsskeygen.LocalPartySaveData, error)
}
