package tss

import (
	"context"
	"encoding/hex"
	"testing"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyGenAndSignEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test in short mode")
	}

	kg, err := NewKeyGenerator()
	require.NoError(t, err)

	keyResult, err := kg.GenerateKey(context.Background())
	require.NoError(t, err)

	signer, err := NewSigner(kg.(ShardStorage))
	require.NoError(t, err)

	messageHash := gethcrypto.Keccak256Hash([]byte("Hello, Agent MPC Wallet!"))
	req := &SignRequest{
		Address:     keyResult.Address,
		MessageHash: hex.EncodeToString(messageHash.Bytes()),
		Shard1:      keyResult.Shard1,
		Shard2ID:    keyResult.Shard2ID,
	}

	sig, err := signer.Sign(context.Background(), req)
	require.NoError(t, err)

	fullBytes, err := hex.DecodeString(sig.FullSignature[2:])
	require.NoError(t, err)
	require.Len(t, fullBytes, 65)

	recoverSig := make([]byte, 65)
	copy(recoverSig, fullBytes)
	recoverSig[64] = sig.V - 27

	pubKey, err := gethcrypto.SigToPub(messageHash.Bytes(), recoverSig)
	require.NoError(t, err)
	recoveredAddr := gethcrypto.PubkeyToAddress(*pubKey).Hex()
	assert.Equal(t, keyResult.Address, recoveredAddr)
}
