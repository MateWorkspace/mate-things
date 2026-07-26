package infrastructurecacheshared

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultNamespace = "default"

type BaseRedis struct {
	client    redis.UniversalClient
	namespace string
	entity    string
	ttl       TtlConfig
}

func NewRedis(client redis.UniversalClient, namespace string, entity string, ttl TtlConfig) BaseRedis {
	if namespace == "" {
		namespace = defaultNamespace
	}

	defaultTtl := DefaultTtlConfig()
	if ttl.Identity <= 0 {
		ttl.Identity = defaultTtl.Identity
	}
	if ttl.Pagination <= 0 {
		ttl.Pagination = defaultTtl.Pagination
	}
	if ttl.Relation <= 0 {
		ttl.Relation = defaultTtl.Relation
	}

	return BaseRedis{
		client:    client,
		namespace: namespace,
		entity:    entity,
		ttl:       ttl,
	}
}

func (r BaseRedis) IdentityTtl() time.Duration {
	return r.ttl.Identity
}

func (r BaseRedis) PaginationTtl() time.Duration {
	return r.ttl.Pagination
}

func (r BaseRedis) RelationTtl() time.Duration {
	return r.ttl.Relation
}

func (r BaseRedis) Get(ctx context.Context, key string, dest any) (bool, error) {
	value, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	if err := json.Unmarshal(value, dest); err != nil {
		return false, err
	}

	return true, nil
}

func (r BaseRedis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if value == nil {
		return nil
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, raw, ttl).Err()
}

func (r BaseRedis) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	return r.client.Del(ctx, keys...).Err()
}

func (r BaseRedis) Invalidate(ctx context.Context, scopes ...string) error {
	for _, scope := range scopes {
		if err := r.client.Incr(ctx, r.versionKey(scope)).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (r BaseRedis) ScopedKey(ctx context.Context, scope string, parts ...string) (string, error) {
	version, err := r.version(ctx, scope)
	if err != nil {
		return "", err
	}

	keyParts := []string{"cache", r.namespace, r.entity, scope, "v" + version}
	keyParts = append(keyParts, parts...)
	return strings.Join(keyParts, ":"), nil
}

func (r BaseRedis) version(ctx context.Context, scope string) (string, error) {
	value, err := r.client.Get(ctx, r.versionKey(scope)).Result()
	if err != nil {
		if err == redis.Nil {
			return "0", nil
		}
		return "", err
	}

	return value, nil
}

func (r BaseRedis) versionKey(scope string) string {
	return strings.Join([]string{"cache", r.namespace, r.entity, scope, "version"}, ":")
}
