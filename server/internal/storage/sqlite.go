package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrations embed.FS

type SQLiteStorage struct {
	db        *sql.DB
	encryptor Encryptor
}

func NewSQLiteStorage(dbPath string, encryptor Encryptor) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}

	s := &SQLiteStorage{db: db, encryptor: encryptor}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

func (s *SQLiteStorage) DB() *sql.DB {
	return s.db
}

func (s *SQLiteStorage) migrate() error {
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		content, err := migrations.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func (s *SQLiteStorage) Store(ctx context.Context, id string, shard2 []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	sealed, err := s.encryptor.Encrypt(shard2)
	if err != nil {
		return err
	}
	if len(sealed) < 12 {
		return ErrInvalidCiphertext
	}
	nonce := sealed[:12]
	ciphertext := sealed[12:]
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO key_shards (id, shard2_encrypted, nonce, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			shard2_encrypted = excluded.shard2_encrypted,
			nonce = excluded.nonce
	`, id, ciphertext, nonce, time.Now().UTC().Unix())
	return err
}

func (s *SQLiteStorage) Load(ctx context.Context, id string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var ciphertext, nonce []byte
	if err := s.db.QueryRowContext(ctx, `SELECT shard2_encrypted, nonce FROM key_shards WHERE id = ?`, id).Scan(&ciphertext, &nonce); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	combined := append(append([]byte{}, nonce...), ciphertext...)
	return s.encryptor.Decrypt(combined)
}

func (s *SQLiteStorage) Exists(ctx context.Context, id string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM key_shards WHERE id = ?`, id).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *SQLiteStorage) List(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM key_shards`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Strings(ids)
	return ids, nil
}

func (s *SQLiteStorage) Create(ctx context.Context, info *WalletInfo) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO wallets (id, address, public_key, shard2_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, info.ID, info.Address, info.PublicKey, info.Shard2ID, info.CreatedAt.Unix(), info.UpdatedAt.Unix())
	return err
}

func (s *SQLiteStorage) GetByAddress(ctx context.Context, address string) (*WalletInfo, error) {
	return s.scanWallet(ctx, `SELECT id, address, public_key, shard2_id, created_at, updated_at FROM wallets WHERE address = ?`, address)
}

func (s *SQLiteStorage) GetByID(ctx context.Context, id string) (*WalletInfo, error) {
	return s.scanWallet(ctx, `SELECT id, address, public_key, shard2_id, created_at, updated_at FROM wallets WHERE id = ?`, id)
}

func (s *SQLiteStorage) Update(ctx context.Context, info *WalletInfo) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE wallets
		SET address = ?, public_key = ?, shard2_id = ?, updated_at = ?
		WHERE id = ?
	`, info.Address, info.PublicKey, info.Shard2ID, info.UpdatedAt.Unix(), info.ID)
	return err
}

// Delete removes rows with matching id from both tables to satisfy both shard and wallet interfaces.
func (s *SQLiteStorage) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM key_shards WHERE id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM wallets WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLiteStorage) scanWallet(ctx context.Context, query, arg string) (*WalletInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var info WalletInfo
	var createdAtUnix, updatedAtUnix int64
	err := s.db.QueryRowContext(ctx, query, arg).Scan(
		&info.ID,
		&info.Address,
		&info.PublicKey,
		&info.Shard2ID,
		&createdAtUnix,
		&updatedAtUnix,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	info.CreatedAt = time.Unix(createdAtUnix, 0).UTC()
	info.UpdatedAt = time.Unix(updatedAtUnix, 0).UTC()
	return &info, nil
}

// LoadShard2 adapts storage for tss signer/keygen integration.
func (s *SQLiteStorage) LoadShard2(ctx context.Context, shard2ID string) ([]byte, error) {
	return s.Load(ctx, shard2ID)
}
