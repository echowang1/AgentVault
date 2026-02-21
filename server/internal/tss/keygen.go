package tss

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	tsskeygen "github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

const (
	partyCount = 2
	threshold  = 1
)

type keyShareData struct {
	ShareID   string
	PartyID   string
	Xi        *big.Int `json:"-"`
	CreatedAt time.Time
}

type encryptedShard struct {
	Nonce      []byte
	Ciphertext []byte
}

type keyGenerator struct {
	mu            sync.RWMutex
	storage       map[string]encryptedShard
	encryptionKey []byte
}

func NewKeyGenerator() (KeyGenerator, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, ErrKeyGenFailed
	}

	return &keyGenerator{
		storage:       make(map[string]encryptedShard),
		encryptionKey: key,
	}, nil
}

func (k *keyGenerator) GenerateKey(ctx context.Context) (*KeyGenResult, error) {
	return k.GenerateKeyWithProgress(ctx, nil)
}

func (k *keyGenerator) GenerateKeyWithProgress(ctx context.Context, callback ProgressCallback) (*KeyGenResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, ErrContextCanceled
	}

	emit(callback, "starting key generation", 5)
	saves, err := runGG18Keygen(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrContextCanceled
		}
		return nil, ErrKeyGenFailed
	}

	emit(callback, "encoding shards", 70)
	shard1Raw, err := marshalShare(keyShareData{
		ShareID:   randomID("s1"),
		PartyID:   "agent",
		Xi:        saves[0].Xi,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return nil, ErrKeyGenFailed
	}

	shard2Raw, err := marshalShare(keyShareData{
		ShareID:   randomID("s2"),
		PartyID:   "server",
		Xi:        saves[1].Xi,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return nil, ErrKeyGenFailed
	}

	emit(callback, "storing server shard", 85)
	shard2ID := randomID("shard2")
	if err := k.saveEncryptedShard2(shard2ID, shard2Raw); err != nil {
		return nil, ErrKeyGenFailed
	}

	privD := new(big.Int).Add(saves[0].Xi, saves[1].Xi)
	privD.Mod(privD, gethcrypto.S256().Params().N)
	pub := privateScalarToPublicKey(privD)

	emit(callback, "completed", 100)
	return &KeyGenResult{
		Address:   publicKeyToAddress(pub),
		Shard1:    base64.StdEncoding.EncodeToString(shard1Raw),
		Shard2ID:  shard2ID,
		PublicKey: publicKeyToHex(pub),
	}, nil
}

func emit(cb ProgressCallback, step string, percent int) {
	if cb != nil {
		cb(KeyGenerateProgress{Step: step, Percent: percent})
	}
}

func runGG18Keygen(ctx context.Context) ([]tsskeygen.LocalPartySaveData, error) {
	if partyCount != 2 {
		return nil, ErrInvalidPartyCount
	}

	partyIDs := tss.GenerateTestPartyIDs(partyCount)
	peerCtx := tss.NewPeerContext(partyIDs)

	outCh := make(chan tss.Message, 64)
	endCh := make(chan tsskeygen.LocalPartySaveData, partyCount)
	errCh := make(chan error, partyCount)

	parties := make([]tss.Party, 0, partyCount)
	partyByID := make(map[string]tss.Party, partyCount)
	partyKey := make(map[string]*big.Int, partyCount)

	for _, pid := range partyIDs {
		params := tss.NewParameters(gethcrypto.S256(), peerCtx, pid, partyCount, threshold)
		party := tsskeygen.NewLocalParty(params, outCh, endCh)
		parties = append(parties, party)
		partyByID[pid.Id] = party
		partyKey[pid.Id] = pid.KeyInt()
	}

	for i := range parties {
		go func(p tss.Party) {
			if err := p.Start(); err != nil {
				errCh <- err
			}
		}(parties[i])
	}

	results := make(map[string]tsskeygen.LocalPartySaveData, partyCount)

	for len(results) < partyCount {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errCh:
			if err != nil {
				return nil, err
			}
		case msg := <-outCh:
			if msg == nil {
				continue
			}
			wireMsg, _, err := msg.WireBytes()
			if err != nil {
				return nil, err
			}

			if msg.IsBroadcast() || len(msg.GetTo()) == 0 {
				for _, p := range parties {
					if p.PartyID().Id == msg.GetFrom().Id {
						continue
					}
					if _, err := p.UpdateFromBytes(wireMsg, msg.GetFrom(), msg.IsBroadcast()); err != nil {
						return nil, err
					}
				}
				continue
			}

			for _, to := range msg.GetTo() {
				target, ok := partyByID[to.Id]
				if !ok {
					return nil, ErrKeyGenFailed
				}
				if _, err := target.UpdateFromBytes(wireMsg, msg.GetFrom(), msg.IsBroadcast()); err != nil {
					return nil, err
				}
			}
		case save := <-endCh:
			matched := ""
			for id, key := range partyKey {
				if save.ShareID != nil && key.Cmp(save.ShareID) == 0 {
					matched = id
					break
				}
			}
			if matched == "" {
				return nil, ErrKeyGenFailed
			}
			results[matched] = save
		}
	}

	ordered := make([]tsskeygen.LocalPartySaveData, 0, partyCount)
	for _, pid := range partyIDs {
		r, ok := results[pid.Id]
		if !ok {
			return nil, ErrKeyGenFailed
		}
		ordered = append(ordered, r)
	}
	return ordered, nil
}

func marshalShare(share keyShareData) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(share); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func unmarshalShare(data []byte) (*keyShareData, error) {
	var out keyShareData
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (k *keyGenerator) saveEncryptedShard2(id string, plain []byte) error {
	block, err := aes.NewCipher(k.encryptionKey)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	ciphertext := gcm.Seal(nil, nonce, plain, nil)

	k.mu.Lock()
	k.storage[id] = encryptedShard{Nonce: nonce, Ciphertext: ciphertext}
	k.mu.Unlock()
	return nil
}

func (k *keyGenerator) loadAndDecryptShard2(id string) ([]byte, error) {
	k.mu.RLock()
	entry, ok := k.storage[id]
	k.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("shard not found")
	}

	block, err := aes.NewCipher(k.encryptionKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, entry.Nonce, entry.Ciphertext, nil)
}

func privateScalarToPublicKey(d *big.Int) *ecdsa.PublicKey {
	x, y := gethcrypto.S256().ScalarBaseMult(d.Bytes())
	return &ecdsa.PublicKey{Curve: gethcrypto.S256(), X: x, Y: y}
}

func publicKeyToAddress(pub *ecdsa.PublicKey) string {
	return gethcrypto.PubkeyToAddress(*pub).Hex()
}

func publicKeyToHex(pub *ecdsa.PublicKey) string {
	return hex.EncodeToString(gethcrypto.FromECDSAPub(pub))
}

func randomID(prefix string) string {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		panic("crypto/rand unavailable")
	}
	return prefix + "_" + hex.EncodeToString(raw)
}
