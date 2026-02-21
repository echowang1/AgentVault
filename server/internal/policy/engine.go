package policy

import (
	"context"
	"math/big"
	"strings"
	"time"
)

type Engine struct {
	storage Storage
}

func NewPolicyEngine(storage Storage) (PolicyEngine, error) {
	return &Engine{storage: storage}, nil
}

func (e *Engine) Check(ctx context.Context, req *SignRequest) error {
	if req == nil || req.Value == nil || req.Value.Sign() < 0 {
		return ErrInvalidAmount
	}
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}

	policy, err := e.storage.GetPolicy(ctx, req.WalletID)
	if err != nil {
		if err == ErrPolicyNotFound {
			return nil
		}
		return err
	}

	if policy.SingleTxLimit != nil && req.Value.Cmp(policy.SingleTxLimit) > 0 {
		return ErrExceedsSingleTxLimit
	}
	if len(policy.Whitelist) > 0 && !isWhitelisted(policy.Whitelist, req.To) {
		return ErrAddressNotWhitelisted
	}
	if policy.StartTime != nil && req.Timestamp.Before(*policy.StartTime) {
		return ErrOutsideTimeWindow
	}
	if policy.EndTime != nil && req.Timestamp.After(*policy.EndTime) {
		return ErrOutsideTimeWindow
	}

	if policy.DailyLimit != nil || policy.DailyTxLimit > 0 {
		usage, err := e.storage.GetDailyUsage(ctx, req.WalletID, req.Timestamp)
		if err != nil {
			return err
		}
		if policy.DailyLimit != nil {
			newTotal := new(big.Int).Add(usage.TotalAmount, req.Value)
			if newTotal.Cmp(policy.DailyLimit) > 0 {
				return ErrExceedsDailyLimit
			}
		}
		if policy.DailyTxLimit > 0 && usage.TxCount >= policy.DailyTxLimit {
			return ErrExceedsDailyTxLimit
		}
	}

	return nil
}

func (e *Engine) SetPolicy(ctx context.Context, policy *Policy) error {
	return e.storage.SetPolicy(ctx, policy)
}

func (e *Engine) GetPolicy(ctx context.Context, walletID string) (*Policy, error) {
	return e.storage.GetPolicy(ctx, walletID)
}

func (e *Engine) DeletePolicy(ctx context.Context, walletID string) error {
	return e.storage.DeletePolicy(ctx, walletID)
}

func (e *Engine) GetDailyUsage(ctx context.Context, walletID string) (*DailyUsage, error) {
	return e.storage.GetDailyUsage(ctx, walletID, time.Now().UTC())
}

func (e *Engine) IncrementUsage(ctx context.Context, walletID string, amount *big.Int, ts time.Time) error {
	if amount == nil || amount.Sign() < 0 {
		return ErrInvalidAmount
	}
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	if err := e.storage.ResetDailyUsage(ctx, walletID, ts); err != nil {
		return err
	}
	return e.storage.IncrementUsage(ctx, walletID, ts, amount)
}

func isWhitelisted(whitelist []string, address string) bool {
	normalized := strings.ToLower(strings.TrimSpace(address))
	for _, addr := range whitelist {
		if strings.ToLower(strings.TrimSpace(addr)) == normalized {
			return true
		}
	}
	return false
}
