package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrRefreshTokenNotFound means the jti is not (or no longer) live: it was
// already used (rotation), revoked (logout / password reset / deactivation),
// expired, or never issued by this server.
var ErrRefreshTokenNotFound = errors.New("refresh token not found")

// RefreshTokenStore keeps the server-side state that makes refresh tokens
// single-use and revocable.
//
//	refresh:token:{jti}     string  -> user_id               TTL = token lifetime
//	refresh:user:{user_id}  set     -> live jtis of the user  TTL reset on every add
//
// The per-user set only exists so RevokeAll can find a user's tokens; a stale
// jti left in it (token already expired) is harmless.
type RefreshTokenStore struct {
	Redis *redis.Client
}

func NewRefreshTokenStore(rdb *redis.Client) *RefreshTokenStore {
	return &RefreshTokenStore{Redis: rdb}
}

func refreshTokenKey(jti string) string   { return "refresh:token:" + jti }
func refreshUserKey(userId string) string { return "refresh:user:" + userId }

func (s *RefreshTokenStore) ready() error {
	if s == nil || s.Redis == nil {
		return errors.New("refresh token store not configured")
	}
	return nil
}

// Save registers a freshly minted refresh token. Both keys are written in one
// MULTI/EXEC so they cannot drift apart.
func (s *RefreshTokenStore) Save(ctx context.Context, userId, jti string, ttl time.Duration) error {
	if err := s.ready(); err != nil {
		return err
	}
	_, err := s.Redis.TxPipelined(ctx, func(p redis.Pipeliner) error {
		p.Set(ctx, refreshTokenKey(jti), userId, ttl)
		p.SAdd(ctx, refreshUserKey(userId), jti)
		p.Expire(ctx, refreshUserKey(userId), ttl)
		return nil
	})
	if err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

// Consume atomically deletes the jti and returns the user id it belonged to.
// GETDEL (not GET then DEL) guarantees that two concurrent requests presenting
// the same refresh token cannot both succeed.
// Returns ErrRefreshTokenNotFound if the jti is not live; any other error is an
// infrastructure failure the caller must surface as 500, not 401.
func (s *RefreshTokenStore) Consume(ctx context.Context, jti string) (string, error) {
	if err := s.ready(); err != nil {
		return "", err
	}
	userId, err := s.Redis.GetDel(ctx, refreshTokenKey(jti)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrRefreshTokenNotFound
	}
	if err != nil {
		return "", fmt.Errorf("consume refresh token: %w", err)
	}
	// Best effort: the token key is already gone, which is what matters.
	s.Redis.SRem(ctx, refreshUserKey(userId), jti)
	return userId, nil
}

// Revoke deletes one refresh token (logout). Revoking an unknown jti is a no-op.
func (s *RefreshTokenStore) Revoke(ctx context.Context, userId, jti string) error {
	if err := s.ready(); err != nil {
		return err
	}
	_, err := s.Redis.TxPipelined(ctx, func(p redis.Pipeliner) error {
		p.Del(ctx, refreshTokenKey(jti))
		p.SRem(ctx, refreshUserKey(userId), jti)
		return nil
	})
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

// RevokeAll deletes every live refresh token of a user (password reset,
// account deactivation, refresh by an inactive/unverified user).
func (s *RefreshTokenStore) RevokeAll(ctx context.Context, userId string) error {
	if err := s.ready(); err != nil {
		return err
	}
	jtis, err := s.Redis.SMembers(ctx, refreshUserKey(userId)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("list refresh tokens: %w", err)
	}
	keys := make([]string, 0, len(jtis)+1)
	for _, jti := range jtis {
		keys = append(keys, refreshTokenKey(jti))
	}
	keys = append(keys, refreshUserKey(userId))
	if err := s.Redis.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}
	return nil
}
