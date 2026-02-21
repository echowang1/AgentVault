package tss

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	tsscommon "github.com/bnb-chain/tss-lib/v2/common"
	tsskeygen "github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	tsssigning "github.com/bnb-chain/tss-lib/v2/ecdsa/signing"
	"github.com/bnb-chain/tss-lib/v2/tss"
	gethcommon "github.com/ethereum/go-ethereum/common"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

type signer struct {
	storage  ShardStorage
	resolver ShareResolver
}

func NewSigner(shardStorage ShardStorage) (Signer, error) {
	if shardStorage == nil {
		return nil, ErrShardNotFound
	}
	resolver, _ := shardStorage.(ShareResolver)
	return &signer{storage: shardStorage, resolver: resolver}, nil
}

func (s *signer) Sign(ctx context.Context, req *SignRequest) (*Signature, error) {
	return s.SignWithProgress(ctx, req, nil)
}

func (s *signer) SignWithProgress(ctx context.Context, req *SignRequest, callback SignProgressCallback) (*Signature, error) {
	if err := ctx.Err(); err != nil {
		return nil, ErrContextCanceled
	}
	if req == nil {
		return nil, ErrSignFailed
	}

	emitSign(callback, "validating request", 10)
	hashBytes, err := parseMessageHash(req.MessageHash)
	if err != nil {
		return nil, err
	}
	if !gethcommon.IsHexAddress(req.Address) {
		return nil, ErrShardMismatch
	}

	shard1Raw, err := base64.StdEncoding.DecodeString(req.Shard1)
	if err != nil {
		return nil, ErrShardMismatch
	}
	shard1, err := unmarshalShare(shard1Raw)
	if err != nil || shard1 == nil {
		return nil, ErrShardMismatch
	}

	emitSign(callback, "loading server shard", 30)
	shard2Raw, err := s.storage.LoadShard2(ctx, req.Shard2ID)
	if err != nil {
		if err == ErrShardNotFound {
			return nil, ErrShardNotFound
		}
		return nil, ErrSignFailed
	}
	shard2, err := unmarshalShare(shard2Raw)
	if err != nil || shard2 == nil {
		return nil, ErrShardMismatch
	}

	emitSign(callback, "checking shard ownership", 45)
	if s.resolver == nil {
		return nil, ErrSignFailed
	}
	save1, err := s.resolver.LoadSaveData(ctx, shard1.CacheID)
	if err != nil {
		return nil, ErrShardMismatch
	}
	save2, err := s.resolver.LoadSaveData(ctx, shard2.CacheID)
	if err != nil {
		return nil, ErrShardMismatch
	}

	emitSign(callback, "running gg18 signing", 70)
	sigData, err := runGG18Signing(ctx, hashBytes, []tsskeygen.LocalPartySaveData{save1, save2})
	if err != nil {
		if err == ErrContextCanceled {
			return nil, err
		}
		return nil, ErrSignFailed
	}

	emitSign(callback, "finalizing signature", 95)
	rBytes := leftPad32(sigData.R)
	sBytes := leftPad32(sigData.S)
	v, err := resolveRecoveryV(hashBytes, rBytes, sBytes, req.Address, sigData.SignatureRecovery)
	if err != nil {
		return nil, ErrInvalidSignature
	}

	rHex := hex.EncodeToString(rBytes)
	sHex := hex.EncodeToString(sBytes)
	full := "0x" + rHex + sHex + fmt.Sprintf("%02x", v)

	emitSign(callback, "done", 100)
	return &Signature{
		R:             rHex,
		S:             sHex,
		V:             v,
		FullSignature: full,
	}, nil
}

func (s *signer) SignBatch(ctx context.Context, reqs []*SignRequest) ([]*Signature, error) {
	results := make([]*Signature, 0, len(reqs))
	for _, req := range reqs {
		sig, err := s.Sign(ctx, req)
		if err != nil {
			return nil, err
		}
		results = append(results, sig)
	}
	return results, nil
}

func emitSign(cb SignProgressCallback, step string, percent int) {
	if cb != nil {
		cb(SignProgress{Step: step, Percent: percent})
	}
}

