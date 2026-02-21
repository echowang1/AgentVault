package tss

import (
	"context"
	"encoding/hex"
	"testing"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyGenAndSignIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test in short mode")
	}

	kg, err := NewKeyGenerator()
	require.NoError(t, err)

	result, err := kg.GenerateKey(context.Background())
	require.NoError(t, err)

	signer, err := NewSigner(kg.(ShardStorage))
	require.NoError(t, err)

	hash := gethcrypto.Keccak256([]byte("integration message"))
	sig, err := signer.Sign(context.Background(), &SignRequest{
		Address:     result.Address,
		MessageHash: hex.EncodeToString(hash),
		Shard1:      result.Shard1,
		Shard2ID:    result.Shard2ID,
	})
	require.NoError(t, err)
	assert.Contains(t, []uint8{27, 28}, sig.V)

	raw, err := hex.DecodeString(sig.FullSignature[2:])
	require.NoError(t, err)
	recoverSig := make([]byte, 65)
	copy(recoverSig, raw)
	recoverSig[64] = sig.V - 27

	pub, err := gethcrypto.SigToPub(hash, recoverSig)
	require.NoError(t, err)
	assert.Equal(t, result.Address, gethcrypto.PubkeyToAddress(*pub).Hex())
}
