package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type Config struct {
	Key string
}

type Hasher struct {
	key []byte
}

func NewHasher(config Config) *Hasher {
	return &Hasher{
		key: []byte(config.Key),
	}
}

func (hasher *Hasher) MakeHash(value string) string {
	hash := hmac.New(sha256.New, hasher.key)
	_, _ = hash.Write([]byte(value))

	return hex.EncodeToString(hash.Sum(nil))
}
