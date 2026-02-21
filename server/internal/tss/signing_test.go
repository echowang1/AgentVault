package tss

import (
	"context"
	"encoding/hex"
	"testing"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSigner(t *testing.T) {
	storage := NewMockShardStorage()
	s, err := NewSigner(storage)
	require.NoError(t, err)
	assert.NotNil(t, s)
}

func TestSign_ValidSignature(t *testing.T) {
	kg, err := NewKeyGenerator()
	require.NoError(t, err)

	result, err := kg.GenerateKey(context.Background())
	require.NoError(t, err)

	s, err := NewSigner(kg.(ShardStorage))
	require.NoError(t, err)

	hash := gethcrypto.Keccak256([]byte("test message"))
	req := &SignRequest{
		Address:     result.Address,
		MessageHash: hex.EncodeToString(hash),
		Shard1:      result.Shard1,
		Shard2ID:    result.Shard2ID,
	}

	sig, err := s.Sign(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, sig)

	assert.Len(t, sig.R, 64)
	assert.Len(t, sig.S, 64)
	assert.Contains(t, []uint8{27, 28}, sig.V)
	assert.Len(t, sig.FullSignature, 132)

	fullBytes, err := hex.DecodeString(sig.FullSignature[2:])
	require.NoError(t, err)
	require.Len(t, fullBytes, 65)

	assert.Equal(t, sig.R, hex.EncodeToString(fullBytes[:32]))
	assert.Equal(t, sig.S, hex.EncodeToString(fullBytes[32:64]))
	assert.Equal(t, sig.V, fullBytes[64])

	recoverSig := make([]byte, 65)
	copy(recoverSig, fullBytes)
	recoverSig[64] = sig.V - 27

	pub, err := gethcrypto.SigToPub(hash, recoverSig)
	require.NoError(t, err)
	recoveredAddr := gethcrypto.PubkeyToAddress(*pub).Hex()
	assert.Equal(t, result.Address, recoveredAddr)
}

func TestSign_InvalidHash(t *testing.T) {
	storage := NewMockShardStorage()
	s, err := NewSigner(storage)
	require.NoError(t, err)

	_, err = s.Sign(context.Background(), &SignRequest{
		Address:     "0x1234567890123456789012345678901234567890",
		MessageHash: "invalid",
		Shard1:      "invalid",
		Shard2ID:    "none",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidHash)
}

func TestSign_ShardNotFound(t *testing.T) {
	kg, err := NewKeyGenerator()
	require.NoError(t, err)
	result, err := kg.GenerateKey(context.Background())
	require.NoError(t, err)

	storage := NewMockShardStorage()
	s, err := NewSigner(storage)
	require.NoError(t, err)

	hash := gethcrypto.Keccak256([]byte("test"))
	_, err = s.Sign(context.Background(), &SignRequest{
		Address:     result.Address,
		MessageHash: hex.EncodeToString(hash),
		Shard1:      result.Shard1,
		Shard2ID:    result.Shard2ID,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShardNotFound)
}

func TestSignBatch(t *testing.T) {
	kg, err := NewKeyGenerator()
	require.NoError(t, err)
	result, err := kg.GenerateKey(context.Background())
	require.NoError(t, err)

	s, err := NewSigner(kg.(ShardStorage))
	require.NoError(t, err)

	h1 := gethcrypto.Keccak256([]byte("msg1"))
	h2 := gethcrypto.Keccak256([]byte("msg2"))
	reqs := []*SignRequest{
		{Address: result.Address, MessageHash: hex.EncodeToString(h1), Shard1: result.Shard1, Shard2ID: result.Shard2ID},
		{Address: result.Address, MessageHash: hex.EncodeToString(h2), Shard1: result.Shard1, Shard2ID: result.Shard2ID},
	}

	sigs, err := s.SignBatch(context.Background(), reqs)
	require.NoError(t, err)
	require.Len(t, sigs, 2)
	assert.NotEmpty(t, sigs[0].FullSignature)
	assert.NotEmpty(t, sigs[1].FullSignature)
}
