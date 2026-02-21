package tss

import (
	"context"
	"encoding/base64"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKeyGenerator(t *testing.T) {
	keygen, err := NewKeyGenerator()
	require.NoError(t, err)
	assert.NotNil(t, keygen)
}

func TestGenerateKey_ValidResult(t *testing.T) {
	keygen, err := NewKeyGenerator()
	require.NoError(t, err)

	result, err := keygen.GenerateKey(context.Background())
	require.NoError(t, err)

	assert.Regexp(t, regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`), result.Address)
	assert.NotEmpty(t, result.PublicKey)
	assert.NotEmpty(t, result.Shard1)
	assert.NotEmpty(t, result.Shard2ID)

	shard1Bytes, err := base64.StdEncoding.DecodeString(result.Shard1)
	require.NoError(t, err)
	assert.NotEmpty(t, shard1Bytes)

	shard, err := unmarshalShare(shard1Bytes)
	require.NoError(t, err)
	assert.NotNil(t, shard.Xi)
	assert.Equal(t, "agent", shard.PartyID)
}

func TestGenerateKey_StoresShard2Encrypted(t *testing.T) {
	g, err := NewKeyGenerator()
	require.NoError(t, err)

	impl, ok := g.(*keyGenerator)
	require.True(t, ok)

	result, err := impl.GenerateKey(context.Background())
	require.NoError(t, err)

	impl.mu.RLock()
	entry, ok := impl.storage[result.Shard2ID]
	impl.mu.RUnlock()
	require.True(t, ok)
	assert.NotEmpty(t, entry.Ciphertext)
	assert.NotEmpty(t, entry.Nonce)

	plain, err := impl.loadAndDecryptShard2(result.Shard2ID)
	require.NoError(t, err)
	assert.NotEmpty(t, plain)

	share, err := unmarshalShare(plain)
	require.NoError(t, err)
	assert.NotNil(t, share.Xi)
	assert.Equal(t, "server", share.PartyID)
}

func TestGenerateKey_ContextCanceled(t *testing.T) {
	g, err := NewKeyGenerator()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := g.GenerateKey(ctx)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrContextCanceled)
}
