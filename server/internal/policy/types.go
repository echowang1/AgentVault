package policy

import (
	"context"
	"errors"
	"math/big"
	"time"
)

var (
	ErrPolicyNotFound        = errors.New("policy not found")
	ErrExceedsSingleTxLimit  = errors.New("exceeds single transaction limit")
	ErrExceedsDailyLimit     = errors.New("exceeds daily limit")
	ErrExceedsDailyTxLimit   = errors.New("exceeds daily transaction limit")
	ErrAddressNotWhitelisted = errors.New("address not whitelisted")
	ErrOutsideTimeWindow     = errors.New("outside allowed time window")
	ErrInvalidAmount         = errors.New("invalid amount")
)

type Policy struct {
	ID            string
	WalletID      string
	SingleTxLimit *big.Int
	DailyLimit    *big.Int
	Whitelist     []string
	DailyTxLimit  int
	StartTime     *time.Time
	EndTime       *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type SignRequest struct {
	WalletID  string
	To        string
	Value     *big.Int
	Timestamp time.Time
}

type DailyUsage struct {
	WalletID    string
	Date        string
	TotalAmount *big.Int
	TxCount     int
}

type Storage interface {
	SetPolicy(ctx context.Context, policy *Policy) error
	GetPolicy(ctx context.Context, walletID string) (*Policy, error)
	DeletePolicy(ctx context.Context, walletID string) error
	GetDailyUsage(ctx context.Context, walletID string, date time.Time) (*DailyUsage, error)
	IncrementUsage(ctx context.Context, walletID string, date time.Time, amount *big.Int) error
	ResetDailyUsage(ctx context.Context, walletID string, date time.Time) error
}

type PolicyEngine interface {
	Check(ctx context.Context, req *SignRequest) error
	SetPolicy(ctx context.Context, policy *Policy) error
	GetPolicy(ctx context.Context, walletID string) (*Policy, error)
	DeletePolicy(ctx context.Context, walletID string) error
	GetDailyUsage(ctx context.Context, walletID string) (*DailyUsage, error)
	IncrementUsage(ctx context.Context, walletID string, amount *big.Int, ts time.Time) error
}