func parseMessageHash(raw string) ([]byte, error) {
	raw = strings.TrimPrefix(raw, "0x")
	decoded, err := hex.DecodeString(raw)
	if err != nil || len(decoded) != 32 {
		return nil, ErrInvalidHash
	}
	return decoded, nil
}

func runGG18Signing(ctx context.Context, messageHash []byte, saves []tsskeygen.LocalPartySaveData) (*tsscommon.SignatureData, error) {
	if len(saves) != partyCount {
		return nil, ErrInvalidPartyCount
	}

	partyIDs := make(tss.UnSortedPartyIDs, 0, partyCount)
	for i, save := range saves {
		normalizeLocalSaveData(&save)
		if save.ShareID == nil {
			return nil, ErrShardMismatch
		}
		pid := tss.NewPartyID(fmt.Sprintf("%d", i+1), fmt.Sprintf("%d", i+1), save.ShareID)
		partyIDs = append(partyIDs, pid)
		saves[i] = save
	}

	sorted := tss.SortPartyIDs(partyIDs)
	peerCtx := tss.NewPeerContext(sorted)
	msgInt := new(big.Int).SetBytes(messageHash)

	outCh := make(chan tss.Message, 128)
	endCh := make(chan tsscommon.SignatureData, partyCount)
	errCh := make(chan error, partyCount)

	parties := make([]tss.Party, 0, partyCount)
	partyByID := make(map[string]tss.Party, partyCount)

	for _, pid := range sorted {
		var (
			key   tsskeygen.LocalPartySaveData
			found bool
		)
		for _, save := range saves {
			if save.ShareID != nil && save.ShareID.Cmp(pid.KeyInt()) == 0 {
				key = save
				found = true
				break
			}
		}
		if !found {
			return nil, ErrShardMismatch
		}
		params := tss.NewParameters(tss.S256(), peerCtx, pid, partyCount, threshold)
		party := tsssigning.NewLocalParty(msgInt, params, key, outCh, endCh)
		parties = append(parties, party)
		partyByID[pid.Id] = party
	}

	for _, party := range parties {
		go func(p tss.Party) {
			if err := p.Start(); err != nil {
				errCh <- err
			}
		}(party)
	}

	var finalSig *tsscommon.SignatureData
	ended := 0
	for ended < partyCount {
		select {
		case <-ctx.Done():
			return nil, ErrContextCanceled
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
					return nil, ErrSignFailed
				}
				if _, err := target.UpdateFromBytes(wireMsg, msg.GetFrom(), msg.IsBroadcast()); err != nil {
					return nil, err
				}
			}
		case sig := <-endCh:
			tmp := sig
			finalSig = &tmp
			ended++
		}
	}

	if finalSig == nil {
		return nil, ErrSignFailed
	}
	return finalSig, nil
}

func equalAddressConstantTime(a, b string) bool {
	aa := strings.ToLower(strings.TrimSpace(a))
	bb := strings.ToLower(strings.TrimSpace(b))
	if len(aa) != len(bb) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(aa), []byte(bb)) == 1
}

func leftPad32(in []byte) []byte {
	out := make([]byte, 32)
	if len(in) >= 32 {
		copy(out, in[len(in)-32:])
		return out
	}
	copy(out[32-len(in):], in)
	return out
}

func resolveRecoveryV(hash, r, s []byte, targetAddr string, recoveryHint []byte) (uint8, error) {
	candidates := make([]byte, 0, 2)

	if len(recoveryHint) > 0 {
		v := recoveryHint[0]
		if v == 27 || v == 28 {
			candidates = append(candidates, v-27)
		}
		if v == 0 || v == 1 {
			candidates = append(candidates, v)
		}
	}
	candidates = append(candidates, 0, 1)

	seen := map[byte]bool{}
	for _, rec := range candidates {
		if rec > 1 || seen[rec] {
			continue
		}
		seen[rec] = true

		sig := make([]byte, 65)
		copy(sig[:32], r)
		copy(sig[32:64], s)
		sig[64] = rec
		pub, err := gethcrypto.SigToPub(hash, sig)
		if err != nil {
			continue
		}
		addr := gethcrypto.PubkeyToAddress(*pub).Hex()
		if equalAddressConstantTime(addr, targetAddr) {
			return 27 + rec, nil
		}
	}

	return 0, ErrInvalidSignature
}
