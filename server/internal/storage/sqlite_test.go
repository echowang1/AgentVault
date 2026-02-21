package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *SQLiteStorage {
	t.Helper()
	key := make([]byte, 32)
	encryptor, err := NewAES256GCMEncryptor(key)
	require.NoError(t, err)

	s, err := NewSQLiteStorage(":memory:", encryptor)
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestStoreAndLoad(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	id := "test-id"
	payload := []byte("test-shard-data")
	require.NoError(t, s.Store(ctx, id, payload))

	loaded, err := s.Load(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, payload, loaded)
}

func TestEncryption(t *testing.T) {
	enc, err := NewAES256GCMEncryptor(make([]byte, 32))
	require.NoError(t, err)

	plaintext := []byte("sensitive-shard-data")
	ciphertext, err := enc.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := enc.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestInvalidKeyLength(t *testing.T) {
	_, err := NewAES256GCMEncryptor(make([]byte, 16))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidKeySize)
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	enc, err := NewAES256GCMEncryptor(make([]byte, 32))
	require.NoError(t, err)

	_, err = enc.Decrypt([]byte{1, 2, 3})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCiphertext)
}

func TestExistsListDeleteShard(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()
	require.NoError(t, s.Store(ctx, "a", []byte("A")))
	require.NoError(t, s.Store(ctx, "b", []byte("B")))

	exists, err := s.Exists(ctx, "a")
	require.NoError(t, err)
	assert.True(t, exists)

	ids, err := s.List(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, ids)

	require.NoError(t, s.Delete(ctx, "a"))
	exists, err = s.Exists(ctx, "a")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestWalletCRUD(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	info := &WalletInfo{
		ID:        "wallet-1",
		Address:   "0x1234567890123456789012345678901234567890",
		PublicKey: "0xabc",
		Shard2ID:  "shard-2",
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, s.Create(ctx, info))

	byAddr, err := s.GetByAddress(ctx, info.Address)
	require.NoError(t, err)
	assert.Equal(t, info.Address, byAddr.Address)

	byID, err := s.GetByID(ctx, info.ID)
	require.NoError(t, err)
	assert.Equal(t, info.ID, byID.ID)

	byID.PublicKey = "0xdef"
	byID.UpdatedAt = now.Add(time.Minute)
	require.NoError(t, s.Update(ctx, byID))

	updated, err := s.GetByID(ctx, info.ID)
	require.NoError(t, err)
	assert.Equal(t, "0xdef", updated.PublicKey)

	require.NoError(t, s.Delete(ctx, info.ID))
	_, err = s.GetByID(ctx, info.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}
