package policy

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/echowang1/agent-vault/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEngine(t *testing.T) PolicyEngine {
	t.Helper()
	encryptor, err := storage.NewAES256GCMEncryptor(make([]byte, 32))
	require.NoError(t, err)

	sqlStore, err := storage.NewSQLiteStorage(":memory:", encryptor)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlStore.Close() })

	policyStore := NewSQLiteStorage(sqlStore.DB())
	engine, err := NewPolicyEngine(policyStore)
	require.NoError(t, err)
	return engine
}

func TestCheck_SingleTxLimit(t *testing.T) {
	engine := setupTestEngine(t)
	ctx := context.Background()

	err := engine.SetPolicy(ctx, &Policy{WalletID: "wallet-1", SingleTxLimit: big.NewInt(100)})
	require.NoError(t, err)

	err = engine.Check(ctx, &SignRequest{WalletID: "wallet-1", To: "0xabc", Value: big.NewInt(50), Timestamp: time.Now()})
	assert.NoError(t, err)

	err = engine.Check(ctx, &SignRequest{WalletID: "wallet-1", To: "0xabc", Value: big.NewInt(200), Timestamp: time.Now()})
	assert.ErrorIs(t, err, ErrExceedsSingleTxLimit)
}

func TestCheck_Whitelist(t *testing.T) {
	engine := setupTestEngine(t)
	ctx := context.Background()
	require.NoError(t, engine.SetPolicy(ctx, &Policy{WalletID: "wallet-1", Whitelist: []string{"0xgood"}}))

	err := engine.Check(ctx, &SignRequest{WalletID: "wallet-1", To: "0xGOOD", Value: big.NewInt(1), Timestamp: time.Now()})
	assert.NoError(t, err)

	err = engine.Check(ctx, &SignRequest{WalletID: "wallet-1", To: "0xbad", Value: big.NewInt(1), Timestamp: time.Now()})
	assert.ErrorIs(t, err, ErrAddressNotWhitelisted)
}

func TestCheck_DailyLimit(t *testing.T) {
	engine := setupTestEngine(t)
	ctx := context.Background()
	require.NoError(t, engine.SetPolicy(ctx, &Policy{WalletID: "wallet-1", DailyLimit: big.NewInt(100)}))
	require.NoError(t, engine.IncrementUsage(ctx, "wallet-1", big.NewInt(80), time.Now()))

	err := engine.Check(ctx, &SignRequest{WalletID: "wallet-1", To: "0xabc", Value: big.NewInt(10), Timestamp: time.Now()})
	assert.NoError(t, err)

	err = engine.Check(ctx, &SignRequest{WalletID: "wallet-1", To: "0xabc", Value: big.NewInt(30), Timestamp: time.Now()})
	assert.ErrorIs(t, err, ErrExceedsDailyLimit)
}

func TestCheck_DailyTxLimit(t *testing.T) {
	engine := setupTestEngine(t)
	ctx := context.Background()
	require.NoError(t, engine.SetPolicy(ctx, &Policy{WalletID: "wallet-1", DailyTxLimit: 2}))
	require.NoError(t, engine.IncrementUsage(ctx, "wallet-1", big.NewInt(1), time.Now()))
	require.NoError(t, engine.IncrementUsage(ctx, "wallet-1", big.NewInt(1), time.Now()))

	err := engine.Check(ctx, &SignRequest{WalletID: "wallet-1", To: "0xabc", Value: big.NewInt(1), Timestamp: time.Now()})
	assert.ErrorIs(t, err, ErrExceedsDailyTxLimit)
}

func TestCheck_TimeWindow(t *testing.T) {
	engine := setupTestEngine(t)
	ctx := context.Background()
	start := time.Date(2026, 2, 21, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 21, 18, 0, 0, 0, time.UTC)
	require.NoError(t, engine.SetPolicy(ctx, &Policy{WalletID: "wallet-1", StartTime: &start, EndTime: &end}))

	err := engine.Check(ctx, &SignRequest{WalletID: "wallet-1", To: "0xabc", Value: big.NewInt(1), Timestamp: time.Date(2026, 2, 21, 12, 0, 0, 0, time.UTC)})
	assert.NoError(t, err)

	err = engine.Check(ctx, &SignRequest{WalletID: "wallet-1", To: "0xabc", Value: big.NewInt(1), Timestamp: time.Date(2026, 2, 21, 20, 0, 0, 0, time.UTC)})
	assert.ErrorIs(t, err, ErrOutsideTimeWindow)
}

func TestGetDailyUsage(t *testing.T) {
	engine := setupTestEngine(t)
	ctx := context.Background()
	require.NoError(t, engine.IncrementUsage(ctx, "wallet-1", big.NewInt(5), time.Now()))
	require.NoError(t, engine.IncrementUsage(ctx, "wallet-1", big.NewInt(7), time.Now()))

	usage, err := engine.GetDailyUsage(ctx, "wallet-1")
	require.NoError(t, err)
	assert.Equal(t, int64(12), usage.TotalAmount.Int64())
	assert.Equal(t, 2, usage.TxCount)
}
