package policy

import (
	"context"
	"database/sql"
	"encoding/json"
	"math/big"
	"strings"
	"time"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(db *sql.DB) *SQLiteStorage {
	return &SQLiteStorage{db: db}
}

func (s *SQLiteStorage) SetPolicy(ctx context.Context, p *Policy) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	p.UpdatedAt = time.Now().UTC()
	wl, err := json.Marshal(normalizeWhitelist(p.Whitelist))
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO policies (wallet_id, single_tx_limit, daily_limit, whitelist, daily_tx_limit, start_time, end_time, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(wallet_id) DO UPDATE SET
			single_tx_limit=excluded.single_tx_limit,
			daily_limit=excluded.daily_limit,
			whitelist=excluded.whitelist,
			daily_tx_limit=excluded.daily_tx_limit,
			start_time=excluded.start_time,
			end_time=excluded.end_time,
			updated_at=excluded.updated_at
	`, p.WalletID, bigToString(p.SingleTxLimit), bigToString(p.DailyLimit), string(wl), p.DailyTxLimit, timeToUnixPtr(p.StartTime), timeToUnixPtr(p.EndTime), p.CreatedAt.Unix(), p.UpdatedAt.Unix())
	return err
}

func (s *SQLiteStorage) GetPolicy(ctx context.Context, walletID string) (*Policy, error) {
	var (
		single, daily, wlJSON string
		dailyTxLimit          int
		startAt, endAt        sql.NullInt64
		createdAt, updatedAt  int64
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT single_tx_limit, daily_limit, whitelist, daily_tx_limit, start_time, end_time, created_at, updated_at
		FROM policies WHERE wallet_id = ?
	`, walletID).Scan(&single, &daily, &wlJSON, &dailyTxLimit, &startAt, &endAt, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPolicyNotFound
		}
		return nil, err
	}

	whitelist := make([]string, 0)
	if wlJSON != "" {
		_ = json.Unmarshal([]byte(wlJSON), &whitelist)
	}

	p := &Policy{
		WalletID:      walletID,
		SingleTxLimit: stringToBig(single),
		DailyLimit:    stringToBig(daily),
		Whitelist:     whitelist,
		DailyTxLimit:  dailyTxLimit,
		CreatedAt:     time.Unix(createdAt, 0).UTC(),
		UpdatedAt:     time.Unix(updatedAt, 0).UTC(),
	}
	if startAt.Valid {
		t := time.Unix(startAt.Int64, 0).UTC()
		p.StartTime = &t
	}
	if endAt.Valid {
		t := time.Unix(endAt.Int64, 0).UTC()
		p.EndTime = &t
	}
	return p, nil
}

func (s *SQLiteStorage) DeletePolicy(ctx context.Context, walletID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM policies WHERE wallet_id = ?`, walletID)
	return err
}

func (s *SQLiteStorage) GetDailyUsage(ctx context.Context, walletID string, date time.Time) (*DailyUsage, error) {
	day := date.UTC().Format("2006-01-02")
	var totalAmount string
	var txCount int
	err := s.db.QueryRowContext(ctx, `SELECT total_amount, tx_count FROM daily_usage WHERE wallet_id = ? AND date = ?`, walletID, day).Scan(&totalAmount, &txCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return &DailyUsage{WalletID: walletID, Date: day, TotalAmount: big.NewInt(0), TxCount: 0}, nil
		}
		return nil, err
	}
	return &DailyUsage{WalletID: walletID, Date: day, TotalAmount: stringToBig(totalAmount), TxCount: txCount}, nil
}

func (s *SQLiteStorage) IncrementUsage(ctx context.Context, walletID string, date time.Time, amount *big.Int) error {
	usage, err := s.GetDailyUsage(ctx, walletID, date)
	if err != nil {
		return err
	}
	newTotal := new(big.Int).Add(usage.TotalAmount, amount)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO daily_usage (wallet_id, date, total_amount, tx_count, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(wallet_id, date) DO UPDATE SET
			total_amount=excluded.total_amount,
			tx_count=excluded.tx_count,
			updated_at=excluded.updated_at
	`, walletID, usage.Date, newTotal.String(), usage.TxCount+1, time.Now().UTC().Unix())
	return err
}

func (s *SQLiteStorage) ResetDailyUsage(ctx context.Context, walletID string, date time.Time) error {
	day := date.UTC().Format("2006-01-02")
	_, err := s.db.ExecContext(ctx, `DELETE FROM daily_usage WHERE wallet_id = ? AND date <> ?`, walletID, day)
	return err
}

func normalizeWhitelist(in []string) []string {
	out := make([]string, 0, len(in))
	for _, addr := range in {
		a := strings.ToLower(strings.TrimSpace(addr))
		if a != "" {
			out = append(out, a)
		}
	}
	return out
}

func bigToString(v *big.Int) string {
	if v == nil {
		return ""
	}
	return v.String()
}

func stringToBig(v string) *big.Int {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	n := new(big.Int)
	if _, ok := n.SetString(v, 10); !ok {
		return nil
	}
	return n
}

func timeToUnixPtr(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.UTC().Unix()
}
