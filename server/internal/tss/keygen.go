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
	"github.com/echowang1/agent-vault/internal/storage"
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
	CacheID   string
	CreatedAt time.Time
}

type encryptedShard struct {
	Nonce      []byte
	Ciphertext []byte
}

type keyGenerator struct {
	mu            sync.RWMutex
	storage       map[string]encryptedShard
	saveDataCache map[string]tsskeygen.LocalPartySaveData
	persistent    storage.ShardStorage
	encryptionKey []byte
}

func NewKeyGenerator() (KeyGenerator, error) {
	return newKeyGenerator(nil)
}

func NewKeyGeneratorWithStorage(shardStore storage.ShardStorage) (KeyGenerator, error) {
	return newKeyGenerator(shardStore)
}

func newKeyGenerator(shardStore storage.ShardStorage) (KeyGenerator, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, ErrKeyGenFailed
	}

	return &keyGenerator{
		storage:       make(map[string]encryptedShard),
		saveDataCache: make(map[string]tsskeygen.LocalPartySaveData),
		persistent:    shardStore,
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
	shard1Save := saves[0]
	shard2Save := saves[1]
	normalizeLocalSaveData(&shard1Save)
	normalizeLocalSaveData(&shard2Save)
	shard1CacheID := randomID("save1")
	shard2CacheID := randomID("save2")
	k.storeSaveData(shard1CacheID, shard1Save)
	k.storeSaveData(shard2CacheID, shard2Save)

	shard1Raw, err := marshalShare(keyShareData{
		ShareID:   randomID("s1"),
		PartyID:   "agent",
		Xi:        shard1Save.Xi,
		CacheID:   shard1CacheID,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return nil, ErrKeyGenFailed
	}

	shard2Raw, err := marshalShare(keyShareData{
		ShareID:   randomID("s2"),
		PartyID:   "server",
		Xi:        shard2Save.Xi,
		CacheID:   shard2CacheID,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return nil, ErrKeyGenFailed
	}

	emit(callback, "storing server shard", 85)
	shard2ID := randomID("shard2")
	if k.persistent != nil {
		if err := k.persistent.Store(ctx, shard2ID, shard2Raw); err != nil {
			return nil, ErrKeyGenFailed
		}
	} else {
		if err := k.saveEncryptedShard2(shard2ID, shard2Raw); err != nil {
			return nil, ErrKeyGenFailed
		}
	}

	pub := saveDataToPublicKey(shard1Save)
	if pub == nil {
		return nil, ErrKeyGenFailed
	}

	emit(callback, "completed", 100)
	return &KeyGenResult{
		Address:   publicKeyToAddress(pub),
		Shard1:    base64.StdEncoding.EncodeToString(shard1Raw),
		Shard2ID:  shard2ID,
		PublicKey: publicKeyToHex(pub),
	}, nil
}

func (k *keyGenerator) LoadShard2(ctx context.Context, shard2ID string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, ErrContextCanceled
	}
	if k.persistent != nil {
		plain, err := k.persistent.Load(ctx, shard2ID)
		if err != nil {
			return nil, ErrShardNotFound
		}
		return plain, nil
	}
	plain, err := k.loadAndDecryptShard2(shard2ID)
	if err != nil {
		return nil, ErrShardNotFound
	}
	return plain, nil
}

func (k *keyGenerator) LoadSaveData(ctx context.Context, cacheID string) (tsskeygen.LocalPartySaveData, error) {
	if err := ctx.Err(); err != nil {
		return tsskeygen.LocalPartySaveData{}, ErrContextCanceled
	}
	k.mu.RLock()
	save, ok := k.saveDataCache[cacheID]
	k.mu.RUnlock()
	if !ok {
		return tsskeygen.LocalPartySaveData{}, ErrShardNotFound
	}
	normalizeLocalSaveData(&save)
	return save, nil
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
		params := tss.NewParameters(tss.S256(), peerCtx, pid, partyCount, threshold)
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

func (k *keyGenerator) storeSaveData(cacheID string, save tsskeygen.LocalPartySaveData) {
	k.mu.Lock()
	k.saveDataCache[cacheID] = save
	k.mu.Unlock()
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

func normalizeLocalSaveData(save *tsskeygen.LocalPartySaveData) {
	for _, point := range save.BigXj {
		if point != nil {
			point.SetCurve(tss.S256())
		}
	}
	if save.ECDSAPub != nil {
		save.ECDSAPub.SetCurve(tss.S256())
	}
}

func saveDataToPublicKey(save tsskeygen.LocalPartySaveData) *ecdsa.PublicKey {
	if save.ECDSAPub == nil {
		return nil
	}
	return &ecdsa.PublicKey{
		Curve: gethcrypto.S256(),
		X:     save.ECDSAPub.X(),
		Y:     save.ECDSAPub.Y(),
	}
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
