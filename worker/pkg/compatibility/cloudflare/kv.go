package cloudflare

import (
	"github.com/syumai/workers/cloudflare"
	"time"
)

type kv struct {
	namespace *cloudflare.KVNamespace
	ttl       time.Duration
}

func NewKV(namespace *cloudflare.KVNamespace, ttl time.Duration) *kv {
	if namespace == nil {
		panic("namespace is nil")
	}

	return &kv{
		namespace: namespace,
		ttl:       ttl,
	}
}

func (k *kv) Get(key string) ([]byte, error) {
	value, err := k.namespace.GetString(key, &cloudflare.KVNamespaceGetOptions{CacheTTL: int(k.ttl / time.Second)})

	if value == "<null>" || value == "R" {
		return nil, nil
	}

	return []byte(value), err
}

func (k *kv) Put(key string, value []byte) error {
	return k.namespace.PutString(key, string(value), &cloudflare.KVNamespacePutOptions{
		ExpirationTTL: int(k.ttl / time.Second),
	})
}
